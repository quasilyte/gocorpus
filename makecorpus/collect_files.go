package main

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func collectFiles(ctx *context) *RepositoryMeta {
	repo := ctx.repo
	meta := &RepositoryMeta{
		Name: repo.name,
		Tags: repo.tags,
		Git:  repo.git,
	}

	// We're going to clone sources to a tmp dir.
	// This dir will be removed when we're finished.
	cloneTmpDir := filepath.Join(ctx.tmpDir, repo.name+"-tmp")
	defer os.RemoveAll(cloneTmpDir)

	ctx.logDebugf("doing a git clone")
	out, err := exec.Command("git", "clone", "--depth=1", repo.git, cloneTmpDir).CombinedOutput()
	if err != nil {
		ctx.logErrorf("git clone: %v: %s", err, out)
		return nil
	}

	out, err = exec.Command("git", "-C", cloneTmpDir, "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		ctx.logErrorf("git rev-parse: %v: %s", err, out)
		return nil
	}
	meta.Commit = strings.TrimSpace(string(out))

	ctx.logDebugf("processing files")

	numFiles := 0
	for _, srcRoot := range repo.srcRoots {
		absSrcRoot := filepath.Join(cloneTmpDir, srcRoot)
		err := filepath.WalkDir(absSrcRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				skipDir := d.Name() == "node_modules" ||
					strings.HasPrefix(d.Name(), "_") ||
					strings.HasSuffix(path, "/testdata") ||
					strings.HasSuffix(path, "/vendor") ||
					strings.HasSuffix(path, "/third_party")
				if skipDir {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}

			stat, err := d.Info()
			if err != nil {
				return err
			}

			rawSrc, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, path, rawSrc, parser.ParseComments)
			if err != nil {
				return err
			}
			sloc := fset.Position(f.End()).Line
			meta.SLOC += sloc
			minifiedSrc := minifyGo(fset, f)

			relPath := strings.TrimPrefix(path, absSrcRoot)
			prettyPath := filepath.Join(repo.name, srcRoot, relPath)

			fileInfo := analyzeFile(d.Name(), f, rawSrc)
			fileMeta := newFileMeta(fileInfo)
			fileMeta.Name = strings.TrimPrefix(prettyPath, repo.name+"/")
			fileMeta.SLOC = sloc
			meta.Files = append(meta.Files, fileMeta)

			ctx.totalDepth += int64(fileInfo.maxDepth)
			if int64(fileInfo.maxDepth) > ctx.maxDepth {
				ctx.maxDepth = int64(fileInfo.maxDepth)
			}

			meta.Size += len(rawSrc)
			meta.MinifiedSize += len(minifiedSrc)

			if err := ctx.tar.AddFile(prettyPath, int64(stat.Mode()), minifiedSrc); err != nil {
				return err
			}

			numFiles++
			return nil
		})
		if err != nil {
			ctx.logErrorf("%v", err)
			return nil
		}
	}

	ctx.numFiles += int64(numFiles)

	ctx.logDebugf("processed %d files (SLOC=%d)", numFiles, meta.SLOC)
	return meta
}
