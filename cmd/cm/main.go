package main

import (
	"fmt"
	"os"
)

const (
	initEnvCmd            = "init-env"
	initReposCmd          = "init-repos"
	cloneRepoCmd          = "clone-repo"
	updateDocTagsCmd      = "update-doc-tags"
	removeSolutionTagsCmd = "remove-solution-tags"
	genReadmeCmd          = "gen-readme"
	renameLegacyTestsCmd  = "rename-tests"
	addLintCheckersCmd    = "add-lint-checkers"
	addMainTestsCmd       = "add-main-tests"
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
	case cloneRepoCmd:
		cloneRepo(args)
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
