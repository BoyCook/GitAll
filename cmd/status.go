package cmd

import (
	"fmt"
	"sync"

	"github.com/boycook/gitall/internal/config"
	"github.com/boycook/gitall/internal/git"
	"github.com/boycook/gitall/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all repositories",
	Long: `Check the git status of all repositories in configured directories.
Shows branch, ahead/behind, staged, unstaged, and untracked counts.
Only dirty repos are shown by default — use --all to include clean repos.`,
	RunE: runStatus,
}

var (
	statusUser        string
	statusDir         string
	statusConcurrency int
	statusAll         bool
)

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringVar(&statusUser, "user", "", "only check repos for this user's directory")
	statusCmd.Flags().StringVar(&statusDir, "dir", "", "directory to scan (overrides config)")
	statusCmd.Flags().IntVarP(&statusConcurrency, "concurrency", "j", 8, "number of concurrent status checks")
	statusCmd.Flags().BoolVar(&statusAll, "all", false, "show all repos including clean ones")
}

func runStatus(cmd *cobra.Command, args []string) error {
	dirs, err := resolveStatusDirs(statusUser, statusDir)
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

		output.Infof(quiet, "Checking %d repos in %s...", len(repos), dir)

		statuses := statusReposConcurrently(repos, statusConcurrency)
		for _, s := range statuses {
			output.PrintRepoStatus(s, statusAll || verbose)
		}

		allStatuses = append(allStatuses, statuses...)
	}

	output.PrintStatusSummary(allStatuses, jsonOut)
	return nil
}

func statusReposConcurrently(repos []string, concurrency int) []git.RepoStatus {
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

func resolveStatusDirs(user, dir string) ([]string, error) {
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
		if user != "" && acct.Username != user {
			continue
		}
		dirs = append(dirs, acct.Dir)
	}

	if len(dirs) == 0 {
		return nil, fmt.Errorf("no matching directories found")
	}

	return dirs, nil
}
