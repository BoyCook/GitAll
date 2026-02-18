package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type ResultStatus int

const (
	Success  ResultStatus = iota
	Skipped
	Failed
	UpToDate
)

func (s ResultStatus) String() string {
	switch s {
	case Success:
		return "success"
	case Skipped:
		return "skipped"
	case Failed:
		return "failed"
	case UpToDate:
		return "up-to-date"
	default:
		return "unknown"
	}
}

type RepoResult struct {
	Name    string       `json:"name"`
	Path    string       `json:"path"`
	Status  ResultStatus `json:"status"`
	Message string       `json:"message"`
}

func Clone(cloneURL, targetDir string) RepoResult {
	name := repoNameFromDir(targetDir)

	if isDir(targetDir) {
		return RepoResult{
			Name:    name,
			Path:    targetDir,
			Status:  Skipped,
			Message: "already exists",
		}
	}

	cmd := exec.Command("git", "clone", cloneURL, targetDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return RepoResult{
			Name:    name,
			Path:    targetDir,
			Status:  Failed,
			Message: strings.TrimSpace(string(output)),
		}
	}

	return RepoResult{
		Name:    name,
		Path:    targetDir,
		Status:  Success,
		Message: "cloned",
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func repoNameFromDir(dir string) string {
	return RepoNameFromPath(dir)
}

func RepoNameFromPath(dir string) string {
	parts := strings.Split(dir, string(os.PathSeparator))
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return dir
}

type RepoStatus struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Branch    string `json:"branch"`
	Upstream  string `json:"upstream,omitempty"`
	Ahead     int    `json:"ahead"`
	Behind    int    `json:"behind"`
	Staged    int    `json:"staged"`
	Unstaged  int    `json:"unstaged"`
	Untracked int    `json:"untracked"`
	Clean     bool   `json:"clean"`
	RemoteURL string `json:"remote_url,omitempty"`
	Error     string `json:"error,omitempty"`
}

func Status(repoPath string) RepoStatus {
	name := repoNameFromDir(repoPath)
	status := RepoStatus{
		Name: name,
		Path: repoPath,
	}

	branch, err := currentBranch(repoPath)
	if err != nil {
		status.Error = fmt.Sprintf("not a git repo or %s", err)
		return status
	}
	status.Branch = branch

	status.RemoteURL = remoteURL(repoPath)
	status.Upstream = upstream(repoPath)

	if status.Upstream != "" {
		status.Ahead, status.Behind = aheadBehind(repoPath, status.Upstream)
	}

	status.Staged, status.Unstaged, status.Untracked = parsePortcelain(repoPath)
	status.Clean = status.Staged == 0 && status.Unstaged == 0 && status.Untracked == 0 && status.Ahead == 0 && status.Behind == 0

	return status
}

func DiscoverRepos(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	var repos []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		gitDir := filepath.Join(dir, entry.Name(), ".git")
		if isDir(gitDir) {
			repos = append(repos, filepath.Join(dir, entry.Name()))
		}
	}
	return repos, nil
}

