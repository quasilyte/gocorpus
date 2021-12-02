namespace pako {
    export declare function ungzip(r);
}

declare function untar(tar);

namespace App {
    declare class Go {
    }

    interface loadRepoResult {
        repo: repositoryInfo;
        archive: Uint8Array;
    }

    interface gogrepResult {
        error?: string;
        matches?: string[];
    }

    interface corpusInfo {
        Repositories: repositoryInfo[];
    }

    interface repositoryInfo {
        Name: string;
        Tags: string[];
        Git: string;
        Commit: string;
        Size: number;
        MinifiedSize: number;
        SLOC: number;
        Files: repositoryFileInfo[];
    }

    interface repositoryFileInfo {
        Name: string;
        Flags: number;   
    }

    class RepoData {
        files: RepoFileData[] = [];
    }

    class RepoFileData {
        name: string;
        contents: string;
    }

    declare function gogrep(pattern: string, filename: string, src: string): gogrepResult;

    const appState = {
        metadata: <corpusInfo>(null),
        corpus: new Map<string, RepoData>(),

        go: null,
        wasm: null,

        busy: false,

        // Current query state.
        searchResults: new Map<string, number>(),
        filesTotal: 0,
        filesScanned: 0,
        hits: 0,
    };

    function updateStatus(status: string) {
        document.getElementById('status').innerText = status;
    }

    function loadMetadata() {
        updateStatus('loading metadata');
        console.log("loadMetadata()");
        return fetch('corpus-output/corpus.json')
            .then(response => response.json());
    }

    function loadWASM(metadata) {
        updateStatus('loading wasm');
        console.log("loadWASM()");
        appState.metadata = metadata;
        appState.go = new Go();
        return new Promise((resolve, reject) => {
            WebAssembly.instantiateStreaming(fetch("main.wasm"), appState.go.importObject)
                .then(result => resolve(result.instance))
                .catch(error => reject(error));
        });
    }

    function doChunks(count, f, finished) {
        var i = 0;
        (function step() {
            f(i);
            i++;
            if (i < count) {
                setTimeout(step, 0);
            } else {
                finished();
            }
        })();
    }

    function runQueryRecursive(pattern: string, toScan: repositoryInfo[]) {
        if (toScan.length == 0) {
            appState.busy = false;
            updateStatus('ready');
            let $progress = document.getElementById('search-progress');
            $progress.innerHTML = `Progress: 100% (hits: ${appState.hits})`;
            var $results = document.getElementById('search-results');
            var parts = [];
            var sortedMatches = [...appState.searchResults.entries()].sort((a, b) => b[1] - a[1]);
            for (let e of sortedMatches) {
                let [m, num] = e;
                let numStr = num == 1 ? '' : ` (${num} matches)`;
                parts.push(`<li><span class="result">${m}${numStr}</span></li>`);
            }
            $results.innerHTML += '<ol>' + parts.join('') + '</ol>';
            return;
        }

        let repo = toScan.pop();
        let repoData = appState.corpus.get(repo.Name);
        let files = repoData.files; 
        doChunks(files.length,
            i => {
                let f = files[i];
                updateStatus(`processing ${f.name}`);
                let $progress = document.getElementById('search-progress');
                let progressValue = Math.round((appState.filesScanned / appState.filesTotal) * 100);
                $progress.innerHTML = `Progress: ${progressValue}% (hits: ${appState.hits})`;
                let result = gogrep(pattern, f.name, f.contents);
                if (result.error) {
                    console.error(`grepping ${f.name}: ${result.error}`);
                    return;
                }
                appState.filesScanned++;
                if (!result.matches) {
                    return;
                }
                appState.hits += result.matches.length;
                for (let m of result.matches) {
                    if (appState.searchResults.size >= 1000) {
                        break;
                    }
                    if (appState.searchResults.has(m)) {
                        appState.searchResults.set(m, appState.searchResults.get(m) + 1);
                    } else {
                        appState.searchResults.set(m, 1);
                    }
                }
            },
            () => {
                runQueryRecursive(pattern, toScan);
            });
    }

    function loadRepo(repo: repositoryInfo) {
        updateStatus(`loading ${repo.Name} repository...`);
        return fetch(`corpus-output/${repo.Name}.tar.gz`).
            then(result => new Promise<loadRepoResult>((resolve, reject) => {
                result.arrayBuffer().
                    then(b => resolve({repo: repo, archive: new Uint8Array(b)}))
            }));
    }

