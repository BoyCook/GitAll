package runner

import (
	"sync"

	"github.com/boycook/gitall/internal/git"
)

type Task struct {
	Name    string
	Execute func() git.RepoResult
}

type Result struct {
	Index  int
	Result git.RepoResult
}

func Run(tasks []Task, concurrency int) []git.RepoResult {
	if concurrency < 1 {
		concurrency = 1
	}

	results := make([]git.RepoResult, len(tasks))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t Task) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = t.Execute()
		}(i, task)
	}

	wg.Wait()
	return results
}

func RunWithProgress(tasks []Task, concurrency int, onProgress func(completed, total int, result git.RepoResult)) []git.RepoResult {
	if concurrency < 1 {
		concurrency = 1
	}

	results := make([]git.RepoResult, len(tasks))
	sem := make(chan struct{}, concurrency)
	resultCh := make(chan Result, len(tasks))

	var wg sync.WaitGroup
	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t Task) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			r := t.Execute()
			resultCh <- Result{Index: idx, Result: r}
		}(i, task)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	completed := 0
	total := len(tasks)
	for r := range resultCh {
		completed++
		results[r.Index] = r.Result
		if onProgress != nil {
			onProgress(completed, total, r.Result)
		}
	}

	return results
}
