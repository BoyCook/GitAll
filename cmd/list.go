package cmd

import (
	"sync"

	"github.com/boycook/gitall/internal/config"
	"github.com/boycook/gitall/internal/git"
	"github.com/boycook/gitall/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all local repositories",
	Long: `List all git repositories found in configured directories.
Shows repo name, branch, clean/dirty state, and remote URL.
This is a local-only operation — no network calls are made.`,
	RunE: runList,
}

var (
	listDir         string
	listConcurrency int
)

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listDir, "dir", "", "directory to scan (overrides config)")
	listCmd.Flags().IntVarP(&listConcurrency, "concurrency", "j", 8, "number of concurrent checks")
}

func runList(cmd *cobra.Command, args []string) error {
	dirs, err := resolveListDirs(listDir)
	if err != nil {
		return err
	}

	var allStatuses []git.RepoStatus

	for _, dir := range dirs {
		repos, err := git.DiscoverRepos(dir)
		if err != nil {
			output.Errorf("Error scanning %s: %s", dir, err)
			continue
		}

		statuses := listReposConcurrently(repos, listConcurrency)
		for _, s := range statuses {
			output.PrintRepoList(s)
		}

		allStatuses = append(allStatuses, statuses...)
	}

	output.PrintStatusSummary(allStatuses, jsonOut)
	return nil
}

func listReposConcurrently(repos []string, concurrency int) []git.RepoStatus {
	if concurrency < 1 {
		concurrency = 1
	}

	results := make([]git.RepoStatus, len(repos))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, repo := range repos {
		wg.Add(1)
		go func(idx int, repoPath string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = git.Status(repoPath)
		}(i, repo)
	}

	wg.Wait()
	return results
}

func resolveListDirs(dir string) ([]string, error) {
	if dir != "" {
		return []string{dir}, nil
	}

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return nil, err
	}

	active := cfg.ActiveAccounts()
	dirs := make([]string, len(active))
	for i, acct := range active {
		dirs[i] = acct.Dir
	}

	return dirs, nil
}
