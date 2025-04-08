package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/scm"
	"go.uber.org/zap"
)

// cloneAllRepos clones all student and group repositories from a GitHub course.
//
// Run with the following command:
//
//	cm clone-all-repos
//
// The student/group repositories for the current year are cloned into the directory
// <year>/student-repos/<repo>.
func cloneAllRepos() {
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

	path := filepath.Join(repoHome(), "student-repos")
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		exitErr(err, "Error creating repository directory")
	}
	repoPath := func(repo string) string {
		return filepath.Join(path, repo)
	}

	fmt.Printf("Cloning %d student/group repositories into %q\n", len(ghRepos), path)

	for _, scmRepo := range ghRepos {
		repo := scmRepo.Repo
		// skipping the main course repository (assignments, info, tests)
		if slices.Contains(courseRepos, repo) {
			continue
		}

		msg := fmt.Sprintf("Cloned %q into %q", repo, path)
		if exists(repoPath(repo)) {
			msg = fmt.Sprintf("Repository %q updated", repo)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, err := ghClient.Clone(ctx, &scm.CloneOptions{
			Organization: ghOrg,
			Repository:   repo,
			DestDir:      path,
		})
		cancel()
		if err != nil {
			fmt.Printf("Error cloning %q: %v\n", repo, err)
			continue
		}
		fmt.Println(msg)
	}
}
