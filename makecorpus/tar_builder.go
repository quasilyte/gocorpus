package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
)

type tarBuilder struct {
	gz  *gzip.Writer
	tar *tar.Writer
}

func newTarBuilder(w io.Writer, compress bool) *tarBuilder {
	b := &tarBuilder{}
	if compress {
		b.gz = gzip.NewWriter(w)
		w = b.gz
	}
	b.tar = tar.NewWriter(w)
	return b
}

func (b *tarBuilder) Flush() error {
	if b.gz != nil {
		if err := b.gz.Flush(); err != nil {
			return err
		}
		if err := b.gz.Close(); err != nil {
			return err
		}
	}
	return b.tar.Close()
}

func (b *tarBuilder) AddFile(filename string, mode int64, data []byte) error {
	header := &tar.Header{
		Name: filename,
		Size: int64(len(data)),
		Mode: mode,
	}
	if err := b.tar.WriteHeader(header); err != nil {
		return err
	}
	if n, err := b.tar.Write(data); err != nil || n != len(data) {
		if err != nil {
			return err
		}
		return errors.New("partial data write")
	}
	if err := b.tar.Flush(); err != nil {
		return err
	}
	return nil
}
