package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"go.uber.org/zap"
)

var cli struct {
	Clone struct {
		Course string `help:"Course organization." default:"dat320-2022"`
		User   string `help:"GitHub user name for student in course." xor:"repo" required:""`
		Group  string `help:"GitHub group name for course." xor:"repo" required:""`
		Token  string `help:"GitHub personal access token." env:"GITHUB_ACCESS_TOKEN"`
		Dir    string `help:"Destination directory for cloned repositories." env:"QUICKFEED_REPOSITORY_PATH"`
		Docker bool   `help:"Run tests using Docker." default:"false"`
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
		logger, client := getSCMClient()
		// Default repository path is $HOME/courses
		destDir := filepath.Join(env.RepositoryPath(), cli.Clone.Course)
		fmt.Printf("Repository path: %s\n", destDir)
		if !exists(destDir) {
			// Only clone if destination directory does not exist
			clone(client, destDir)
		}
		if cli.Clone.Lab != "" {
			// Only run tests if lab is specified
			runTests(logger, client, destDir)
		}

	default:
		panic(ctx.Command())
	}
}

func runTests(logger *zap.SugaredLogger, client scm.SCM, destDir string) {
	fmt.Printf("Running tests for %s\n", cli.Clone.Lab)
	dockerfile := readFile(destDir, "Dockerfile")

	courseCode := cli.Clone.Course[:len(cli.Clone.Course)-5] // assume course has four digit year (-YYYY)
	course := &qf.Course{
		ID:                  1, // Must have an ID field to cache the dockerfile
		Code:                courseCode,
		ScmOrganizationName: cli.Clone.Course,
	}
	course.UpdateDockerfile(dockerfile)

	runData := &ci.RunData{
		Course: course,
		Assignment: &qf.Assignment{
			Name:             cli.Clone.Lab,
			ContainerTimeout: 1, // minutes
		},
		Repo: &qf.Repository{
			HTMLURL: studentRepoURL(),
		},
		JobOwner: studentRepo(),
		CommitID: "dummy",
	}
	if !cli.Clone.Docker {
		runData.EnvVarsFn = func(secret, home string) []string {
			return ci.EnvVars(secret, home, runData.Repo.Name(), runData.Assignment.GetName())
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	results, err := runData.RunTests(ctx, logger, client, runner(logger))
	check(err)

	fmt.Println("***********************")
	fmt.Println(results.BuildInfo.BuildLog)
	fmt.Println("***********************")
	// TODO print with tab writer
	for _, score := range results.Scores {
		fmt.Printf("%s: %d/%d (%d)\n", score.TestName, score.Score, score.MaxScore, score.Weight)
	}
	fmt.Printf("Score sum: %d\n", results.Sum())
}

func runner(logger *zap.SugaredLogger) ci.Runner {
	if cli.Clone.Docker {
		runner, err := ci.NewDockerCI(logger)
		check(err)
		return runner
	}
	return &ci.Local{}
}

func getSCMClient() (*zap.SugaredLogger, scm.SCM) {
	logger, err := qlog.Zap()
	check(err)
	sugar := logger.Sugar()
	client, err := scm.NewSCMClient(sugar, cli.Clone.Token)
	check(err)
	return sugar, client
}

func studentRepo() string {
	studRepo := cli.Clone.Group
	if studRepo == "" {
		studRepo = cli.Clone.User
	}
	return studRepo
}

func studentRepoURL() string {
	repo := qf.RepoURL{ProviderURL: "github.com", Organization: cli.Clone.Course}
	return repo.StudentRepoURL(studentRepo())
}

func clone(client scm.SCM, dstDir string) {
	fmt.Printf("Cloning tests and assignments into %s", dstDir)
	ctx := context.Background()
	clonedAssignmentsRepo, err := client.Clone(ctx, &scm.CloneOptions{
		Organization: cli.Clone.Course,
		Repository:   qf.AssignmentsRepo,
		DestDir:      dstDir,
	})
	check(err)
	fmt.Printf("Successfully cloned assignments repository to: %s", clonedAssignmentsRepo)

	clonedTestsRepo, err := client.Clone(ctx, &scm.CloneOptions{
		Organization: cli.Clone.Course,
		Repository:   qf.TestsRepo,
		DestDir:      dstDir,
	})
	check(err)
	fmt.Printf("Successfully cloned tests repository to: %s", clonedTestsRepo)
}

func readFile(destDir, filename string) string {
	path := filepath.Join(destDir, qf.TestsRepo, "scripts", filename)
	b, err := os.ReadFile(path)
	check(err)
	return string(b)
}

func exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
