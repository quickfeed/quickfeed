package main

import (
	"bytes"
	"cmp"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// cloneRepo clones a repository from GitHub and pulls assignments from the main repository.
//
// Run with the following command:
//
//	cm clone-repo -repo <repo> [-pull]
//
// Example:
//
//	cm clone-repo -repo meling-labs
//	cm clone-repo -repo group-repo
//	cm clone-repo -pull
//	cm clone-repo -all
//
// The -repo flag specifies the GitHub repository to clone. If the flag is not set,
// the repository is assumed to be the GitHub username-labs repository.
//
// The -pull flag specifies that assignments should be pulled from the main repository
// after cloning the repository. This is useful when setting up a new repository after
// signing up for the course.
//
// The -all flag clones all student/group repositories for the current year.
// The repositories are cloned into the directory <year>/student-repos/<repo>.
func cloneRepo(args []string) {
	if err := loadEnv(); err != nil {
		exitErr(err, "Error loading environment variables")
	}

	fs := flag.NewFlagSet(cloneRepoCmd, flag.ExitOnError)
	var repo string
	var pull, all bool
	fs.StringVar(&repo, "repo", "", "GitHub repository (default: GitHub username-labs)")
	fs.BoolVar(&pull, "pull", false, "Pull assignments from main repository")
	fs.BoolVar(&all, "all", false, "Clone all student/group repositories")

	if err := fs.Parse(args); err != nil {
		exitErr(err, "Error parsing flags")
	}

	if all {
		cloneAllRepos()
		return
	}

	repo = cmp.Or(repo, githubUsername()+"-labs")
	path := repoPath(repo)
	if exists(path) {
		fmt.Printf("Repository %q already exists, aborting.\n", repo)
		return
	}

	fmt.Printf("Cloning %q into %q\n", repo, repoHome())
	if err := runCommand(repoHome(), "git", "clone", gitURL(courseOrg(), repo)); err != nil {
		exitErr(err, "Error cloning repository")
	}

	if !pull {
		return
	}
	if hasBranches(path) {
		fmt.Printf("Repository %q is not empty, skipping pull.\n", repo)
		return
	}

	fmt.Printf("Pulling main %q repository into %q\n", assignmentsRepo, repo)
	if err := pullAssignments(path, courseOrg(), assignmentsRepo); err != nil {
		exitErr(err, "Error pulling assignments")
	}
}

func repoHome() string {
	return filepath.Join(filepath.Dir(gitRoot), year())
}

func cloneAllRepos() {
	ghOrg := courseOrg()
	// gh repo list dat520-2025 --limit 100 --json name,url --template '{{.name}}'
	// TODO(meling): implement listRepos using the scm package
	ghRepos := []string{"meling-labs", "group-repo"}
	// ghRepos, err := listRepos(ghOrg)
	// if err != nil {
	// 	exitErr(err, "Error listing repositories")
	// }

	path := filepath.Join(repoHome(), "student-repos")
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		exitErr(err, "Error creating repository directory")
	}
	repoPath := func(repo string) string {
		return filepath.Join(path, repo)
	}

	for _, repo := range ghRepos {
		// skipping the main course repository (assignments, info, tests)
		if slices.Contains(courseRepos, repo) {
			continue
		}

		path := repoPath(repo)
		if exists(path) {
			fmt.Printf("Repository %q already exists, skipping.\n", repo)
			continue
		}

		fmt.Printf("Cloning %q into %q\n", repo, path)
		if err := runCommand(path, "git", "clone", gitURL(ghOrg, repo)); err != nil {
			exitErr(err, "Error cloning repository")
		}
	}
}

// pullAssignments sets up a [repo] remote in repoPath and
// pulls from the [repo] repository.
func pullAssignments(repoPath, ghOrg, repo string) error {
	// Example:
	// cd ../2025/meling-labs
	// git remote add assignments git@github.com:dat520-2025/assignments
	// git pull assignments main
	commands := [][]string{
		{"git", "remote", "add", repo, gitURL(ghOrg, repo)},
		{"git", "pull", repo, "main"},
	}
	for _, cmd := range commands {
		if err := runCommand(repoPath, cmd...); err != nil {
			return err
		}
	}
	return nil
}

// hasBranches checks if there are any branches in the given repository.
func hasBranches(repoPath string) bool {
	headsPath := filepath.Join(repoPath, ".git", "refs", "heads")
	if _, err := os.Stat(headsPath); os.IsNotExist(err) {
		fmt.Printf("Warning: not a valid git repository: %s\n", repoPath)
		return false
	}
	files, err := os.ReadDir(headsPath)
	if err != nil {
		fmt.Printf("Error reading heads directory: %v\n", err)
		return false
	}
	// return true if there are any branches (files) in the heads directory
	return len(files) > 0
}

// githubUsername returns the GitHub username of the current user.
// This assumes that the user has set up SSH keys for GitHub.
func githubUsername() string {
	b := bytes.Buffer{}
	// this ssh command writes to stderr and always returns exit code 1.
	c := exec.Command("ssh", "-T", "git@github.com")
	c.Stdout, c.Stderr = os.Stdout, &b
	if err := c.Run(); err != nil {
		// the command fails with exit code 1 even if the user is found,
		// so we only fail if the exit code is different.
		if c.ProcessState.ExitCode() != 1 {
			exitErr(err, "Error getting GitHub username")
		}
	}
	user := b.String()
	// extract username from the output: Hi meling! You've successfully...
	first := strings.Index(user, " ") + 1
	last := strings.Index(user, "!")
	return strings.TrimSpace(user[first:last])
}
