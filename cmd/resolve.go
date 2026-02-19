package cmd

import (
	"fmt"
	"strings"

	"github.com/boycook/gitall/internal/config"
	"github.com/boycook/gitall/internal/git"
)

func resolveRepoPaths(user, dir string) ([]string, error) {
	if dir != "" {
		repos, err := git.DiscoverRepos(dir)
		if err != nil {
			return nil, err
		}
		return repos, nil
	}

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return nil, fmt.Errorf("no config found.\nRun 'gitall config init' to create one, or use --dir to specify a directory")
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