    function loadRecursive(toLoad: repositoryInfo[]) {
        if (toLoad.length == 0) {
            appState.busy = false;
            updateStatus('ready');
            return;
        }
        let repo = toLoad.pop();
        loadRepo(repo)
            .then(result => pako.ungzip(result.archive))
            .then(arr => arr.buffer)
            .then(buf => untar(buf))
            .then(files => {
                var repoData = new RepoData();
                var dec = new TextDecoder("utf-8");
                for (let rawFile of files) {
                    let f = new RepoFileData();
                    f.name = rawFile.name;
                    f.contents = dec.decode(rawFile.buffer);
                    repoData.files.push(f);
                }
                console.log(`loaded ${repo.Name} repo`);
                appState.corpus.set(repo.Name, repoData);

                let $checkbox = <HTMLInputElement>(document.getElementById(`repository-${repo.Name}`));
                $checkbox.parentElement.classList.add('blue-text');

                loadRecursive(toLoad);
            });
    }

    function getSelectedRepos(): repositoryInfo[] {
        var selected = [];
        for (let i = 0; i < appState.metadata.Repositories.length; i++) {
            let repo = appState.metadata.Repositories[i];
            let $checkbox = <HTMLInputElement>(document.getElementById(`repository-${repo.Name}`));
            if ($checkbox.checked) {
                selected.push(repo);
            }
        }
        return selected;
    }

    function loadRepositories() {
        if (appState.busy) {
            return;
        }

        appState.busy = true;
        let toLoad: repositoryInfo[] = [];
        for (let repo of getSelectedRepos()) {
            if (!appState.corpus.has(repo.Name)) {
                toLoad.push(repo);
            }
        }
        loadRecursive(toLoad);
    }

    function setupUI() {
        var $corpusSelection = document.getElementById('corpus-selection');
        var selectionHTML = '';
        var buildSelectionTable = () => {
            for (let i = 0; i < appState.metadata.Repositories.length; i++) {
                selectionHTML += '<tr>';
                for (let j = 0; j < 6; j++) {
                    var index = (i * 6) + j;
                    if (index >= appState.metadata.Repositories.length) {
                        return;
                    }
                    var repo = appState.metadata.Repositories[index];
                    let sizeMB = (repo.MinifiedSize * 0.000001).toFixed(1);
                    let hint = `Tags: [${repo.Tags}], Size: ${sizeMB} MB, Files: ${repo.Files.length}, SLOC: ${repo.SLOC}`;
                    selectionHTML += `<td><label title="${hint}"><input id="repository-${repo.Name}" type="checkbox"> ${repo.Name}</label></td>`;
                }
                selectionHTML += '</tr>';
            }
        };
        buildSelectionTable();
        $corpusSelection.innerHTML = '<table style="width: 100%; display: table">' + selectionHTML + '</table>';

        let $loadRepos = document.getElementById('load-button');
        $loadRepos.onclick = function() {
            loadRepositories();
        };

        var $run = document.getElementById('run-button');
        var $results = document.getElementById('search-results');
        $results.innerHTML = '';
        $run.onclick = function() {
            if (appState.busy) {
                return;
            }
            appState.busy = true;
            let pattern = (<HTMLTextAreaElement>document.getElementById('search-pattern')).value;
            $results.innerHTML = '';
            appState.searchResults.clear();
            appState.filesScanned = 0;
            appState.filesTotal = 0;
            appState.hits = 0;
            let repos = getSelectedRepos();
            for (let repo of repos) {
                appState.filesTotal += repo.Files.length;
            }
            let $progress = document.getElementById('search-progress');
            $progress.innerHTML = '';
            runQueryRecursive(pattern, repos);
        };
    }

    function runApp(wasm) {
        updateStatus('ready');
        console.log("runApp()");
        appState.wasm = wasm;
        appState.go.run(wasm);
        setupUI();
    }

    function initPolyfills() {
        if (!WebAssembly.instantiateStreaming) { 
            WebAssembly.instantiateStreaming = async (resp, importObject) => {
                const r = await (await resp).arrayBuffer();
                return await WebAssembly.instantiate(r, importObject);
            }
        }
    }

    export function main() {
        initPolyfills();
        loadMetadata().
            then(metadata => loadWASM(metadata)).
            then(wasm => runApp(wasm));
    }
}

window.onload = function() { 
    App.main();
};
