package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/boycook/gitall/internal/config"
	"github.com/boycook/gitall/internal/git"
	"github.com/boycook/gitall/internal/github"
	"github.com/boycook/gitall/internal/output"
	"github.com/boycook/gitall/internal/runner"
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone all repositories for configured accounts",
	Long: `Clone all GitHub repositories for each configured account (or a single
account specified via flags). Repositories that already exist locally
are skipped.`,
	RunE: runClone,
}

var (
	cloneUser       string
	cloneDir        string
	cloneProtocol   string
	cloneConcurrency int
	cloneDryRun     bool
	cloneNoForks    bool
	cloneNoArchived bool
	cloneFilter     string
)

func init() {
	rootCmd.AddCommand(cloneCmd)

	cloneCmd.Flags().StringVar(&cloneUser, "user", "", "GitHub username or organisation")
	cloneCmd.Flags().StringVar(&cloneDir, "dir", ".", "target directory for cloned repos")
	cloneCmd.Flags().StringVar(&cloneProtocol, "protocol", "ssh", "clone protocol (ssh, https)")
	cloneCmd.Flags().IntVarP(&cloneConcurrency, "concurrency", "j", 4, "number of concurrent clones")
	cloneCmd.Flags().BoolVar(&cloneDryRun, "dry-run", false, "show what would be cloned without cloning")
	cloneCmd.Flags().BoolVar(&cloneNoForks, "no-forks", false, "exclude forked repositories")
	cloneCmd.Flags().BoolVar(&cloneNoArchived, "no-archived", false, "exclude archived repositories")
	cloneCmd.Flags().StringVar(&cloneFilter, "filter", "", "filter repos by name pattern (e.g. \"prefix-*\")")
}

func runClone(cmd *cobra.Command, args []string) error {
	accounts, err := resolveAccounts(cloneUser, cloneDir, cloneProtocol)
	if err != nil {
		return err
	}

	listOpts := github.ListOptions{
		NoForks:    cloneNoForks,
		NoArchived: cloneNoArchived,
		Filter:     cloneFilter,
	}

	var allResults []git.RepoResult

	for _, acct := range accounts {
		results, err := cloneAccount(acct, listOpts)
		if err != nil {
			output.Errorf("Error cloning for %s: %s", acct.Username, err)
			continue
		}
		allResults = append(allResults, results...)
	}

	output.PrintSummary(allResults, "Clone", jsonOut)
	return nil
}

func cloneAccount(acct config.Account, listOpts github.ListOptions) ([]git.RepoResult, error) {
	token := resolveToken(acct)
	client := github.NewClient(acct.APIURL, token)

	output.Infof(quiet, "Fetching repos for %s...", acct.Username)

	repos, err := client.ListRepos(acct.Username, listOpts)
	if err != nil {
		return nil, err
	}

	output.Infof(quiet, "Found %d repos for %s", len(repos), acct.Username)

	if cloneDryRun {
		names := make([]string, len(repos))
		for i, r := range repos {
			names[i] = r.Name
		}
		output.PrintDryRun(names, "clone")
		return nil, nil
	}

	tasks := make([]runner.Task, len(repos))
	for i, repo := range repos {
		cloneURL := github.CloneURL(repo, acct.Protocol, acct.Username)
		targetDir := filepath.Join(acct.Dir, strings.ToLower(repo.Name))

		tasks[i] = runner.Task{
			Name: repo.Name,
			Execute: func() git.RepoResult {
				return git.Clone(cloneURL, targetDir)
			},
		}
	}

	results := runner.RunWithProgress(tasks, cloneConcurrency, func(completed, total int, result git.RepoResult) {
		output.Progress(completed, total, result, quiet)
	})

	return results, nil
}

func resolveAccounts(user, dir, protocol string) ([]config.Account, error) {
	if user != "" {
		return []config.Account{{
			Username: user,
			Dir:      dir,
			Protocol: protocol,
		}}, nil
	}

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return nil, fmt.Errorf("no --user flag and %w\nRun 'gitall config init' to create a config file", err)
	}

	active := cfg.ActiveAccounts()
	if len(active) == 0 {
		return nil, fmt.Errorf("no active accounts in config")
	}

	if !quiet {
		for _, acct := range cfg.Accounts {
			if !acct.IsActive() {
				output.Infof(quiet, "Skipping inactive account: %s", acct.Username)
			}
		}
	}

	return active, nil
}

func resolveToken(acct config.Account) string {
	if acct.Token != "" {
		return acct.Token
	}
	return os.Getenv("GITHUB_TOKEN")
}
