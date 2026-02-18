package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("setup %v failed: %s %v", args, out, err)
		}
	}

	return dir
}

func commitFile(t *testing.T, dir, filename, content string) {
	t.Helper()

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing file: %v", err)
	}

	cmd := exec.Command("git", "add", filename)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %s %v", out, err)
	}

	cmd = exec.Command("git", "commit", "-m", "add "+filename)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %s %v", out, err)
	}
}

func TestStatus_CleanRepo(t *testing.T) {
	dir := initTestRepo(t)
	commitFile(t, dir, "README.md", "hello")

	status := Status(dir)

	expectedClean := true
	if status.Clean != expectedClean {
		t.Errorf("expected clean=%v, got %v", expectedClean, status.Clean)
	}

	expectedBranch := "master"
	if status.Branch != expectedBranch && status.Branch != "main" {
		t.Errorf("expected branch %q or 'main', got %q", expectedBranch, status.Branch)
	}
}

func TestStatus_UntrackedFiles(t *testing.T) {
	dir := initTestRepo(t)
	commitFile(t, dir, "README.md", "hello")

	os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("new"), 0o644)

	status := Status(dir)

	expectedUntracked := 1
	if status.Untracked != expectedUntracked {
		t.Errorf("expected %d untracked, got %d", expectedUntracked, status.Untracked)
	}

	expectedClean := false
	if status.Clean != expectedClean {
		t.Errorf("expected clean=%v, got %v", expectedClean, status.Clean)
	}
}

func TestStatus_StagedChanges(t *testing.T) {
	dir := initTestRepo(t)
	commitFile(t, dir, "README.md", "hello")

	os.WriteFile(filepath.Join(dir, "staged.txt"), []byte("staged"), 0o644)
	cmd := exec.Command("git", "add", "staged.txt")
	cmd.Dir = dir
	cmd.CombinedOutput()

	status := Status(dir)

	expectedStaged := 1
	if status.Staged != expectedStaged {
		t.Errorf("expected %d staged, got %d", expectedStaged, status.Staged)
	}
}

func TestStatus_UnstagedChanges(t *testing.T) {
	dir := initTestRepo(t)
	commitFile(t, dir, "README.md", "hello")

	os.WriteFile(filepath.Join(dir, "README.md"), []byte("modified"), 0o644)

	status := Status(dir)

	expectedUnstaged := 1
	if status.Unstaged != expectedUnstaged {
		t.Errorf("expected %d unstaged, got %d", expectedUnstaged, status.Unstaged)
	}
}

func TestStatus_NotAGitRepo(t *testing.T) {
	dir := t.TempDir()
	status := Status(dir)

	if status.Error == "" {
		t.Error("expected error for non-git directory")
	}
}

func TestStatus_DetectsRemoteURL(t *testing.T) {
	dir := initTestRepo(t)
	commitFile(t, dir, "README.md", "hello")

	cmd := exec.Command("git", "remote", "add", "origin", "git@github.com:test/repo.git")
	cmd.Dir = dir
	cmd.CombinedOutput()

	status := Status(dir)

	expectedRemote := "git@github.com:test/repo.git"
	if status.RemoteURL != expectedRemote {
		t.Errorf("expected remote %q, got %q", expectedRemote, status.RemoteURL)
	}
}

func TestClone_SkipsExistingDir(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "existing-repo")
	os.MkdirAll(existing, 0o755)

	result := Clone("https://github.com/test/repo.git", existing)

	expectedStatus := Skipped
	if result.Status != expectedStatus {
		t.Errorf("expected status %v, got %v", expectedStatus, result.Status)
	}
}

func TestClone_FailsWithBadURL(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "bad-clone")

	result := Clone("not-a-valid-url", target)

	expectedStatus := Failed
	if result.Status != expectedStatus {
		t.Errorf("expected status %v, got %v", expectedStatus, result.Status)
	}
}

func TestDiscoverRepos_FindsGitRepos(t *testing.T) {
	dir := t.TempDir()

	repo1 := filepath.Join(dir, "repo1")
	os.MkdirAll(filepath.Join(repo1, ".git"), 0o755)

	repo2 := filepath.Join(dir, "repo2")
	os.MkdirAll(filepath.Join(repo2, ".git"), 0o755)

	notARepo := filepath.Join(dir, "not-a-repo")
	os.MkdirAll(notARepo, 0o755)

	repos, err := DiscoverRepos(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCount := 2
	if len(repos) != expectedCount {
		t.Fatalf("expected %d repos, got %d", expectedCount, len(repos))
	}
}

func TestDiscoverRepos_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	repos, err := DiscoverRepos(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCount := 0
	if len(repos) != expectedCount {
		t.Fatalf("expected %d repos, got %d", expectedCount, len(repos))
	}
}

func TestDiscoverRepos_InvalidDirectory(t *testing.T) {
	_, err := DiscoverRepos("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for invalid directory")
	}
}

func TestRepoNameFromDir(t *testing.T) {
	result := repoNameFromDir("/Users/craig/code/boycook/GitAll")

	expected := "GitAll"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func initTestRepoWithRemote(t *testing.T) (string, string) {
	t.Helper()

	bare := t.TempDir()
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = bare
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init --bare failed: %s %v", out, err)
	}

	clone := t.TempDir()
	cmd = exec.Command("git", "clone", bare, clone)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone failed: %s %v", out, err)
	}

	for _, args := range [][]string{
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	} {
		cmd = exec.Command(args[0], args[1:]...)
		cmd.Dir = clone
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("setup %v failed: %s %v", args, out, err)
		}
	}

	commitFile(t, clone, "README.md", "initial")
	cmd = exec.Command("git", "push", "origin", "HEAD")
	cmd.Dir = clone
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git push failed: %s %v", out, err)
	}

	return clone, bare
}

