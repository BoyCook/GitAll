package runner

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/boycook/gitall/internal/git"
)

func TestRun_ExecutesAllTasks(t *testing.T) {
	tasks := []Task{
		{Name: "repo1", Execute: func() git.RepoResult { return git.RepoResult{Name: "repo1", Status: git.Success} }},
		{Name: "repo2", Execute: func() git.RepoResult { return git.RepoResult{Name: "repo2", Status: git.Success} }},
		{Name: "repo3", Execute: func() git.RepoResult { return git.RepoResult{Name: "repo3", Status: git.Success} }},
	}

	results := Run(tasks, 2)

	expectedCount := 3
	if len(results) != expectedCount {
		t.Fatalf("expected %d results, got %d", expectedCount, len(results))
	}

	for i, r := range results {
		if r.Status != git.Success {
			t.Errorf("result %d: expected Success, got %v", i, r.Status)
		}
	}
}

func TestRun_PreservesOrder(t *testing.T) {
	tasks := []Task{
		{Name: "a", Execute: func() git.RepoResult {
			time.Sleep(50 * time.Millisecond)
			return git.RepoResult{Name: "a"}
		}},
		{Name: "b", Execute: func() git.RepoResult { return git.RepoResult{Name: "b"} }},
		{Name: "c", Execute: func() git.RepoResult { return git.RepoResult{Name: "c"} }},
	}

	results := Run(tasks, 4)

	expectedFirst := "a"
	if results[0].Name != expectedFirst {
		t.Errorf("expected first result %q, got %q", expectedFirst, results[0].Name)
	}

	expectedSecond := "b"
	if results[1].Name != expectedSecond {
		t.Errorf("expected second result %q, got %q", expectedSecond, results[1].Name)
	}
}

func TestRun_RespectsConcurrencyLimit(t *testing.T) {
	var maxConcurrent int64
	var currentConcurrent int64

	tasks := make([]Task, 10)
	for i := range tasks {
		name := string(rune('a' + i))
		tasks[i] = Task{
			Name: name,
			Execute: func() git.RepoResult {
				current := atomic.AddInt64(&currentConcurrent, 1)
				for {
					old := atomic.LoadInt64(&maxConcurrent)
					if current <= old || atomic.CompareAndSwapInt64(&maxConcurrent, old, current) {
						break
					}
				}
				time.Sleep(20 * time.Millisecond)
				atomic.AddInt64(&currentConcurrent, -1)
				return git.RepoResult{Name: name, Status: git.Success}
			},
		}
	}

	concurrencyLimit := 2
	Run(tasks, concurrencyLimit)

	observedMax := atomic.LoadInt64(&maxConcurrent)
	if observedMax > int64(concurrencyLimit) {
		t.Errorf("expected max concurrency %d, observed %d", concurrencyLimit, observedMax)
	}
}

func TestRun_HandlesEmptyTasks(t *testing.T) {
	results := Run(nil, 4)

	expectedCount := 0
	if len(results) != expectedCount {
		t.Fatalf("expected %d results, got %d", expectedCount, len(results))
	}
}

func TestRunWithProgress_CallsProgressCallback(t *testing.T) {
	tasks := []Task{
		{Name: "repo1", Execute: func() git.RepoResult { return git.RepoResult{Name: "repo1", Status: git.Success} }},
		{Name: "repo2", Execute: func() git.RepoResult { return git.RepoResult{Name: "repo2", Status: git.Failed} }},
	}

	var progressCalls int
	var lastCompleted int
	var lastTotal int

	results := RunWithProgress(tasks, 1, func(completed, total int, result git.RepoResult) {
		progressCalls++
		lastCompleted = completed
		lastTotal = total
	})

	expectedCalls := 2
	if progressCalls != expectedCalls {
		t.Errorf("expected %d progress calls, got %d", expectedCalls, progressCalls)
	}

	expectedLastCompleted := 2
	if lastCompleted != expectedLastCompleted {
		t.Errorf("expected last completed %d, got %d", expectedLastCompleted, lastCompleted)
	}

	expectedTotal := 2
	if lastTotal != expectedTotal {
		t.Errorf("expected total %d, got %d", expectedTotal, lastTotal)
	}

	expectedResults := 2
	if len(results) != expectedResults {
		t.Fatalf("expected %d results, got %d", expectedResults, len(results))
	}
}
