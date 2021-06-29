package main

import (
	"log"
	"os"
	"path/filepath"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/urfave/cli"
)

// Example usage (to set admin user to the first user registered):
// agctl set admin -id 1

func main() {
	var db database.GormDB

	app := cli.NewApp()
	app.Name = "agctl"
	app.Usage = "CLI tool for interacting with a running autograder instance."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "database",
			Usage: "Path to the autograder database",
			Value: tempFile("ag.db"),
		},
	}
	app.Before = before(&db)
	app.After = after(&db)
	app.Commands = []cli.Command{
		{
			Name:  "set",
			Usage: "Set commands.",
			Subcommands: cli.Commands{
				{
					Name:  "admin",
					Usage: "Set user as administrator. [database id or provider/username]",
					Flags: []cli.Flag{
						cli.Uint64Flag{
							Name:  "id",
							Usage: "User id.",
						},
						cli.StringFlag{
							Name:  "provider",
							Usage: "Remote identity provider.",
						},
						cli.StringFlag{
							Name:  "username",
							Usage: "Remote identity username.",
						},
					},
					Action: func(c *cli.Context) error {
						if !c.IsSet("id") {
							return cli.NewExitError("not implemented", 9)
						}
						return db.UpdateUser(&pb.User{ID: c.Uint64("id"), IsAdmin: true})
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func before(db *database.GormDB) cli.BeforeFunc {
	return func(c *cli.Context) error {
		tdb, err := database.NewGormDB(c.String("database"),
			database.NewGormLogger(database.BuildLogger()),
		)
		if err != nil {
			return err
		}
		*db = *tdb
		return nil
	}
}

func after(db *database.GormDB) cli.AfterFunc {
	return func(c *cli.Context) error {
		return nil
	}
}

func tempFile(name string) string {
	return filepath.Join(os.TempDir(), name)
}
