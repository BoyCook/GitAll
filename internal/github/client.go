package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const defaultAPIURL = "https://api.github.com"

type Repo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	CloneURL string `json:"clone_url"`
	SSHURL   string `json:"ssh_url"`
	Fork     bool   `json:"fork"`
	Archived bool   `json:"archived"`
}

type ListOptions struct {
	NoForks    bool
	NoArchived bool
	Filter     string // glob-style pattern (e.g. "prefix-*")
}

type Client struct {
	apiURL     string
	token      string
	httpClient *http.Client
}

func NewClient(apiURL, token string) *Client {
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	return &Client{
		apiURL:     strings.TrimRight(apiURL, "/"),
		token:      token,
		httpClient: &http.Client{},
	}
}

func (c *Client) ListRepos(username string, opts ListOptions) ([]Repo, error) {
	var allRepos []Repo
	page := 1

	for {
		url := fmt.Sprintf("%s/users/%s/repos?per_page=100&page=%d", c.apiURL, username, page)
		repos, nextPage, err := c.fetchPage(url)
		if err != nil {
			return nil, err
		}

		allRepos = append(allRepos, repos...)

		if nextPage == 0 {
			break
		}
		page = nextPage
	}

	return filterRepos(allRepos, opts), nil
}

func (c *Client) fetchPage(url string) ([]Repo, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "gitall-cli")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("fetching repos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, 0, fmt.Errorf("user or organisation not found")
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		return nil, 0, fmt.Errorf("GitHub API rate limit exceeded — use a token to increase limits")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("reading response: %w", err)
	}

	var repos []Repo
	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, 0, fmt.Errorf("parsing response: %w", err)
	}

	nextPage := parseNextPage(resp.Header.Get("Link"))
	return repos, nextPage, nil
}

func parseNextPage(linkHeader string) int {
	if linkHeader == "" {
		return 0
	}

	for _, part := range strings.Split(linkHeader, ",") {
		part = strings.TrimSpace(part)
		if !strings.Contains(part, `rel="next"`) {
			continue
		}

		start := strings.Index(part, "<")
		end := strings.Index(part, ">")
		if start == -1 || end == -1 {
			continue
		}

		rawURL := part[start+1 : end]
		parsed, err := url.Parse(rawURL)
		if err != nil {
			continue
		}
		if pageStr := parsed.Query().Get("page"); pageStr != "" {
			if page, err := strconv.Atoi(pageStr); err == nil {
				return page
			}
		}
	}

	return 0
}

func filterRepos(repos []Repo, opts ListOptions) []Repo {
	var filtered []Repo

	var pattern *regexp.Regexp
	if opts.Filter != "" {
		regexStr := "^" + globToRegex(opts.Filter) + "$"
		pattern, _ = regexp.Compile(regexStr)
	}

	for _, repo := range repos {
		if opts.NoForks && repo.Fork {
			continue
		}
		if opts.NoArchived && repo.Archived {
			continue
		}
		if pattern != nil && !pattern.MatchString(repo.Name) {
			continue
		}
		filtered = append(filtered, repo)
	}

	return filtered
}

func globToRegex(glob string) string {
	var result strings.Builder
	for _, ch := range glob {
		switch ch {
		case '*':
			result.WriteString(".*")
		case '?':
			result.WriteString(".")
		case '.', '(', ')', '+', '|', '^', '$', '[', ']', '{', '}', '\\':
			result.WriteRune('\\')
			result.WriteRune(ch)
		default:
			result.WriteRune(ch)
		}
	}
	return result.String()
}

func CloneURL(repo Repo, protocol, username string) string {
	switch protocol {
	case "https":
		return repo.CloneURL
	default:
		return "git@github.com:" + path.Join(username, repo.Name) + ".git"
	}
}