func TestPull_SkipsDirtyRepo(t *testing.T) {
	clone, _ := initTestRepoWithRemote(t)

	os.WriteFile(filepath.Join(clone, "dirty.txt"), []byte("dirty"), 0o644)

	result := Pull(clone, PullOptions{})

	expectedStatus := Skipped
	if result.Status != expectedStatus {
		t.Errorf("expected status %v, got %v (%s)", expectedStatus, result.Status, result.Message)
	}
}

func TestPull_SkipsUnpushedCommits(t *testing.T) {
	clone, _ := initTestRepoWithRemote(t)

	commitFile(t, clone, "local.txt", "local only")

	result := Pull(clone, PullOptions{})

	expectedStatus := Skipped
	if result.Status != expectedStatus {
		t.Errorf("expected status %v, got %v (%s)", expectedStatus, result.Status, result.Message)
	}
}

func TestPull_AllowsUnpushedWithRebase(t *testing.T) {
	clone, _ := initTestRepoWithRemote(t)

	commitFile(t, clone, "local.txt", "local only")

	result := Pull(clone, PullOptions{Rebase: true})

	if result.Status == Skipped {
		t.Errorf("expected pull to proceed with rebase, got skipped: %s", result.Message)
	}
}

func TestPull_StashesDirtyRepo(t *testing.T) {
	clone, _ := initTestRepoWithRemote(t)

	os.WriteFile(filepath.Join(clone, "README.md"), []byte("modified"), 0o644)

	result := Pull(clone, PullOptions{Stash: true})

	if result.Status == Failed {
		t.Errorf("expected stash+pull to succeed, got failed: %s", result.Message)
	}

	content, _ := os.ReadFile(filepath.Join(clone, "README.md"))
	expectedContent := "modified"
	if string(content) != expectedContent {
		t.Errorf("expected stashed changes restored, got %q", string(content))
	}
}

func TestPull_ReportsUpToDate(t *testing.T) {
	clone, _ := initTestRepoWithRemote(t)

	result := Pull(clone, PullOptions{})

	expectedStatus := UpToDate
	if result.Status != expectedStatus {
		t.Errorf("expected status %v, got %v (%s)", expectedStatus, result.Status, result.Message)
	}
}

func TestPull_SkipsNoUpstream(t *testing.T) {
	dir := initTestRepo(t)
	commitFile(t, dir, "README.md", "hello")

	result := Pull(dir, PullOptions{})

	expectedStatus := Skipped
	if result.Status != expectedStatus {
		t.Errorf("expected status %v, got %v (%s)", expectedStatus, result.Status, result.Message)
	}
}

func TestFetch_SucceedsWithRemote(t *testing.T) {
	clone, _ := initTestRepoWithRemote(t)

	result := Fetch(clone)

	expectedStatus := Success
	if result.Status != expectedStatus {
		t.Errorf("expected status %v, got %v (%s)", expectedStatus, result.Status, result.Message)
	}
}

func TestFetch_SucceedsWithNoRemote(t *testing.T) {
	dir := initTestRepo(t)
	commitFile(t, dir, "README.md", "hello")

	result := Fetch(dir)

	expectedStatus := Success
	if result.Status != expectedStatus {
		t.Errorf("expected status %v, got %v (%s)", expectedStatus, result.Status, result.Message)
	}
}

func TestRemoteOwner_SSH(t *testing.T) {
	dir := initTestRepo(t)
	commitFile(t, dir, "README.md", "hello")

	cmd := exec.Command("git", "remote", "add", "origin", "git@github.com:BoyCook/GitAll.git")
	cmd.Dir = dir
	cmd.CombinedOutput()

	owner := RemoteOwner(dir)

	expectedOwner := "BoyCook"
	if owner != expectedOwner {
		t.Errorf("expected owner %q, got %q", expectedOwner, owner)
	}
}

func TestRemoteOwner_HTTPS(t *testing.T) {
	dir := initTestRepo(t)
	commitFile(t, dir, "README.md", "hello")

	cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/BoyCook/GitAll.git")
	cmd.Dir = dir
	cmd.CombinedOutput()

	owner := RemoteOwner(dir)

	expectedOwner := "BoyCook"
	if owner != expectedOwner {
		t.Errorf("expected owner %q, got %q", expectedOwner, owner)
	}
}

func TestRemoteOwner_NoRemote(t *testing.T) {
	dir := initTestRepo(t)
	commitFile(t, dir, "README.md", "hello")

	owner := RemoteOwner(dir)

	expectedOwner := ""
	if owner != expectedOwner {
		t.Errorf("expected owner %q, got %q", expectedOwner, owner)
	}
}

func TestParsePortcelain_EmptyOutput(t *testing.T) {
	dir := initTestRepo(t)
	commitFile(t, dir, "README.md", "hello")

	staged, unstaged, untracked := parsePortcelain(dir)

	if staged != 0 || unstaged != 0 || untracked != 0 {
		t.Errorf("expected all zeros, got staged=%d unstaged=%d untracked=%d", staged, unstaged, untracked)
	}
}
