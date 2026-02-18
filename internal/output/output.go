package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/boycook/gitall/internal/git"
	"github.com/fatih/color"
)

var (
	bold    = color.New(color.Bold)
	green   = color.New(color.FgGreen)
	yellow  = color.New(color.FgYellow)
	red     = color.New(color.FgRed)
	cyan    = color.New(color.FgCyan)
	dimWhite = color.New(color.FgHiBlack)
)

type Summary struct {
	Action   string
	Total    int
	Success  int
	Skipped  int
	Failed   int
	UpToDate int
}

func Progress(completed, total int, result git.RepoResult, quiet bool) {
	if quiet {
		return
	}

	prefix := fmt.Sprintf("[%d/%d]", completed, total)

	switch result.Status {
	case git.Success:
		fmt.Fprintf(os.Stdout, "%s %s %s\n", dimWhite.Sprint(prefix), green.Sprint(result.Name), result.Message)
	case git.Skipped:
		fmt.Fprintf(os.Stdout, "%s %s %s\n", dimWhite.Sprint(prefix), yellow.Sprint(result.Name), result.Message)
	case git.Failed:
		fmt.Fprintf(os.Stdout, "%s %s %s\n", dimWhite.Sprint(prefix), red.Sprint(result.Name), result.Message)
	case git.UpToDate:
		fmt.Fprintf(os.Stdout, "%s %s %s\n", dimWhite.Sprint(prefix), cyan.Sprint(result.Name), result.Message)
	}
}

func PrintSummary(results []git.RepoResult, action string, asJSON bool) {
	summary := buildSummary(results, action)

	if asJSON {
		printJSONSummary(results, summary)
		return
	}

	fmt.Println()
	bold.Printf("%s summary: ", action)
	parts := []string{}

	if summary.Success > 0 {
		parts = append(parts, green.Sprintf("%d succeeded", summary.Success))
	}
	if summary.Skipped > 0 {
		parts = append(parts, yellow.Sprintf("%d skipped", summary.Skipped))
	}
	if summary.Failed > 0 {
		parts = append(parts, red.Sprintf("%d failed", summary.Failed))
	}
	if summary.UpToDate > 0 {
		parts = append(parts, cyan.Sprintf("%d up-to-date", summary.UpToDate))
	}

	for i, part := range parts {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(part)
	}
	fmt.Printf(" (%d total)\n", summary.Total)
}

func PrintDryRun(items []string, action string) {
	bold.Printf("Dry run — would %s %d repos:\n", action, len(items))
	for _, item := range items {
		fmt.Printf("  %s\n", item)
	}
}

func Infof(quiet bool, format string, args ...any) {
	if !quiet {
		fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

func Errorf(format string, args ...any) {
	red.Fprintf(os.Stderr, format+"\n", args...)
}

func PrintRepoStatus(s git.RepoStatus, verboseMode bool) {
	if s.Error != "" {
		fmt.Fprintf(os.Stdout, "%s %s\n", red.Sprint(s.Name), dimWhite.Sprint(s.Error))
		return
	}

	if s.Clean && !verboseMode {
		return
	}

	var indicators []string

	branchStr := cyan.Sprint(s.Branch)

	if s.Ahead > 0 {
		indicators = append(indicators, green.Sprintf("+%d", s.Ahead))
	}
	if s.Behind > 0 {
		indicators = append(indicators, red.Sprintf("-%d", s.Behind))
	}
	if s.Staged > 0 {
		indicators = append(indicators, yellow.Sprintf("~%d staged", s.Staged))
	}
	if s.Unstaged > 0 {
		indicators = append(indicators, yellow.Sprintf("~%d unstaged", s.Unstaged))
	}
	if s.Untracked > 0 {
		indicators = append(indicators, dimWhite.Sprintf("+%d untracked", s.Untracked))
	}

	nameColor := green
	if !s.Clean {
		nameColor = yellow
	}
	if s.Behind > 0 {
		nameColor = red
	}

	if len(indicators) == 0 {
		fmt.Fprintf(os.Stdout, "%s %s %s\n", nameColor.Sprint(s.Name), branchStr, green.Sprint("clean"))
	} else {
		fmt.Fprintf(os.Stdout, "%s %s %s\n", nameColor.Sprint(s.Name), branchStr, strings.Join(indicators, " "))
	}
}

func PrintRepoList(s git.RepoStatus) {
	branchStr := cyan.Sprint(s.Branch)

	stateStr := green.Sprint("clean")
	if s.Error != "" {
		stateStr = red.Sprint("error")
	} else if !s.Clean {
		stateStr = yellow.Sprint("dirty")
	}

	remoteStr := dimWhite.Sprint(s.RemoteURL)
	if s.RemoteURL == "" {
		remoteStr = dimWhite.Sprint("no remote")
	}

	fmt.Fprintf(os.Stdout, "%s %s %s %s\n", bold.Sprint(s.Name), branchStr, stateStr, remoteStr)
}

type StatusSummary struct {
	Total    int `json:"total"`
	Clean    int `json:"clean"`
	Dirty    int `json:"dirty"`
	Errored  int `json:"errored"`
}

func PrintStatusSummary(statuses []git.RepoStatus, asJSON bool) {
	summary := StatusSummary{Total: len(statuses)}
	for _, s := range statuses {
		if s.Error != "" {
			summary.Errored++
		} else if s.Clean {
			summary.Clean++
		} else {
			summary.Dirty++
		}
	}

	if asJSON {
		out := struct {
			Summary  StatusSummary    `json:"summary"`
			Repos    []git.RepoStatus `json:"repos"`
		}{Summary: summary, Repos: statuses}
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Println()
	bold.Print("Status summary: ")
	parts := []string{}
	if summary.Clean > 0 {
		parts = append(parts, green.Sprintf("%d clean", summary.Clean))
	}
	if summary.Dirty > 0 {
		parts = append(parts, yellow.Sprintf("%d dirty", summary.Dirty))
	}
	if summary.Errored > 0 {
		parts = append(parts, red.Sprintf("%d errored", summary.Errored))
	}
	for i, part := range parts {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(part)
	}
	fmt.Printf(" (%d total)\n", summary.Total)
}

func buildSummary(results []git.RepoResult, action string) Summary {
	s := Summary{Action: action, Total: len(results)}
	for _, r := range results {
		switch r.Status {
		case git.Success:
			s.Success++
		case git.Skipped:
			s.Skipped++
		case git.Failed:
			s.Failed++
		case git.UpToDate:
			s.UpToDate++
		}
	}
	return s
}

type jsonOutput struct {
	Summary Summary          `json:"summary"`
	Repos   []git.RepoResult `json:"repos"`
}

func printJSONSummary(results []git.RepoResult, summary Summary) {
	out := jsonOutput{Summary: summary, Repos: results}
	data, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(data))
}
