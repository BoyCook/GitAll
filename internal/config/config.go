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
	Repos []Repo `yaml:"repos,omitempty"`
}

type Repo struct {
	Name     string `yaml:"name"`
	Owner    string `yaml:"owner"`
	Dir      string `yaml:"dir"`
	Protocol string `yaml:"protocol"`
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
	if len(c.Repos) == 0 {
		return fmt.Errorf("config must contain at least one repo")
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

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Repos: []Repo{
			{
				Name:     "example-repo",
				Owner:    "your-github-username",
				Dir:      filepath.Join(home, "code", "example-repo"),
				Protocol: "ssh",
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
