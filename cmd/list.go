package cmd

import (
	"sync"

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
	listUser        string
	listDir         string
	listConcurrency int
)

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listUser, "user", "", "only list repos for this user")
	listCmd.Flags().StringVar(&listDir, "dir", "", "directory to scan (overrides config)")
	listCmd.Flags().IntVarP(&listConcurrency, "concurrency", "j", 8, "number of concurrent checks")
}

func runList(cmd *cobra.Command, args []string) error {
	repoPaths, err := resolveRepoPaths(listUser, listDir)
	if err != nil {
		return err
	}

	statuses := listReposConcurrently(repoPaths, listConcurrency)
	for _, s := range statuses {
		output.PrintRepoList(s)
	}
	output.PrintStatusSummary(statuses, jsonOut)
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
