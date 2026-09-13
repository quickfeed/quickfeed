package main

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/kit/score"
)

// newTabWriter returns a tab writer configured the way every table in this
// command is formatted.
func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 2, 8, 2, ' ', 0)
}

// printResults prints the build log of a test run, followed by the score of
// each test and the weighted total.
func printResults(w io.Writer, results *score.Results) {
	if log := results.GetBuildInfo().GetBuildLog(); strings.TrimSpace(log) != "" {
		fmt.Fprintf(w, "%s\n%s\n%s\n", strings.Repeat("-", 70), log, strings.Repeat("-", 70))
	}
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "TEST\tSCORE\tMAX\tWEIGHT")
	for _, sc := range results.Scores {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\n", sc.GetTestName(), sc.GetScore(), sc.GetMaxScore(), sc.GetWeight())
	}
	tw.Flush()
	fmt.Fprintf(w, "Total: %d%%\n", results.Sum())
}

// printIssues prints the problems found in the course repositories as warnings.
// These do not prevent a test run, but they often explain a surprising result.
func printIssues(w io.Writer, issues []assignments.RepoIssue) {
	for _, issue := range issues {
		fmt.Fprintf(w, "warning: %s\n", issue)
	}
}
