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
        err?: string;
        skipped?: boolean;
        matches?: string[];
    }

    interface corpusInfo {
        Version: number;
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
        SLOC: number;
    }

    class RepoData {
        files: RepoFileData[] = [];
    }

    class RepoFileData {
        name: string;
        contents: string;
    }

    interface gogrepArgs {
        pattern: string;
        filter: string;
        fileFlags: number;
        targetName: string;
        targetSrc: string;
    }

    declare function gogrep(args: gogrepArgs): gogrepResult;

    const appState = {
        metadata: <corpusInfo>(null),
        corpus: new Map<string, RepoData>(),

        go: null,
        wasm: null,

        busy: false,
        running: false,

        // Current query state.
        runError: '',
        runStartTime: 0,
        searchResults: new Map<string, number>(),
        filesTotal: 0,
        filesScanned: 0,
        slocProcessed: 0,
        hits: 0,
    };

    function updateStatus(status: string) {
        document.getElementById('status').innerText = status;
    }

    function ready() {
        if (appState.runError === '') {
            updateStatus('ready');
        } else {
            updateStatus('ERROR ' + appState.runError);
        }
        appState.runError = '';
    }

    function loadMetadata() {
        updateStatus('loading metadata');
        console.log("loadMetadata()");
        return fetch('corpus-output/corpus.json')
            .then(response => response.json());
    }

    function loadWASM(metadata) {
        appState.metadata = metadata;
        let $totalSLOC = document.getElementById('total-sloc');
        let totalSLOC = 0;
        for (let repo of appState.metadata.Repositories) {
            totalSLOC += repo.SLOC;
        }
        $totalSLOC.innerText = totalSLOC.toLocaleString();

        updateStatus('loading wasm');
        console.log("loadWASM()");
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
            if (!f(i)) {
                finished(true);
                return;
            }
            i++;
            if (i < count) {
                setTimeout(step, 0);
            } else {
                finished(false);
            }
        })();
    }

    function calculateFrequencyScore(): number {
        const baselineFrequency = 70.0; // `err != nil` score
        const resultFrequency = appState.slocProcessed / appState.hits;
        return 100.0 * (baselineFrequency / resultFrequency);
    }

    function searchDone() {
        let endTime = window.performance.now();
        let elapsedMillis = endTime - appState.runStartTime;
        let elapsedSeconds = elapsedMillis / 1000.0;
        appState.busy = false;
        ready();
        var $run = document.getElementById('run-button');
        $run.innerText = 'Run';
        appState.running = false;
        let $progress = document.getElementById('search-progress');
        $progress.innerHTML = `Progress: 100% (hits: ${appState.hits})`;
        var $results = document.getElementById('search-results');
        var parts = [];
        var sortedMatches = [...appState.searchResults.entries()].sort((a, b) => b[1] - a[1]);
        let freqScore = calculateFrequencyScore();
        parts.push(`<p><i>Frequency score: ${freqScore.toFixed(4)}</i></p>`);
        parts.push(`<p><i>Time elapsed: ${elapsedSeconds.toFixed(2)} sec</i></p>`);
        for (let e of sortedMatches) {
            let [m, num] = e;
            let numStr = num == 1 ? '' : ` (${num} matches)`;
            parts.push(`<li><span class="result">${m}${numStr}</span></li>`);
        }
        $results.innerHTML += '<ol>' + parts.join('') + '</ol>';
    }

    function runQueryRecursive(pattern: string, filter: string, toScan: repositoryInfo[]) {
        if (toScan.length == 0) {
            searchDone();
            return;
        }

        let repo = toScan.pop();
        let repoData = appState.corpus.get(repo.Name);
        let files = repoData.files; 
        doChunks(files.length,
            i => {
                if (!appState.running) {
                    return false;
                }
                let f = files[i];
                let filename = f.name;
                if (filename.length > 58) {
                    let nameParts = filename.split('/');
                    let baseName = nameParts[nameParts.length-1];
                    let prefixLen = baseName.length >= 28 ? 32 : 48;
                    filename = filename.substr(0, prefixLen) + '{...}/' + baseName;
                }
                updateStatus(`processing ${filename}`);
                let $progress = document.getElementById('search-progress');
                let progressValue = Math.round((appState.filesScanned / appState.filesTotal) * 100);
                $progress.innerHTML = `Progress: ${progressValue}% (hits: ${appState.hits})`;

                let fileInfo = repo.Files[i];
                let result = gogrep({
                    pattern: pattern,
                    filter: filter,
                    fileFlags: fileInfo.Flags,
                    targetName: f.name,
                    targetSrc: f.contents,
                });
                if (result.err) {
                    console.error(`grepping ${f.name}: ${result.err}`);
                    let canContinue = result.err.includes('parse Go:');
                    if (canContinue) {
                        return true;
                    }
                    appState.runError = result.err;
                    appState.running = false;
                    return false;
                }
                if (!result.skipped) {
                    appState.slocProcessed += fileInfo.SLOC;
                }
                appState.filesScanned++;
                if (!result.matches) {
                    return true;
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
                return true;
            },
            (stopped) => {
                if (stopped) {
                    searchDone();
                } else {
                    runQueryRecursive(pattern, filter, toScan);
                }
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

    function loadRecursive(toLoad: repositoryInfo[], onFinish) {
        if (toLoad.length == 0) {
            appState.busy = false;
            ready();
            if (onFinish) {
                onFinish();
            }
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

                loadRecursive(toLoad, onFinish);
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

    function forcedLoadRepositories(onFinish = function() {}) {
        appState.busy = true;
        let toLoad: repositoryInfo[] = [];
        for (let repo of getSelectedRepos()) {
            if (!appState.corpus.has(repo.Name)) {
                toLoad.push(repo);
            }
        }
        loadRecursive(toLoad, onFinish);
    }

    function loadRepositories() {
        if (appState.busy) {
            return;
        }
        forcedLoadRepositories();
    }

    function allReposSelected() {
        return getSelectedRepos().length === appState.metadata.Repositories.length;   
    }

    function updateCheckallButton(unselect) {
        let $selectAll = document.getElementById('selectall-button');
        if (unselect) {
            $selectAll.innerText = 'Unselect all';
        } else {
            $selectAll.innerText = 'Select all';
        }
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
                    let hint = `Tags: [${repo.Tags}], Size: ${sizeMB} MB, Files: ${repo.Files.length}, SLOC: ${repo.SLOC.toLocaleString()}`;
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

        let $selectAll = document.getElementById('selectall-button');
        $selectAll.onclick = function() {
            let allSelected = allReposSelected();
            let value = !allSelected;
            for (let repo of appState.metadata.Repositories) {
                let $checkbox = <HTMLInputElement>(document.getElementById(`repository-${repo.Name}`));
                $checkbox.checked = value;
            }
            updateCheckallButton(value);
        };

        for (let repo of appState.metadata.Repositories) {
            let $checkbox = <HTMLInputElement>(document.getElementById(`repository-${repo.Name}`));
            $checkbox.onchange = function() {
                updateCheckallButton(allReposSelected());
            };
        }

        var $run = document.getElementById('run-button');
        var $results = document.getElementById('search-results');
        $results.innerHTML = '';
        $run.onclick = function() {
            if (appState.running) {
                appState.running = false;
                return;
            }
            if (appState.busy) {
                return;
            }
            appState.busy = true;
            let pattern = (<HTMLTextAreaElement>document.getElementById('search-pattern')).value;
            let filter = (<HTMLTextAreaElement>document.getElementById('results-filter')).value;
            $results.innerHTML = '';
            appState.searchResults.clear();
            appState.filesScanned = 0;
            appState.slocProcessed = 0;
            appState.filesTotal = 0;
            appState.hits = 0;
            let repos = getSelectedRepos();
            for (let repo of repos) {
                appState.filesTotal += repo.Files.length;
            }
            forcedLoadRepositories(() => {
                let $progress = document.getElementById('search-progress');
                $progress.innerHTML = '';
                $run.innerText = 'Stop';
                appState.running = true;
                appState.runStartTime = window.performance.now();
                runQueryRecursive(pattern, filter, repos);
            });
        };
    }

    function runApp(wasm) {
        ready();
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
