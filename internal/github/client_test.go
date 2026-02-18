package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := NewClient(server.URL, "")
	return server, client
}

func TestListRepos_ReturnsRepos(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"name":"repo1","fork":false,"archived":false},{"name":"repo2","fork":false,"archived":false}]`)
	})
	defer server.Close()

	repos, err := client.ListRepos("testuser", ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCount := 2
	if len(repos) != expectedCount {
		t.Fatalf("expected %d repos, got %d", expectedCount, len(repos))
	}

	expectedName := "repo1"
	if repos[0].Name != expectedName {
		t.Errorf("expected name %q, got %q", expectedName, repos[0].Name)
	}
}

func TestListRepos_HandlesPagination(t *testing.T) {
	requestCount := 0
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")

		if requestCount == 1 {
			w.Header().Set("Link", fmt.Sprintf(`<http://%s/users/testuser/repos?per_page=100&page=2>; rel="next"`, r.Host))
			fmt.Fprint(w, `[{"name":"repo1"}]`)
		} else {
			fmt.Fprint(w, `[{"name":"repo2"}]`)
		}
	})
	defer server.Close()

	repos, err := client.ListRepos("testuser", ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCount := 2
	if len(repos) != expectedCount {
		t.Fatalf("expected %d repos, got %d", expectedCount, len(repos))
	}

	expectedRequests := 2
	if requestCount != expectedRequests {
		t.Errorf("expected %d requests, got %d", expectedRequests, requestCount)
	}
}

func TestListRepos_UserNotFound(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	_, err := client.ListRepos("nonexistent", ListOptions{})
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestListRepos_RateLimited(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	defer server.Close()

	_, err := client.ListRepos("testuser", ListOptions{})
	if err == nil {
		t.Fatal("expected error for rate limit, got nil")
	}
}

func TestListRepos_SendsAuthToken(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[]`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "ghp_testtoken123")
	client.ListRepos("testuser", ListOptions{})

	expectedAuth := "Bearer ghp_testtoken123"
	if receivedAuth != expectedAuth {
		t.Errorf("expected auth %q, got %q", expectedAuth, receivedAuth)
	}
}

func TestListRepos_FiltersForks(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"name":"original","fork":false},{"name":"forked","fork":true}]`)
	})
	defer server.Close()

	repos, err := client.ListRepos("testuser", ListOptions{NoForks: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCount := 1
	if len(repos) != expectedCount {
		t.Fatalf("expected %d repo, got %d", expectedCount, len(repos))
	}

	expectedName := "original"
	if repos[0].Name != expectedName {
		t.Errorf("expected %q, got %q", expectedName, repos[0].Name)
	}
}

func TestListRepos_FiltersArchived(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"name":"active","archived":false},{"name":"old","archived":true}]`)
	})
	defer server.Close()

	repos, err := client.ListRepos("testuser", ListOptions{NoArchived: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCount := 1
	if len(repos) != expectedCount {
		t.Fatalf("expected %d repo, got %d", expectedCount, len(repos))
	}
}

func TestListRepos_FiltersbyNamePattern(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"name":"api-gateway"},{"name":"api-users"},{"name":"web-frontend"}]`)
	})
	defer server.Close()

	repos, err := client.ListRepos("testuser", ListOptions{Filter: "api-*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCount := 2
	if len(repos) != expectedCount {
		t.Fatalf("expected %d repos, got %d", expectedCount, len(repos))
	}
}

func TestParseNextPage_WithNextLink(t *testing.T) {
	header := `<https://api.github.com/users/test/repos?per_page=100&page=2>; rel="next", <https://api.github.com/users/test/repos?per_page=100&page=5>; rel="last"`

	result := parseNextPage(header)

	expectedPage := 2
	if result != expectedPage {
		t.Errorf("expected page %d, got %d", expectedPage, result)
	}
}

func TestParseNextPage_EmptyHeader(t *testing.T) {
	result := parseNextPage("")

	expectedPage := 0
	if result != expectedPage {
		t.Errorf("expected page %d, got %d", expectedPage, result)
	}
}

func TestParseNextPage_NoNextRel(t *testing.T) {
	header := `<https://api.github.com/users/test/repos?per_page=100&page=1>; rel="prev"`

	result := parseNextPage(header)

	expectedPage := 0
	if result != expectedPage {
		t.Errorf("expected page %d, got %d", expectedPage, result)
	}
}

func TestCloneURL_SSH(t *testing.T) {
	repo := Repo{Name: "myrepo"}
	result := CloneURL(repo, "ssh", "testuser")

	expected := "git@github.com:testuser/myrepo.git"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCloneURL_HTTPS(t *testing.T) {
	repo := Repo{Name: "myrepo", CloneURL: "https://github.com/testuser/myrepo.git"}
	result := CloneURL(repo, "https", "testuser")

	expected := "https://github.com/testuser/myrepo.git"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		glob     string
		input    string
		expected bool
	}{
		{"api-*", "api-gateway", true},
		{"api-*", "web-frontend", false},
		{"*-service", "auth-service", true},
		{"*-service", "auth-controller", false},
		{"exact", "exact", true},
		{"exact", "notexact", false},
	}

	for _, tc := range tests {
		regexStr := "^" + globToRegex(tc.glob) + "$"
		re, err := regexp.Compile(regexStr)
		if err != nil {
			t.Fatalf("invalid regex %q from glob %q: %v", regexStr, tc.glob, err)
		}
		result := re.MatchString(tc.input)
		if result != tc.expected {
			t.Errorf("glob %q against %q: expected %v, got %v", tc.glob, tc.input, tc.expected, result)
		}
	}
}
