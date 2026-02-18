package cmd

import (
	"strings"

	"github.com/boycook/gitall/internal/git"
	"github.com/boycook/gitall/internal/output"
	"github.com/boycook/gitall/internal/runner"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull latest changes for all repositories",
	Long: `Pull the latest changes for all git repositories in configured directories.
Repos with uncommitted changes or unpushed commits are skipped by default.
Use --stash to auto-stash dirty repos, and --rebase to pull with rebase.`,
	RunE: runPull,
}

var (
	pullUser        string
	pullDir         string
	pullConcurrency int
	pullStash       bool
	pullRebase      bool
	pullOwnedOnly   bool
	pullOwner       string
)

func init() {
	rootCmd.AddCommand(pullCmd)

	pullCmd.Flags().StringVar(&pullUser, "user", "", "only pull repos for this user's directory")
	pullCmd.Flags().StringVar(&pullDir, "dir", "", "directory to scan (overrides config)")
	pullCmd.Flags().IntVarP(&pullConcurrency, "concurrency", "j", 4, "number of concurrent pulls")
	pullCmd.Flags().BoolVar(&pullStash, "stash", false, "auto-stash dirty repos before pulling")
	pullCmd.Flags().BoolVar(&pullRebase, "rebase", false, "use git pull --rebase")
	pullCmd.Flags().BoolVar(&pullOwnedOnly, "owned-only", false, "only pull repos owned by the configured user")
	pullCmd.Flags().StringVar(&pullOwner, "owner", "", "only pull repos owned by this GitHub user/org")
}

func runPull(cmd *cobra.Command, args []string) error {
	opts := git.PullOptions{
		Stash:  pullStash,
		Rebase: pullRebase,
	}

	repoPaths, err := resolveRepoPaths(pullUser, pullDir)
	if err != nil {
		return err
	}

	if repoPaths != nil {
		repos := filterOwnedRepos(repoPaths)
		output.Infof(quiet, "Pulling %d repos...", len(repos))
		results := pullRepos(repos, opts)
		output.PrintSummary(results, "Pull", jsonOut)
		return nil
	}

	dirs, err := resolveDirs(pullUser, pullDir)
	if err != nil {
		return err
	}

	var allResults []git.RepoResult

	for _, dir := range dirs {
		repos, err := git.DiscoverRepos(dir)
		if err != nil {
			output.Errorf("Error scanning %s: %s", dir, err)
			continue
		}

		repos = filterOwnedRepos(repos)

		output.Infof(quiet, "Pulling %d repos in %s...", len(repos), dir)

		results := pullRepos(repos, opts)
		allResults = append(allResults, results...)
	}

	output.PrintSummary(allResults, "Pull", jsonOut)
	return nil
}

func pullRepos(repos []string, opts git.PullOptions) []git.RepoResult {
	tasks := make([]runner.Task, len(repos))
	for i, repoPath := range repos {
		rp := repoPath
		tasks[i] = runner.Task{
			Name: git.RepoNameFromPath(rp),
			Execute: func() git.RepoResult {
				return git.Pull(rp, opts)
			},
		}
	}

	return runner.RunWithProgress(tasks, pullConcurrency, func(completed, total int, result git.RepoResult) {
		output.Progress(completed, total, result, quiet)
	})
}

func filterOwnedRepos(repos []string) []string {
	owner := resolveOwnerFilter()
	if owner == "" {
		return repos
	}

	var filtered []string
	for _, repoPath := range repos {
		repoOwner := git.RemoteOwner(repoPath)
		if strings.EqualFold(repoOwner, owner) {
			filtered = append(filtered, repoPath)
		} else {
			output.Infof(quiet, "Skipping %s (owned by %s)", git.RepoNameFromPath(repoPath), repoOwner)
		}
	}
	return filtered
}

func resolveOwnerFilter() string {
	if pullOwner != "" {
		return pullOwner
	}
	if pullOwnedOnly && pullUser != "" {
		return pullUser
	}
	return ""
}
