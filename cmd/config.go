package cmd

import (
	"fmt"
	"os"
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

var configPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove repos whose directories no longer exist",
	RunE:  runConfigPrune,
}

var configDiscoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Auto-generate config by scanning a directory for existing repos",
	Long: `Recursively scan a directory for git repositories, group them by
GitHub owner (from remote URL), and generate config entries automatically.
Use --dry-run to preview without writing.`,
	RunE: runConfigDiscover,
}

var pruneDryRun bool

var (
	discoverDir    string
	discoverDryRun bool
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configPruneCmd)
	configCmd.AddCommand(configDiscoverCmd)

	configPruneCmd.Flags().BoolVar(&pruneDryRun, "dry-run", false, "preview what would be removed without writing config")

	configDiscoverCmd.Flags().StringVar(&discoverDir, "dir", "", "directory to scan recursively (required)")
	configDiscoverCmd.Flags().BoolVar(&discoverDryRun, "dry-run", false, "preview discovered repos without writing config")
	configDiscoverCmd.MarkFlagRequired("dir")
}

func runConfigList(cmd *cobra.Command, args []string) error {
	path := config.DefaultPath()
	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("no config found at %s — run 'gitall config init' to create one", path)
	}

	bold := color.New(color.Bold)

	bold.Fprintf(os.Stdout, "Config: %s\n\n", path)

	bold.Fprintln(os.Stdout, "Repos:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tOWNER\tDIR\tPROTOCOL")
	fmt.Fprintln(w, "----\t-----\t---\t--------")

	for _, repo := range cfg.Repos {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			repo.Name, repo.Owner, repo.Dir, repo.Protocol)
	}

	w.Flush()

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
	fmt.Println("Edit it to add your repositories.")
	return nil
}

func runConfigPrune(cmd *cobra.Command, args []string) error {
	path := config.DefaultPath()
	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("no config found at %s — run 'gitall config init' to create one", path)
	}

	removed := cfg.PruneRepos()

	if len(removed) == 0 {
		fmt.Println("All repo directories exist — nothing to prune.")
		return nil
	}

	red := color.New(color.FgRed)
	for _, repo := range removed {
		red.Printf("  removed: %s", repo.Name)
		fmt.Printf("  %s\n", repo.Dir)
	}

	if pruneDryRun {
		fmt.Printf("\nDry run — would remove %d repo(s).\n", len(removed))
		return nil
	}

	if err := config.Save(cfg, path); err != nil {
		return err
	}

	fmt.Printf("\nPruned %d repo(s) from %s\n", len(removed), path)
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
	}

	if len(repos) == 0 {
		fmt.Println("No repos with GitHub remotes found.")
		return nil
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

	addedRepos := 0
	for _, repo := range repos {
		if err := cfg.AddRepo(repo); err == nil {
			addedRepos++
		}
	}

	if addedRepos == 0 {
		fmt.Println("\nAll discovered repos already exist in config.")
		return nil
	}

	if err := config.Save(cfg, path); err != nil {
		return err
	}

	fmt.Printf("\nAdded %d repo(s) to %s\n", addedRepos, path)
	return nil
}
