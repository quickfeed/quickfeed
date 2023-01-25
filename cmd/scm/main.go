package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"

	"github.com/urfave/cli"
)

// To use this tool, there are two options:
// (1) you either need to have an existing ag.db database file for a running
//     AG server instance with the appropriate access tokens for an admin user.
// (2) you need to set up a GITHUB_ACCESS_TOKEN environment variable
//     for your organization. To use this option with GitHub navigate to
//     Settings -> Developer settings -> Personal access tokens and from
//     there generate a new token. Copy this token to the GITHUB_ACCESS_TOKEN
//     environment variable.
//
// Example usage if you have an organization on github called qf101:
// % scm --provider github get repo -all -namespace qf101
// OR
// % scm get repo -all -namespace qf101
//
// Another example usage to delete all repos in organization on github
// % scm delete repo -all -namespace qf101
//
// Here is an example usage for creating a team with two members
// % scm create team -namespace qf101 -team teachers -users s111,meling
//
// Here is how to fetch the login name of a specific user id:
// % scm get user -id 810999
// OR to fetch the login name of the currently logged in user:
// % scm get user

func main() {
	var client scm.GithubSCM

	app := cli.NewApp()
	app.Name = "scm"
	app.Usage = "SCM-agnostic CLI tool."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "provider",
			Usage: "SCM provider to use. [github|gitlab]",
			Value: "github",
		},
		cli.StringFlag{
			Name:  "token",
			Usage: "Environment variable with access token.",
			Value: "GITHUB_ACCESS_TOKEN",
		},
		cli.StringFlag{
			Name:  "database",
			Usage: "Path to the quickfeed database",
			Value: tempFile("ag.db"),
		},
		cli.Uint64Flag{
			Name:  "admin",
			Usage: "Admin user id",
			Value: 1,
		},
	}
	app.Before = before(&client)
	app.Commands = []cli.Command{
		{
			Name:  "delete",
			Usage: "Delete commands.",
			Subcommands: cli.Commands{
				{
					Name:  "repo",
					Usage: "Delete repositories.",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Usage: "Repository name.",
						},
						cli.StringFlag{
							Name:  "namespace",
							Usage: "Where to find the repository, i.e., user/group/organization.",
						},
						cli.BoolFlag{
							Name:  "all",
							Usage: "Delete all repositories in namespace.",
						},
					},
					Action: deleteRepositories(&client),
				},
				{
					Name:  "team",
					Usage: "Delete teams.",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Usage: "Team name.",
						},
						cli.StringFlag{
							Name:  "namespace",
							Usage: "Organization the team belongs to.",
						},
						cli.BoolFlag{
							Name:  "all",
							Usage: "Delete all teams in namespace.",
						},
					},
					Action: deleteTeams(&client),
				},
			},
		},
		{
			Name:  "get",
			Usage: "Get commands.",
			Subcommands: cli.Commands{
				{
					Name:  "repo",
					Usage: "Get repository information.",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Usage: "Repository name.",
						},
						cli.StringFlag{
							Name:  "namespace",
							Usage: "Where to find the repository, i.e., user/group/organization.",
						},
						cli.BoolFlag{
							Name:  "all",
							Usage: "Get all repositories in namespace.",
						},
					},
					Action: getRepositories(&client),
				},
			},
		},
		{
			Name:  "create",
			Usage: "Create commands.",
			Subcommands: cli.Commands{
				{
					Name:  "team",
					Usage: "Create team.",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "namespace",
							Usage: "Where to find the repository, i.e., user/group/organization.",
						},
						cli.StringFlag{
							Name:  "team",
							Usage: "Team name",
						},
						cli.StringFlag{
							Name:  "users",
							Usage: "User names to add to team",
						},
					},
					Action: createTeam(&client),
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func before(client *scm.GithubSCM) cli.BeforeFunc {
	return func(c *cli.Context) (err error) {
		accessToken := os.Getenv(c.String("token"))
		if accessToken == "" {
			return fmt.Errorf("required access token not provided")
		}
		logger, err := qlog.Zap()
		if err != nil {
			return err
		}
		*client = *scm.NewGithubSCMClient(logger.Sugar(), accessToken)
		return
	}
}

func deleteRepositories(client *scm.GithubSCM) cli.ActionFunc {
	ctx := context.Background()

	return func(c *cli.Context) error {
		if !c.IsSet("name") && !c.Bool("all") {
			return cli.NewExitError("name must be provided", 3)
		}
		if !c.IsSet("namespace") {
			return cli.NewExitError("namespace must be provided", 3)
		}
		if c.IsSet("name") && !c.IsSet("namespace") {
			return cli.NewExitError("name and namespace must be provided", 3)
		}
		if c.Bool("all") {
			msg := fmt.Sprintf("Are you sure you want to delete all repositories in %s?", c.String("namespace"))
			if ok, err := confirm(msg); !ok || err != nil {
				fmt.Println("Canceled")
				return err
			}

			repos, err := (*client).GetRepositories(ctx, &qf.Organization{ScmOrganizationName: c.String("namespace")})
			if err != nil {
				return err
			}

			for _, repo := range repos {
				var errs []error
				if _, err := (*client).Client().Repositories.Delete(ctx, repo.Owner, repo.Path); err != nil {
					errs = append(errs, err)
				} else {
					fmt.Println("Deleted repository", repo.HTMLURL)
				}
				if len(errs) > 0 {
					return cli.NewMultiError(errs...)
				}
			}
			return nil
		}
		if _, err := (*client).Client().Repositories.Delete(ctx, c.String("namespace"), c.String("name")); err != nil {
			return err
		}
		fmt.Println("Deleted repository ", c.String("name"), " on organization ", c.String("namespace"))
		return cli.NewExitError("not implemented", 9)
	}
}

func getRepositories(client *scm.GithubSCM) cli.ActionFunc {
	ctx := context.Background()

	return func(c *cli.Context) error {
		if !c.IsSet("name") && !c.Bool("all") {
			return cli.NewExitError("name must be provided", 3)
		}
		if !c.IsSet("namespace") {
			return cli.NewExitError("namespace must be provided", 3)
		}
		if c.IsSet("name") && !c.IsSet("namespace") {
			return cli.NewExitError("name and namespace must be provided", 3)
		}
		if c.Bool("all") {
			repos, err := (*client).GetRepositories(ctx, &qf.Organization{ScmOrganizationName: c.String("namespace")})
			if err != nil {
				return err
			}
			s, err := toJSON(&repos)
			if err != nil {
				return err
			}
			fmt.Println(s)
			return nil
		}
		repo, _, err := (*client).Client().Repositories.Get(ctx, c.String("namespace"), c.String("name"))
		if err != nil {
			return err
		}
		fmt.Println("Found repository ", *repo.HTMLURL)
		return nil
	}
}

func createTeam(client *scm.GithubSCM) cli.ActionFunc {
	ctx := context.Background()

	return func(c *cli.Context) error {
		if !c.IsSet("namespace") {
			return cli.NewExitError("namespace must be provided", 3)
		}
		if !c.IsSet("team") {
			return cli.NewExitError("team name must be provided", 3)
		}
		if !c.IsSet("users") {
			return cli.NewExitError("team user names must be provided (comma separated)", 3)
		}
		users := strings.Split(c.String("users"), ",")
		if len(users) < 1 {
			return cli.NewExitError("team user names must be provided (comma separated)", 3)
		}
		opt := github.NewTeam{
			Name: c.String("team"),
		}
		_, _, err := (*client).Client().Teams.CreateTeam(ctx, c.String("namespace"), opt)
		for _, user := range users {
			_, _, err := (*client).Client().Teams.AddTeamMembershipBySlug(ctx, c.String("namespace"), c.String("team"), user, &github.TeamAddTeamMembershipOptions{})
			if err != nil {
				return err
			}
		}
		return err
	}
}

func deleteTeams(client *scm.GithubSCM) cli.ActionFunc {
	ctx := context.Background()

	return func(c *cli.Context) error {
		if !c.IsSet("name") && !c.Bool("all") {
			return cli.NewExitError("name must be provided", 3)
		}
		if !c.IsSet("namespace") {
			return cli.NewExitError("namespace must be provided", 3)
		}
		if c.Bool("all") {
			msg := fmt.Sprintf("Are you sure you want to delete all teams in %s?", c.String("namespace"))
			if ok, err := confirm(msg); !ok || err != nil {
				fmt.Println("Canceled")
				return err
			}

			teams, _, err := (*client).Client().Teams.ListTeams(ctx, c.String("namespace"), &github.ListOptions{})
			if err != nil {
				return err
			}

			for _, team := range teams {
				var errs []error
				if _, err := (*client).Client().Teams.DeleteTeamBySlug(ctx, c.String("namespace"), *team.Name); err != nil {
					errs = append(errs, err)
				} else {
					fmt.Println("Deleted team", *team.Name)
				}
				if len(errs) > 0 {
					return cli.NewMultiError(errs...)
				}
			}
			return nil
		}
		// delete team by name
		teamName := c.String("name")
		msg := fmt.Sprintf("Are you sure you want to delete team %s in %s?", teamName, c.String("namespace"))
		if ok, err := confirm(msg); !ok || err != nil {
			fmt.Println("Canceled")
			return err
		}
		_, err := (*client).Client().Teams.DeleteTeamBySlug(ctx, c.String("namespace"), teamName)
		return err
	}
}

func toJSON(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func confirm(msg string) (bool, error) {
	fmt.Printf("%s (y/N): ", msg)

	var input string
	if _, err := fmt.Scan(&input); err != nil {
		return false, err
	}

	input = strings.TrimSpace(input)
	input = strings.ToLower(input)
	if input == "y" || input == "yes" {
		return true, nil
	}
	return false, nil
}

func tempFile(name string) string {
	return filepath.Join(os.TempDir(), name)
}
