package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/boycook/gitall/internal/config"
	"github.com/boycook/gitall/internal/git"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gitall configuration",
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "Display current configuration",
	RunE:  runConfigList,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default configuration file",
	RunE:  runConfigInit,
}

var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new account to the configuration",
	RunE:  runConfigAdd,
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove [username]",
	Short: "Remove an account from the configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigRemove,
}

var configDiscoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Auto-generate config by scanning a directory for existing repos",
	Long: `Recursively scan a directory for git repositories, group them by
GitHub owner (from remote URL), and generate config entries automatically.
Use --dry-run to preview without writing.`,
	RunE: runConfigDiscover,
}

var (
	discoverDir    string
	discoverDryRun bool
)

var (
	addUsername string
	addDir     string
	addProtocol string
	addToken   string
	addAPIURL  string
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configRemoveCmd)
	configCmd.AddCommand(configDiscoverCmd)

	configDiscoverCmd.Flags().StringVar(&discoverDir, "dir", "", "directory to scan recursively (required)")
	configDiscoverCmd.Flags().BoolVar(&discoverDryRun, "dry-run", false, "preview discovered accounts without writing config")
	configDiscoverCmd.MarkFlagRequired("dir")

	configAddCmd.Flags().StringVar(&addUsername, "username", "", "GitHub username or organisation (required)")
	configAddCmd.Flags().StringVar(&addDir, "dir", "", "target directory for repos (required)")
	configAddCmd.Flags().StringVar(&addProtocol, "protocol", "ssh", "clone protocol (ssh, https)")
	configAddCmd.Flags().StringVar(&addToken, "token", "", "GitHub personal access token")
	configAddCmd.Flags().StringVar(&addAPIURL, "api-url", "", "GitHub Enterprise API URL")
	configAddCmd.MarkFlagRequired("username")
	configAddCmd.MarkFlagRequired("dir")
}

func runConfigList(cmd *cobra.Command, args []string) error {
	path := config.DefaultPath()
	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("no config found at %s — run 'gitall config init' to create one", path)
	}

	bold := color.New(color.Bold)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	bold.Fprintf(os.Stdout, "Config: %s\n\n", path)

	if len(cfg.Accounts) > 0 {
		bold.Fprintln(os.Stdout, "Accounts:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "USERNAME\tDIR\tPROTOCOL\tACTIVE\tAPI URL")
		fmt.Fprintln(w, "--------\t---\t--------\t------\t-------")

		for _, acct := range cfg.Accounts {
			activeStr := green.Sprint("yes")
			if !acct.IsActive() {
				activeStr = red.Sprint("no")
			}

			apiURL := "github.com"
			if acct.APIURL != "" {
				apiURL = acct.APIURL
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				acct.Username, acct.Dir, acct.Protocol, activeStr, apiURL)
		}

		w.Flush()
	}

	if len(cfg.Repos) > 0 {
		if len(cfg.Accounts) > 0 {
			fmt.Fprintln(os.Stdout)
		}
		bold.Fprintln(os.Stdout, "Repos:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tOWNER\tDIR\tPROTOCOL")
		fmt.Fprintln(w, "----\t-----\t---\t--------")

		for _, repo := range cfg.Repos {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				repo.Name, repo.Owner, repo.Dir, repo.Protocol)
		}

		w.Flush()
	}

	return nil
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	path := config.DefaultPath()

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config already exists at %s — edit it directly or delete it first", path)
	}

	cfg := config.DefaultConfig()
	if err := config.Save(cfg, path); err != nil {
		return err
	}

	fmt.Printf("Created default config at %s\n", path)
	fmt.Println("Edit it to add your GitHub accounts.")
	return nil
}

