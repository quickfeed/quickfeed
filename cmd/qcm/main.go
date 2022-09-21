package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
)

var cli struct {
	Clone struct {
		Course string `help:"Course code." default:"dat320-2022"`
		User   string `help:"GitHub user name for student in course." xor:"repo" required:""`
		Group  string `help:"GitHub group name for course." xor:"repo" required:""`
		Token  string `help:"GitHub personal access token." env:"GITHUB_ACCESS_TOKEN"`
		Dir    string `help:"Destination directory for cloned repositories." default:"."`
		Merge  bool   `help:"Merge tests repository into user/group directory." default:"false"`
	} `cmd:"" help:"Clone repositories for local test execution."`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("qcm"),
		kong.Description("QuickFeed course manager."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	switch ctx.Command() {
	case "clone":
		client := getSCMClient()
		err := clone(client)
		check(err)
		fmt.Printf("merge: %t\n", cli.Clone.Merge)
	default:
		panic(ctx.Command())
	}
}

func getSCMClient() scm.SCM {
	logger, err := qlog.Zap()
	check(err)
	client, err := scm.NewSCMClient(logger.Sugar(), cli.Clone.Token)
	check(err)
	return client
}

func clone(client scm.SCM) error {
	repo := cli.Clone.Group
	if repo == "" {
		repo = qf.StudentRepoName(cli.Clone.User)
	}
	dir := filepath.Join(cli.Clone.Dir, cli.Clone.Course)
	fmt.Printf("course=%s, repo=%s, dir=%s\n", cli.Clone.Course, repo, dir)

	cloneDir, err := client.Clone(context.Background(), &scm.CloneOptions{
		Organization: cli.Clone.Course,
		Repository:   repo,
		DestDir:      dir,
	})
	if err != nil {
		return err
	}
	assignmentsDir, err := client.Clone(context.Background(), &scm.CloneOptions{
		Organization: cli.Clone.Course,
		Repository:   qf.AssignmentRepo,
		DestDir:      dir,
	})
	if err != nil {
		return err
	}
	testsDir, err := client.Clone(context.Background(), &scm.CloneOptions{
		Organization: cli.Clone.Course,
		Repository:   qf.TestsRepo,
		DestDir:      dir,
	})
	if err != nil {
		return err
	}
	fmt.Printf("cloneDir       %s\n", cloneDir)
	fmt.Printf("assignmentsDir %s\n", assignmentsDir)
	fmt.Printf("testsDir       %s\n", testsDir)
	return nil
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
