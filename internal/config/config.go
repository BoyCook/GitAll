package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var validProtocols = map[string]bool{
	"ssh":   true,
	"https": true,
}

type Config struct {
	Accounts []Account `yaml:"accounts,omitempty"`
	Repos    []Repo    `yaml:"repos,omitempty"`
}

type Repo struct {
	Name     string `yaml:"name"`
	Owner    string `yaml:"owner"`
	Dir      string `yaml:"dir"`
	Protocol string `yaml:"protocol"`
}

type Account struct {
	Username string `yaml:"username"`
	Dir      string `yaml:"dir"`
	Protocol string `yaml:"protocol"`
	Token    string `yaml:"token,omitempty"`
	APIURL   string `yaml:"api_url,omitempty"`
	Active   *bool  `yaml:"active,omitempty"`
}

func (a Account) IsActive() bool {
	if a.Active == nil {
		return true
	}
	return *a.Active
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gitall", "config.yaml")
}

func Load(path string) (*Config, error) {
	expanded := expandPath(path)
	data, err := os.ReadFile(expanded)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	for i := range cfg.Accounts {
		cfg.Accounts[i].Dir = expandPath(cfg.Accounts[i].Dir)

		if cfg.Accounts[i].Protocol == "" {
			cfg.Accounts[i].Protocol = "ssh"
		}
	}

	for i := range cfg.Repos {
		cfg.Repos[i].Dir = expandPath(cfg.Repos[i].Dir)

		if cfg.Repos[i].Protocol == "" {
			cfg.Repos[i].Protocol = "ssh"
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Accounts) == 0 && len(c.Repos) == 0 {
		return fmt.Errorf("config must contain at least one account or repo")
	}

	for i, acct := range c.Accounts {
		if acct.Username == "" {
			return fmt.Errorf("account %d: username is required", i+1)
		}
		if acct.Dir == "" {
			return fmt.Errorf("account %d (%s): dir is required", i+1, acct.Username)
		}
		if !validProtocols[acct.Protocol] {
			return fmt.Errorf("account %d (%s): invalid protocol %q (must be ssh or https)", i+1, acct.Username, acct.Protocol)
		}
	}

	for i, repo := range c.Repos {
		if repo.Name == "" {
			return fmt.Errorf("repo %d: name is required", i+1)
		}
		if repo.Dir == "" {
			return fmt.Errorf("repo %d (%s): dir is required", i+1, repo.Name)
		}
		if !validProtocols[repo.Protocol] {
			return fmt.Errorf("repo %d (%s): invalid protocol %q (must be ssh or https)", i+1, repo.Name, repo.Protocol)
		}
	}

	return nil
}

func (c *Config) ActiveAccounts() []Account {
	var active []Account
	for _, acct := range c.Accounts {
		if acct.IsActive() {
			active = append(active, acct)
		}
	}
	return active
}

func (c *Config) AddAccount(acct Account) error {
	for _, existing := range c.Accounts {
		if strings.EqualFold(existing.Username, acct.Username) {
			return fmt.Errorf("account %q already exists", acct.Username)
		}
	}
	c.Accounts = append(c.Accounts, acct)
	return nil
}

func (c *Config) HasRepos() bool {
	return len(c.Repos) > 0
}

func (c *Config) RepoDirs() []string {
	dirs := make([]string, len(c.Repos))
	for i, repo := range c.Repos {
		dirs[i] = repo.Dir
	}
	return dirs
}

func (c *Config) AddRepo(repo Repo) error {
	for _, existing := range c.Repos {
		if existing.Dir == repo.Dir {
			return fmt.Errorf("repo at %q already exists", repo.Dir)
		}
	}
	c.Repos = append(c.Repos, repo)
	return nil
}

func (c *Config) RemoveAccount(username string) error {
	for i, acct := range c.Accounts {
		if strings.EqualFold(acct.Username, username) {
			c.Accounts = append(c.Accounts[:i], c.Accounts[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("account %q not found", username)
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	active := true
	return &Config{
		Accounts: []Account{
			{
				Username: "your-github-username",
				Dir:      filepath.Join(home, "code"),
				Protocol: "ssh",
				Active:   &active,
			},
		},
	}
}

func Save(cfg *Config, path string) error {
	expanded := expandPath(path)

	dir := filepath.Dir(expanded)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	if err := os.WriteFile(expanded, data, 0o644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}

	if strings.Contains(path, "$HOME") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return strings.ReplaceAll(path, "$HOME", home)
	}

	return path
}
