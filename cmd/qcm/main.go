package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
)

var cli struct {
	Clone struct {
		Course string `help:"Course organization." default:"dat320-2022"`
		User   string `help:"GitHub user name for student in course." xor:"repo" required:""`
		Group  string `help:"GitHub group name for course." xor:"repo" required:""`
		Token  string `help:"GitHub personal access token." env:"GITHUB_ACCESS_TOKEN"`
		Dir    string `help:"Destination directory for cloned repositories." default:"."`
		Merge  bool   `help:"Merge tests repository into user/group directory." default:"false"`
		Lab    string `help:"Assignment to test."`
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
		studRepo := studentRepo()
		destDir := filepath.Join(cli.Clone.Dir, cli.Clone.Course)
		clone(client, []string{studRepo, qf.AssignmentRepo, qf.TestsRepo}, destDir)
		if cli.Clone.Merge {
			fmt.Printf("Merging: %s -> %s\n", qf.TestsRepo, qf.AssignmentRepo)
			err := copyDir(filepath.Join(destDir, qf.TestsRepo), filepath.Join(destDir, qf.AssignmentRepo))
			check(err)
			fmt.Printf("Merging: %s -> %s\n", qf.TestsRepo, studRepo)
			err = copyDir(filepath.Join(destDir, qf.TestsRepo), filepath.Join(destDir, studRepo))
			check(err)
		}
		if cli.Clone.Lab != "" {
			fmt.Printf("Running tests for %s\n", cli.Clone.Lab)
			scriptContent, err := os.ReadFile(filepath.Join(destDir, qf.AssignmentRepo, "scripts", "run.sh"))
			check(err)
			// destDir = home folder for the test execution
			home := filepath.Join(os.Getenv("PWD"), destDir)
			envVars := ci.EnvVars("secret", home, studRepo, cli.Clone.Lab)
			for _, envVar := range envVars {
				fmt.Println(envVar)
			}
			job, err := ci.ParseRunScript(string(scriptContent), envVars)
			check(err)
			runner := ci.Local{}
			out, err := runner.Run(context.Background(), job)
			check(err)
			fmt.Println(out)
		}

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

func studentRepo() string {
	studRepo := cli.Clone.Group
	if studRepo == "" {
		studRepo = qf.StudentRepoName(cli.Clone.User)
	}
	return studRepo
}

func clone(client scm.SCM, cloneRepos []string, dstDir string) {
	for _, repo := range cloneRepos {
		_, err := client.Clone(context.Background(), &scm.CloneOptions{
			Organization: cli.Clone.Course,
			Repository:   repo,
			DestDir:      dstDir,
		})
		check(err)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
