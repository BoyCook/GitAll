package cmd

import (
	"github.com/boycook/gitall/internal/git"
	"github.com/boycook/gitall/internal/output"
	"github.com/boycook/gitall/internal/runner"
	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch latest changes for all repositories",
	Long: `Fetch the latest changes from remotes for all git repositories in
configured directories. This updates remote tracking branches without
modifying your working tree — a safe way to check for updates.`,
	RunE: runFetch,
}

var (
	fetchUser        string
	fetchDir         string
	fetchConcurrency int
)

func init() {
	rootCmd.AddCommand(fetchCmd)

	fetchCmd.Flags().StringVar(&fetchUser, "user", "", "only fetch repos for this user's directory")
	fetchCmd.Flags().StringVar(&fetchDir, "dir", "", "directory to scan (overrides config)")
	fetchCmd.Flags().IntVarP(&fetchConcurrency, "concurrency", "j", 4, "number of concurrent fetches")
}

func runFetch(cmd *cobra.Command, args []string) error {
	repoPaths, err := resolveRepoPaths(fetchUser, fetchDir)
	if err != nil {
		return err
	}

	output.Infof(quiet, "Fetching %d repos...", len(repoPaths))
	results := fetchRepos(repoPaths)
	output.PrintSummary(results, "Fetch", jsonOut)
	return nil
}

func fetchRepos(repos []string) []git.RepoResult {
	tasks := make([]runner.Task, len(repos))
	for i, repoPath := range repos {
		rp := repoPath
		tasks[i] = runner.Task{
			Name: git.RepoNameFromPath(rp),
			Execute: func() git.RepoResult {
				return git.Fetch(rp)
			},
		}
	}

	return runner.RunWithProgress(tasks, fetchConcurrency, func(completed, total int, result git.RepoResult) {
		output.Progress(completed, total, result, quiet)
	})
}
