package cmd

import (
	"sync"

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
	repoPaths, err := resolveRepoPaths(statusUser, statusDir)
	if err != nil {
		return err
	}

	if repoPaths != nil {
		output.Infof(quiet, "Checking %d repos...", len(repoPaths))
		statuses := statusReposConcurrently(repoPaths, statusConcurrency)
		for _, s := range statuses {
			output.PrintRepoStatus(s, statusAll || verbose)
		}
		output.PrintStatusSummary(statuses, jsonOut)
		return nil
	}

	dirs, err := resolveDirs(statusUser, statusDir)
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

