package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(0)

	ctx := &context{
		tmpDir: os.TempDir(),
		meta: &CorpusMeta{
			Version: corpusVersion,
		},
	}

	flag.BoolVar(&ctx.verbose, "v", false, "whether to print debug output")
	outputDir := flag.String("o", "corpus-output", "the output directory")
	noCompression := flag.Bool("no-gzip", false, "if provided, raw tars will be produced, without gz compression")
	flag.Parse()

	if err := os.MkdirAll(*outputDir, os.ModePerm); err != nil {
		panic(err)
	}

	ctx.outDir = *outputDir
	ctx.numRepos = len(repositoryList)

	for i, s := range repositoryList {
		ctx.i = i + 1
		ctx.repo = s
		ctx.logDebugf("start loading (%d/%d)", i+1, ctx.numRepos)
		compress := !*noCompression
		suffix := ".tar"
		if compress {
			suffix += ".gz"
		}
		f, err := os.Create(filepath.Join(ctx.outDir, s.name+suffix))
		if err != nil {
			ctx.logErrorf("create output file: %v", err)
			continue
		}
		ctx.tar = newTarBuilder(f, compress)
		meta := collectFiles(ctx)
		if err := ctx.tar.Flush(); err != nil {
			ctx.logErrorf("flush output file: %v", err)
			continue
		}
		if meta != nil {
			ctx.meta.Repositories = append(ctx.meta.Repositories, meta)
		}
	}

	metaFilename := filepath.Join(ctx.outDir, "corpus.json")
	var metaFileData bytes.Buffer
	ctx.meta.WriteJSON(&metaFileData, 0)
	if err := os.WriteFile(metaFilename, metaFileData.Bytes(), 0o666); err != nil {
		panic(err)
	}

	if ctx.numWarnings != 0 {
		log.Printf("warnings: %d", ctx.numWarnings)
	}
	if ctx.numErrors != 0 {
		log.Printf("errors: %d", ctx.numErrors)
	}

	avgDepth := ctx.totalDepth / ctx.numFiles
	log.Printf("max file depth: %d", ctx.maxDepth)
	log.Printf("avg file depth: %d", avgDepth)

	exitCode := 0
	if ctx.numErrors != 0 {
		exitCode = 1
	}
	os.Exit(exitCode)
}

type context struct {
	i        int
	numRepos int

	meta *CorpusMeta
	repo *repository
	tar  *tarBuilder

	tmpDir string

	outDir  string
	verbose bool

	maxDepth   int64
	totalDepth int64
	numFiles   int64

	numErrors   int
	numWarnings int
}

// logDebugf logs a formatted message if verbose mode is enabled.
// The message is automatically annotated with the current code source name tag.
func (ctx *context) logDebugf(format string, args ...interface{}) {
	if ctx.verbose {
		log.Printf("[%s] %s", ctx.repo.name, fmt.Sprintf(format, args...))
	}
}

func (ctx *context) logWarnf(format string, args ...interface{}) {
	if ctx.verbose {
		log.Printf("[%s] WARNING: %s", ctx.repo.name, fmt.Sprintf(format, args...))
	}
	ctx.numWarnings++
}

func (ctx *context) logErrorf(format string, args ...interface{}) {
	log.Printf("[%s] ERROR: %s", ctx.repo.name, fmt.Sprintf(format, args...))
	ctx.numErrors++
}
