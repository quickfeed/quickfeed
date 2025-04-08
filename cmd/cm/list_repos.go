package main

import (
	"context"
	"flag"
	"fmt"
	"slices"
	"time"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/scm"
	"go.uber.org/zap"
)

// listRepos lists all student and group repositories from a GitHub course.
//
// Run with the following command:
//
//	cm list-repos
//	cm list-repos -url
//
// The -url flag specifies that the URL of the repository should be printed instead of the name.
func listRepos(args []string) {
	fs := flag.NewFlagSet(cloneRepoCmd, flag.ExitOnError)
	var url bool
	fs.BoolVar(&url, "url", false, "Print only the URL of the repository")

	if err := fs.Parse(args); err != nil {
		exitErr(err, "Error parsing flags")
	}

	if err := loadEnv(); err != nil {
		exitErr(err, "Error loading environment variables")
	}
	// use the GITHUB_ACCESS_TOKEN environment variable if the .env-github file is not found
	_ = env.Load(".env-github")
	token, err := env.GetAccessToken()
	if err != nil {
		exitErr(err, "cm: GitHub access token required for this operation")
	}

	ghClient, err := scm.NewSCMClient(zap.NewNop().Sugar(), token)
	if err != nil {
		exitErr(err, "Error creating GitHub client")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Alternative: use the GitHub CLI to list repositories:
	// gh repo list dat520-2025 --limit 100 --json name,url
	// gh repo list dat520-2025 --limit 100 --json name
	ghOrg := courseOrg()
	ghRepos, err := ghClient.GetRepositories(ctx, ghOrg)
	if err != nil {
		exitErr(err, "Error listing repositories")
	}

	fmt.Printf("Listing %d student/group repositories:\n", len(ghRepos))

	for _, scmRepo := range ghRepos {
		repo := scmRepo.Repo
		// skipping the main course repository (assignments, info, tests)
		if slices.Contains(courseRepos, repo) {
			continue
		}
		if url {
			repo = scmRepo.HTMLURL
		}
		fmt.Printf("  %s\n", repo)
	}
}
