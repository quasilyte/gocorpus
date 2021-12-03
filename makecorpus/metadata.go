package main

import (
	"fmt"
	"io"

	"github.com/quasilyte/gocorpus/internal/filebits"
)

type CorpusMeta struct {
	Repositories []*RepositoryMeta
}

func (m *CorpusMeta) WriteJSON(w io.Writer, indent int) {
	w.Write([]byte("{\n"))
	w.Write([]byte("\t\"Repositories\": [\n"))
	for i, s := range m.Repositories {
		s.WriteJSON(w, indent+1)
		if i != len(m.Repositories)-1 {
			w.Write([]byte(","))
		}
		w.Write([]byte("\n"))
	}
	w.Write([]byte("\t]\n"))
	w.Write([]byte("}\n"))
}

type RepositoryMeta struct {
	Name         string
	Tags         []string
	Git          string
	Commit       string
	Size         int
	MinifiedSize int
	SLOC         int
	Files        []FileMeta
}

func (m *RepositoryMeta) WriteJSON(w io.Writer, indent int) {
	fmt.Fprintf(w, "%s{\n", tabs[indent+1])
	fmt.Fprintf(w, "%s\"Name\": %q,\n", tabs[indent+2], m.Name)
	{
		fmt.Fprintf(w, "%s\"Tags\": [", tabs[indent+2])
		for i, tag := range m.Tags {
			fmt.Fprintf(w, "%q", tag)
			if i != len(m.Tags)-1 {
				w.Write([]byte(", "))
			}
		}
		fmt.Fprintf(w, "],\n")
	}
	fmt.Fprintf(w, "%s\"Git\": %q,\n", tabs[indent+2], m.Git)
	fmt.Fprintf(w, "%s\"Commit\": %q,\n", tabs[indent+2], m.Commit)
	fmt.Fprintf(w, "%s\"Size\": %d,\n", tabs[indent+2], m.Size)
	fmt.Fprintf(w, "%s\"MinifiedSize\": %d,\n", tabs[indent+2], m.MinifiedSize)
	fmt.Fprintf(w, "%s\"SLOC\": %d,\n", tabs[indent+2], m.SLOC)
	fmt.Fprintf(w, "%s\"Files\": [\n", tabs[indent+2])
	for i, f := range m.Files {
		f.WriteJSON(w, indent+3)
		if i != len(m.Files)-1 {
			w.Write([]byte(","))
		}
		w.Write([]byte("\n"))
	}
	fmt.Fprintf(w, "%s]\n", tabs[indent+2])
	fmt.Fprintf(w, "%s}", tabs[indent+1])
}

type FileMeta struct {
	Name  string
	Flags int
}

func (m *FileMeta) WriteJSON(w io.Writer, indent int) {
	fmt.Fprintf(w, `%s{"Name": %q, "Flags": %d}`, tabs[indent], m.Name, m.Flags)
}

func newFileMeta(info *repositoryFileInfo) FileMeta {
	var m FileMeta
	if info.isTest {
		m.Flags |= filebits.IsTest
	}
	if info.isAutogen {
		m.Flags |= filebits.IsAutogen
	}
	if info.isMain {
		m.Flags |= filebits.IsMain
	}
	if info.importsC {
		m.Flags |= filebits.ImportsC
	}
	if info.importsUnsafe {
		m.Flags |= filebits.ImportsUnsafe
	}
	if info.importsReflect {
		m.Flags |= filebits.ImportsReflect
	}
	return m
}

var tabs = [...]string{
	"",
	"\t",
	"\t\t",
	"\t\t\t",
	"\t\t\t\t",
	"\t\t\t\t\t",
	"\t\t\t\t\t\t",
	"\t\t\t\t\t\t\t",
}
