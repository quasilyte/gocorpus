package main

import (
	"errors"
	"fmt"
	"strings"
)

var knownRepoTags = map[string]struct{}{
	"db":         {},
	"grpc":       {},
	"compiler":   {},
	"parser":     {},
	"lib":        {},
	"go-tools":   {},
	"tool":       {},
	"encoder":    {},
	"decoder":    {},
	"logging":    {},
	"net":        {},
	"crypto":     {},
	"math":       {},
	"cli":        {},
	"framework":  {},
	"testing":    {},
	"sql":        {},
	"orm":        {},
	"kubernetes": {},
	"metrics":    {},
	"os":         {},
	"ci":         {},
	"nfs":        {},
	"ebpf":       {},
}

func validateRepo(repo *repository) error {
	if repo.name == "" {
		return errors.New("empty repo name")
	}
	if repo.git == "" {
		return errors.New("empty repo git")
	}
	if !strings.HasSuffix(repo.git, ".git") {
		return errors.New("git link doesn't end with '.git'")
	}
	if len(repo.srcRoots) == 0 {
		return errors.New("empty repo src roots list")
	}
	if len(repo.tags) == 0 {
		return errors.New("empty repo tags list")
	}
	for _, tag := range repo.tags {
		if _, ok := knownRepoTags[tag]; !ok {
			return fmt.Errorf("unknown %s tag", tag)
		}
	}
	return nil
}
