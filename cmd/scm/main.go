package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/autograde/aguis/scm"
	"github.com/urfave/cli"
)

func main() {
	var client scm.SCM

	app := cli.NewApp()
	app.Name = "scm"
	app.Usage = "SCM-agnostic CLI tool"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "provider",
			Usage: "SCM provider to use. [github|gitlab]",
			Value: "github",
		},
		cli.StringFlag{
			Name:   "accesstoken",
			EnvVar: "SCMAccessToken",
			Usage:  "Provider access token.",
		},
	}
	app.Before = setup(&client)
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
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func setup(client *scm.SCM) cli.BeforeFunc {
	return func(c *cli.Context) (err error) {
		if !c.IsSet("provider") {
			return cli.NewExitError("provider must be provided", 3)
		}
		if !c.IsSet("accesstoken") {
			return cli.NewExitError("accesstoken must be provided", 3)
		}
		*client, err = scm.NewSCMClient(c.String("provider"), c.String("accesstoken"))
		return
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
