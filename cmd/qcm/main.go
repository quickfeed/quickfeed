package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
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
	"gopkg.in/yaml.v2"
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
	
	Convert struct {
		Source string `help:"Source directory containing assignment.yml files." required:""`
		DryRun bool   `help:"Show what would be converted without making changes." default:"false"`
	} `cmd:"" help:"Convert assignment.yml files to assignment.json format."`
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

	case "convert":
		convertAssignments()

	default:
		panic(ctx.Command())
	}
}

func runTests(logger *zap.SugaredLogger, client scm.SCM, destDir string) {
	fmt.Printf("Running tests for %s\n", cli.Clone.Lab)
	dockerfile := readFile(destDir, "Dockerfile")

	courseCode := cli.Clone.Course[:len(cli.Clone.Course)-5] // assume course has four digit year (-YYYY)
	course := &qf.Course{
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
	fmt.Println(results.GetBuildInfo().GetBuildLog())
	fmt.Println("***********************")
	// TODO print with tab writer
	for _, score := range results.Scores {
		fmt.Printf("%s: %d/%d (%d)\n", score.GetTestName(), score.GetScore(), score.GetMaxScore(), score.GetWeight())
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

// assignmentData structure for conversion (copied from assignments package)
type assignmentData struct {
	Order            uint32 `yaml:"order" json:"order"`
	Deadline         string `yaml:"deadline" json:"deadline"`
	IsGroupLab       bool   `yaml:"isgrouplab" json:"isgrouplab"`
	AutoApprove      bool   `yaml:"autoapprove" json:"autoapprove"`
	ScoreLimit       uint32 `yaml:"scorelimit" json:"scorelimit"`
	Reviewers        uint32 `yaml:"reviewers" json:"reviewers"`
	ContainerTimeout uint32 `yaml:"containertimeout" json:"containertimeout"`
}

func convertAssignments() {
	sourceDir := cli.Convert.Source
	
	if !exists(sourceDir) {
		fmt.Printf("Error: Source directory '%s' does not exist\n", sourceDir)
		os.Exit(1)
	}
	
	var conversions []conversionInfo
	var errors []string
	
	// Walk the source directory to find assignment.yml/yaml files
	err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if d.IsDir() {
			return nil
		}
		
		filename := d.Name()
		if filename == "assignment.yml" || filename == "assignment.yaml" {
			info, convErr := processAssignmentFile(path)
			if convErr != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", path, convErr))
			} else {
				conversions = append(conversions, info)
			}
		}
		
		return nil
	})
	
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}
	
	if len(errors) > 0 {
		fmt.Printf("Conversion errors found:\n")
		for _, errMsg := range errors {
			fmt.Printf("  %s\n", errMsg)
		}
		os.Exit(1)
	}
	
	if len(conversions) == 0 {
		fmt.Printf("No assignment.yml or assignment.yaml files found in '%s'\n", sourceDir)
		return
	}
	
	fmt.Printf("Found %d assignment files to convert:\n", len(conversions))
	for _, conv := range conversions {
		fmt.Printf("  %s -> %s\n", conv.SourcePath, conv.TargetPath)
	}
	
	if cli.Convert.DryRun {
		fmt.Printf("\nDry run mode: no files were modified\n")
		return
	}
	
	// Perform the actual conversions
	for _, conv := range conversions {
		err := os.WriteFile(conv.TargetPath, conv.JSONContent, 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", conv.TargetPath, err)
			continue
		}
		fmt.Printf("Converted: %s\n", conv.TargetPath)
	}
	
	fmt.Printf("\nConversion complete! Converted %d files.\n", len(conversions))
	fmt.Printf("Note: Original YAML files were not removed. You can delete them manually after verifying the JSON files work correctly.\n")
}

type conversionInfo struct {
	SourcePath  string
	TargetPath  string
	JSONContent []byte
}

func processAssignmentFile(yamlPath string) (conversionInfo, error) {
	var info conversionInfo
	info.SourcePath = yamlPath
	
	// Determine target path (replace .yml/.yaml with .json)
	dir := filepath.Dir(yamlPath)
	info.TargetPath = filepath.Join(dir, "assignment.json")
	
	// Read YAML file
	yamlContent, err := os.ReadFile(yamlPath)
	if err != nil {
		return info, fmt.Errorf("failed to read file: %w", err)
	}
	
	// Parse YAML
	var assignment assignmentData
	err = yaml.Unmarshal(yamlContent, &assignment)
	if err != nil {
		return info, fmt.Errorf("failed to parse YAML: %w", err)
	}
	
	// Convert to JSON with proper formatting
	jsonContent, err := json.MarshalIndent(assignment, "", "  ")
	if err != nil {
		return info, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	info.JSONContent = jsonContent
	return info, nil
}
