package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing test config: %v", err)
	}
	return path
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTestConfig(t, `
accounts:
  - username: testuser
    dir: /tmp/repos
    protocol: ssh
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedAccountCount := 1
	if len(cfg.Accounts) != expectedAccountCount {
		t.Fatalf("expected %d account, got %d", expectedAccountCount, len(cfg.Accounts))
	}

	expectedUsername := "testuser"
	if cfg.Accounts[0].Username != expectedUsername {
		t.Errorf("expected username %q, got %q", expectedUsername, cfg.Accounts[0].Username)
	}

	expectedDir := "/tmp/repos"
	if cfg.Accounts[0].Dir != expectedDir {
		t.Errorf("expected dir %q, got %q", expectedDir, cfg.Accounts[0].Dir)
	}
}

func TestLoad_DefaultsProtocolToSSH(t *testing.T) {
	path := writeTestConfig(t, `
accounts:
  - username: testuser
    dir: /tmp/repos
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedProtocol := "ssh"
	if cfg.Accounts[0].Protocol != expectedProtocol {
		t.Errorf("expected protocol %q, got %q", expectedProtocol, cfg.Accounts[0].Protocol)
	}
}

func TestLoad_MultipleAccounts(t *testing.T) {
	path := writeTestConfig(t, `
accounts:
  - username: user1
    dir: /tmp/repos1
    protocol: ssh
  - username: user2
    dir: /tmp/repos2
    protocol: https
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedAccountCount := 2
	if len(cfg.Accounts) != expectedAccountCount {
		t.Fatalf("expected %d accounts, got %d", expectedAccountCount, len(cfg.Accounts))
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeTestConfig(t, `{{{not yaml`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestValidate_EmptyAccounts(t *testing.T) {
	cfg := &Config{Accounts: []Account{}}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty accounts, got nil")
	}
}

func TestValidate_MissingUsername(t *testing.T) {
	cfg := &Config{
		Accounts: []Account{{Dir: "/tmp", Protocol: "ssh"}},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing username, got nil")
	}
}

func TestValidate_MissingDir(t *testing.T) {
	cfg := &Config{
		Accounts: []Account{{Username: "user", Protocol: "ssh"}},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing dir, got nil")
	}
}

func TestValidate_InvalidProtocol(t *testing.T) {
	cfg := &Config{
		Accounts: []Account{{Username: "user", Dir: "/tmp", Protocol: "svn"}},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for invalid protocol, got nil")
	}
}

func TestAccount_IsActive_DefaultsToTrue(t *testing.T) {
	acct := Account{Username: "user"}

	expectedActive := true
	if acct.IsActive() != expectedActive {
		t.Errorf("expected IsActive() to be %v, got %v", expectedActive, acct.IsActive())
	}
}

func TestAccount_IsActive_RespectsExplicitFalse(t *testing.T) {
	active := false
	acct := Account{Username: "user", Active: &active}

	expectedActive := false
	if acct.IsActive() != expectedActive {
		t.Errorf("expected IsActive() to be %v, got %v", expectedActive, acct.IsActive())
	}
}

func TestActiveAccounts_FiltersInactive(t *testing.T) {
	active := true
	inactive := false
	cfg := &Config{
		Accounts: []Account{
			{Username: "active1", Dir: "/tmp", Protocol: "ssh", Active: &active},
			{Username: "inactive", Dir: "/tmp", Protocol: "ssh", Active: &inactive},
			{Username: "active2", Dir: "/tmp", Protocol: "ssh"},
		},
	}

	result := cfg.ActiveAccounts()

	expectedCount := 2
	if len(result) != expectedCount {
		t.Fatalf("expected %d active accounts, got %d", expectedCount, len(result))
	}

	expectedFirst := "active1"
	if result[0].Username != expectedFirst {
		t.Errorf("expected first active account %q, got %q", expectedFirst, result[0].Username)
	}

	expectedSecond := "active2"
	if result[1].Username != expectedSecond {
		t.Errorf("expected second active account %q, got %q", expectedSecond, result[1].Username)
	}
}

func TestAddAccount_AddsNewAccount(t *testing.T) {
	cfg := &Config{
		Accounts: []Account{
			{Username: "existing", Dir: "/tmp", Protocol: "ssh"},
		},
	}

	err := cfg.AddAccount(Account{Username: "newuser", Dir: "/tmp/new", Protocol: "https"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCount := 2
	if len(cfg.Accounts) != expectedCount {
		t.Fatalf("expected %d accounts, got %d", expectedCount, len(cfg.Accounts))
	}

	expectedUsername := "newuser"
	if cfg.Accounts[1].Username != expectedUsername {
		t.Errorf("expected username %q, got %q", expectedUsername, cfg.Accounts[1].Username)
	}
}

func TestAddAccount_RejectsDuplicate(t *testing.T) {
	cfg := &Config{
		Accounts: []Account{
			{Username: "existing", Dir: "/tmp", Protocol: "ssh"},
		},
	}

	err := cfg.AddAccount(Account{Username: "existing", Dir: "/tmp/other", Protocol: "ssh"})
	if err == nil {
		t.Fatal("expected error for duplicate account, got nil")
	}
}

func TestAddAccount_RejectsDuplicateCaseInsensitive(t *testing.T) {
	cfg := &Config{
		Accounts: []Account{
			{Username: "BoyCook", Dir: "/tmp", Protocol: "ssh"},
		},
	}

	err := cfg.AddAccount(Account{Username: "boycook", Dir: "/tmp/other", Protocol: "ssh"})
	if err == nil {
		t.Fatal("expected error for case-insensitive duplicate, got nil")
	}
}

func TestRemoveAccount_RemovesExisting(t *testing.T) {
	cfg := &Config{
		Accounts: []Account{
			{Username: "user1", Dir: "/tmp", Protocol: "ssh"},
			{Username: "user2", Dir: "/tmp", Protocol: "ssh"},
		},
	}

	err := cfg.RemoveAccount("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCount := 1
	if len(cfg.Accounts) != expectedCount {
		t.Fatalf("expected %d accounts, got %d", expectedCount, len(cfg.Accounts))
	}

	expectedUsername := "user2"
	if cfg.Accounts[0].Username != expectedUsername {
		t.Errorf("expected remaining account %q, got %q", expectedUsername, cfg.Accounts[0].Username)
	}
}

func TestRemoveAccount_CaseInsensitive(t *testing.T) {
	cfg := &Config{
		Accounts: []Account{
			{Username: "BoyCook", Dir: "/tmp", Protocol: "ssh"},
		},
	}

	err := cfg.RemoveAccount("boycook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCount := 0
	if len(cfg.Accounts) != expectedCount {
		t.Fatalf("expected %d accounts, got %d", expectedCount, len(cfg.Accounts))
	}
}

func TestRemoveAccount_NotFound(t *testing.T) {
	cfg := &Config{
		Accounts: []Account{
			{Username: "user1", Dir: "/tmp", Protocol: "ssh"},
		},
	}

	err := cfg.RemoveAccount("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent account, got nil")
	}
}

func TestExpandPath_TildeExpansion(t *testing.T) {
	home, _ := os.UserHomeDir()
	result := expandPath("~/code/repos")

	expected := filepath.Join(home, "code/repos")
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandPath_HomeVarExpansion(t *testing.T) {
	home, _ := os.UserHomeDir()
	result := expandPath("$HOME/code/repos")

	expected := filepath.Join(home, "code/repos")
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandPath_AbsolutePathUnchanged(t *testing.T) {
	result := expandPath("/usr/local/repos")

	expected := "/usr/local/repos"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.yaml")

	active := true
	original := &Config{
		Accounts: []Account{
			{
				Username: "testuser",
				Dir:      "/tmp/repos",
				Protocol: "ssh",
				Active:   &active,
			},
		},
	}

	if err := Save(original, path); err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}

	expectedUsername := "testuser"
	if loaded.Accounts[0].Username != expectedUsername {
		t.Errorf("expected username %q, got %q", expectedUsername, loaded.Accounts[0].Username)
	}
}
