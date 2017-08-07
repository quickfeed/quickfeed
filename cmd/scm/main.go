package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Example usage if you have an organization on github called autograder-test:
// % scm --provider github get repository --all --namespace autograder-test
// OR
// % scm get repository --all --namespace autograder-test

// Another example usage to delete all repos in organzation on github
// % scm delete repository --all --namespace autograder-test

func main() {
	var client scm.SCM
	var db database.GormDB

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
			Name:  "database",
			Usage: "Path to the autograder database",
			Value: tempFile("ag.db"),
		},
		cli.Uint64Flag{
			Name:  "admin",
			Usage: "Admin user id",
			Value: 1,
		},
	}
	app.Before = before(&client, &db)
	app.After = after(&db)
	app.Commands = []cli.Command{
		{
			Name:  "delete",
			Usage: "Delete commands.",
			Subcommands: cli.Commands{
				{
					Name:  "repository",
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
			},
		},
		{
			Name:  "get",
			Usage: "Get commands.",
			Subcommands: cli.Commands{
				{
					Name:  "repository",
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
					Name:  "hook",
					Usage: "Create webhook.",
					Flags: []cli.Flag{
						cli.Uint64Flag{
							Name:  "id",
							Usage: "Repository id. [required by GitLab]",
						},
						cli.StringFlag{
							Name:  "owner",
							Usage: "Repository owner [required by GitHub]",
						},
						cli.StringFlag{
							Name:  "repo",
							Usage: "Repository name. [required by GitHub]",
						},
						cli.StringFlag{
							Name:  "secret",
							Usage: "Webhook secret",
						},
						cli.StringFlag{
							Name:  "url",
							Usage: "Webhook endpoint URL [required]",
						},
					},
					Action: createHook(&client),
				},
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

func before(client *scm.SCM, db *database.GormDB) cli.BeforeFunc {
	return func(c *cli.Context) (err error) {
		l := logrus.New()
		l.Out = ioutil.Discard
		tdb, err := database.NewGormDB("sqlite3", c.String("database"), database.Logger{Logger: l})
		if err != nil {
			return err
		}
		*db = *tdb

		u, err := db.GetUser(c.Uint64("admin"))
		if err != nil {
			return err
		}

		provider := c.String("provider")
		var accessToken string
		for _, ri := range u.RemoteIdentities {
			if ri.Provider == provider {
				accessToken = ri.AccessToken
			}
		}
		if accessToken == "" {
			return fmt.Errorf("access token not found in database for provider %s", provider)
		}
		*client, err = scm.NewSCMClient(provider, accessToken)
		return
	}
}

func after(db *database.GormDB) cli.AfterFunc {
	return func(c *cli.Context) error {
		if db != nil {
			return db.Close()
		}
		return nil
	}
}

func deleteRepositories(client *scm.SCM) cli.ActionFunc {
	ctx := context.Background()

	return func(c *cli.Context) error {
		if !c.IsSet("name") && !c.Bool("all") {
			return cli.NewExitError("name must be provided", 3)
		}
		if !c.IsSet("namespace") {
			return cli.NewExitError("namespace must be provided", 3)
		}
		if c.Bool("all") {
			msg := fmt.Sprintf("Are you sure you want to delete all repositories in %s?", c.String("namespace"))
			if ok, err := confirm(msg); !ok || err != nil {
				fmt.Println("Canceled")
				return err
			}

			repos, err := (*client).GetRepositories(ctx, &scm.Directory{Path: c.String("namespace")})
			if err != nil {
				return err
			}

			for _, repo := range repos {
				var errs []error
				if err := (*client).DeleteRepository(ctx, repo.ID); err != nil {
					errs = append(errs, err)
				} else {
					fmt.Println("Deleted repository", repo.WebURL)
				}
				if len(errs) > 0 {
					return cli.NewMultiError(errs...)
				}
			}
			return nil
		}

		return cli.NewExitError("not implemented", 9)
	}
}

func getRepositories(client *scm.SCM) cli.ActionFunc {
	ctx := context.Background()

	return func(c *cli.Context) error {
		if !c.IsSet("name") && !c.Bool("all") {
			return cli.NewExitError("name must be provided", 3)
		}
		if !c.IsSet("namespace") {
			return cli.NewExitError("namespace must be provided", 3)
		}
		if c.Bool("all") {
			repos, err := (*client).GetRepositories(ctx, &scm.Directory{Path: c.String("namespace")})
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

		return cli.NewExitError("not implemented", 9)
	}
}

// TODO: Validate input.
func createHook(client *scm.SCM) cli.ActionFunc {
	ctx := context.Background()

	return func(c *cli.Context) error {
		return (*client).CreateHook(ctx, &scm.CreateHookOptions{
			URL:    c.String("url"),
			Secret: c.String("secret"),
			Repository: &scm.Repository{
				ID:    c.Uint64("id"),
				Path:  c.String("repo"),
				Owner: c.String("owner"),
			},
		})
	}
}

func createTeam(client *scm.SCM) cli.ActionFunc {
	ctx := context.Background()

	return func(c *cli.Context) error {
		if !c.IsSet("namespace") {
			return cli.NewExitError("namespace must be provided", 3)
		}
		if !c.IsSet("team") {
			return cli.NewExitError("team name must be provided", 3)
		}
		return (*client).CreateTeam(ctx,
			&scm.Directory{Path: c.String("namespace")},
			c.String("team"),
		)
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
