package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/boycook/gitall/internal/config"
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

	if len(cfg.Accounts) == 0 {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("removing empty config: %w", err)
		}
		fmt.Printf("Removed account %q — config file deleted (no accounts remaining)\n", username)
		return nil
	}

	if err := config.Save(cfg, path); err != nil {
		return err
	}

	fmt.Printf("Removed account %q from %s\n", username, path)
	return nil
}
