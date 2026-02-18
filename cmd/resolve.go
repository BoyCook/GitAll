package cmd

import (
	"fmt"
	"strings"

	"github.com/boycook/gitall/internal/config"
)

func resolveRepoPaths(user, dir string) ([]string, error) {
	if dir != "" {
		return nil, nil
	}

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return nil, nil
	}

	if !cfg.HasRepos() {
		return nil, nil
	}

	var paths []string
	for _, repo := range cfg.Repos {
		if user != "" && !strings.EqualFold(repo.Owner, user) {
			continue
		}
		paths = append(paths, repo.Dir)
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("no matching repos found")
	}

	return paths, nil
}

func resolveDirs(user, dir string) ([]string, error) {
	if dir != "" {
		return []string{dir}, nil
	}

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return nil, fmt.Errorf("no --dir flag and no config found.\nRun 'gitall config init' to create one, or use --dir to specify a directory")
	}

	active := cfg.ActiveAccounts()
	var dirs []string

	for _, acct := range active {
		if user != "" && !strings.EqualFold(acct.Username, user) {
			continue
		}
		dirs = append(dirs, acct.Dir)
	}

	if len(dirs) == 0 {
		return nil, fmt.Errorf("no matching directories found")
	}

	return dirs, nil
}