func runConfigAdd(cmd *cobra.Command, args []string) error {
	path := config.DefaultPath()

	cfg, err := config.Load(path)
	if err != nil {
		cfg = &config.Config{}
	}

	acct := config.Account{
		Username: addUsername,
		Dir:      addDir,
		Protocol: addProtocol,
		Token:    addToken,
		APIURL:   addAPIURL,
	}

	if err := cfg.AddAccount(acct); err != nil {
		return err
	}

	if err := config.Save(cfg, path); err != nil {
		return err
	}

	fmt.Printf("Added account %q to %s\n", addUsername, path)
	return nil
}

func runConfigDiscover(cmd *cobra.Command, args []string) error {
	discoveredPaths, err := git.DiscoverReposRecursive(discoverDir)
	if err != nil {
		return err
	}

	if len(discoveredPaths) == 0 {
		fmt.Println("No git repositories found.")
		return nil
	}

	type ownerInfo struct {
		protocol string
		dirs     map[string]bool
	}

	owners := make(map[string]*ownerInfo)
	var repos []config.Repo

	for _, repoPath := range discoveredPaths {
		owner := git.RemoteOwner(repoPath)
		if owner == "" {
			continue
		}

		protocol := git.RemoteProtocol(repoPath)
		name := git.RepoNameFromPath(repoPath)
		lowerOwner := strings.ToLower(owner)

		repos = append(repos, config.Repo{
			Name:     name,
			Owner:    lowerOwner,
			Dir:      repoPath,
			Protocol: protocol,
		})

		parentDir := filepath.Dir(repoPath)

		if _, ok := owners[lowerOwner]; !ok {
			owners[lowerOwner] = &ownerInfo{
				protocol: protocol,
				dirs:     map[string]bool{parentDir: true},
			}
		} else {
			owners[lowerOwner].dirs[parentDir] = true
		}
	}

	if len(repos) == 0 {
		fmt.Println("No repos with GitHub remotes found.")
		return nil
	}

	var accounts []config.Account
	for owner, info := range owners {
		for dir := range info.dirs {
			accounts = append(accounts, config.Account{
				Username: owner,
				Dir:      dir,
				Protocol: info.protocol,
			})
		}
	}

	bold := color.New(color.Bold)
	green := color.New(color.FgGreen)

	bold.Printf("Discovered %d repo(s):\n\n", len(repos))
	for _, repo := range repos {
		green.Printf("  %s", repo.Name)
		fmt.Printf("  %s/%s  (%s)\n", repo.Owner, repo.Name, repo.Protocol)
	}

	if discoverDryRun {
		fmt.Println("\nDry run — no config written.")
		return nil
	}

	path := config.DefaultPath()
	cfg, err := config.Load(path)
	if err != nil {
		cfg = &config.Config{}
	}

	addedAccounts := 0
	for _, acct := range accounts {
		if err := cfg.AddAccount(acct); err == nil {
			addedAccounts++
		}
	}

	addedRepos := 0
	for _, repo := range repos {
		if err := cfg.AddRepo(repo); err == nil {
			addedRepos++
		}
	}

	if addedAccounts == 0 && addedRepos == 0 {
		fmt.Println("\nAll discovered entries already exist in config.")
		return nil
	}

	if err := config.Save(cfg, path); err != nil {
		return err
	}

	fmt.Printf("\nAdded %d account(s) and %d repo(s) to %s\n", addedAccounts, addedRepos, path)
	return nil
}

func runConfigRemove(cmd *cobra.Command, args []string) error {
	path := config.DefaultPath()
	username := args[0]

	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("no config found at %s", path)
	}

	if err := cfg.RemoveAccount(username); err != nil {
		return err
	}

	if len(cfg.Accounts) == 0 && len(cfg.Repos) == 0 {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("removing empty config: %w", err)
		}
		fmt.Printf("Removed account %q — config file deleted (no entries remaining)\n", username)
		return nil
	}

	if err := config.Save(cfg, path); err != nil {
		return err
	}

	fmt.Printf("Removed account %q from %s\n", username, path)
	return nil
}