func currentBranch(dir string) (string, error) {
	out, err := runGit(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return out, nil
}

func remoteURL(dir string) string {
	out, _ := runGit(dir, "config", "--get", "remote.origin.url")
	return out
}

func upstream(dir string) string {
	out, err := runGit(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if err != nil {
		return ""
	}
	return out
}

func aheadBehind(dir, upstreamRef string) (int, int) {
	out, err := runGit(dir, "rev-list", "--left-right", "--count", "HEAD..."+upstreamRef)
	if err != nil {
		return 0, 0
	}

	parts := strings.Fields(out)
	if len(parts) != 2 {
		return 0, 0
	}

	ahead, _ := strconv.Atoi(parts[0])
	behind, _ := strconv.Atoi(parts[1])
	return ahead, behind
}

func parsePortcelain(dir string) (staged, unstaged, untracked int) {
	out, err := runGitRaw(dir, "status", "--porcelain")
	if err != nil || out == "" {
		return 0, 0, 0
	}

	for _, line := range strings.Split(out, "\n") {
		if len(line) < 2 {
			continue
		}
		x := line[0]
		y := line[1]

		if x == '?' {
			untracked++
			continue
		}
		if x != ' ' && x != '?' {
			staged++
		}
		if y != ' ' && y != '?' {
			unstaged++
		}
	}
	return staged, unstaged, untracked
}

type PullOptions struct {
	Stash  bool
	Rebase bool
}

func Pull(repoPath string, opts PullOptions) RepoResult {
	name := repoNameFromDir(repoPath)

	staged, unstaged, untracked := parsePortcelain(repoPath)
	isDirty := staged > 0 || unstaged > 0 || untracked > 0

	if isDirty && !opts.Stash {
		return RepoResult{
			Name:    name,
			Path:    repoPath,
			Status:  Skipped,
			Message: "dirty working tree (use --stash to auto-stash)",
		}
	}

	upstreamRef := upstream(repoPath)
	if upstreamRef == "" {
		return RepoResult{
			Name:    name,
			Path:    repoPath,
			Status:  Skipped,
			Message: "no upstream tracking branch",
		}
	}

	ahead, _ := aheadBehind(repoPath, upstreamRef)
	if ahead > 0 && !opts.Rebase {
		return RepoResult{
			Name:    name,
			Path:    repoPath,
			Status:  Skipped,
			Message: fmt.Sprintf("%d unpushed commits (use --rebase to pull with rebase)", ahead),
		}
	}

	if isDirty && opts.Stash {
		if _, err := runGit(repoPath, "stash", "push", "-m", "gitall-auto-stash"); err != nil {
			return RepoResult{
				Name:    name,
				Path:    repoPath,
				Status:  Failed,
				Message: "failed to stash: " + err.Error(),
			}
		}
	}

	pullArgs := []string{"pull"}
	if opts.Rebase {
		pullArgs = append(pullArgs, "--rebase")
	}

	out, err := runGit(repoPath, pullArgs...)

	if isDirty && opts.Stash {
		runGit(repoPath, "stash", "pop")
	}

	if err != nil {
		return RepoResult{
			Name:    name,
			Path:    repoPath,
			Status:  Failed,
			Message: out,
		}
	}

	if strings.Contains(out, "Already up to date") {
		return RepoResult{
			Name:    name,
			Path:    repoPath,
			Status:  UpToDate,
			Message: "already up to date",
		}
	}

	return RepoResult{
		Name:    name,
		Path:    repoPath,
		Status:  Success,
		Message: "pulled",
	}
}

func Fetch(repoPath string) RepoResult {
	name := repoNameFromDir(repoPath)

	out, err := runGit(repoPath, "fetch", "--all", "--prune")
	if err != nil {
		return RepoResult{
			Name:    name,
			Path:    repoPath,
			Status:  Failed,
			Message: out,
		}
	}

	return RepoResult{
		Name:    name,
		Path:    repoPath,
		Status:  Success,
		Message: "fetched",
	}
}

func HasRemote(repoPath string) bool {
	out, err := runGit(repoPath, "remote")
	return err == nil && out != ""
}

func RemoteOwner(repoPath string) string {
	url := remoteURL(repoPath)
	if url == "" {
		return ""
	}
	if strings.Contains(url, ":") && strings.Contains(url, "git@") {
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 2 {
			pathParts := strings.Split(parts[1], "/")
			if len(pathParts) >= 1 {
				return pathParts[0]
			}
		}
	}
	if strings.Contains(url, "github.com/") {
		parts := strings.SplitAfter(url, "github.com/")
		if len(parts) == 2 {
			pathParts := strings.Split(parts[1], "/")
			if len(pathParts) >= 1 {
				return pathParts[0]
			}
		}
	}
	return ""
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output)), fmt.Errorf("git %s: %s", args[0], strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func runGitRaw(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("git %s: %s", args[0], strings.TrimSpace(string(output)))
	}
	return strings.TrimRight(string(output), "\n"), nil
}
