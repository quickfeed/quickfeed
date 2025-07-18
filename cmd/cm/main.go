package main

import (
	"fmt"
	"os"
)

const (
	initEnvCmd            = "init-env"
	initReposCmd          = "init-repos"
	listReposCmd          = "list-repos"
	cloneRepoCmd          = "clone-repo"
	cloneAllReposCmd      = "clone-all-repos"
	updateDocTagsCmd      = "update-doc-tags"
	removeSolutionTagsCmd = "remove-solution-tags"
	genReadmeCmd          = "gen-readme"
	renameLegacyTestsCmd  = "rename-tests"
	addLintCheckersCmd    = "add-lint-checkers"
	addMainTestsCmd       = "add-main-tests"
	convertAssignmentsCmd = "convert-assignments"
)

func main() {
	if len(os.Args) < 2 {
		usageMsg()
	}

	cmd, args := os.Args[1], os.Args[2:]
	switch cmd {
	case initEnvCmd:
		initEnv(args)
	case initReposCmd:
		initRepos()
	case listReposCmd:
		listRepos(args)
	case cloneRepoCmd:
		cloneRepo(args)
	case cloneAllReposCmd:
		cloneAllRepos()
	case updateDocTagsCmd:
		updateDocTags(args)
	case removeSolutionTagsCmd:
		removeSolutionTags(args)
	case genReadmeCmd:
		genReadme()
	case renameLegacyTestsCmd:
		renameLegacyTests()
	case addLintCheckersCmd:
		addLintCheckers(args)
	case addMainTestsCmd:
		addMainTests()
	case convertAssignmentsCmd:
		convertAssignments(args)
	// case "sync":
	// case "help":
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		usageMsg()
	}
}

func usageMsg() {
	fmt.Println("Usage: cm <command> [options]")
	// TODO(meling): print available commands
	os.Exit(1)
}

func exitErr(err error, msg string) {
	fmt.Printf("%s: %s\n", msg, err)
	os.Exit(1)
}
