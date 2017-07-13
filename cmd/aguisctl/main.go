package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/logger"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/urfave/cli"
)

func main() {
	var db database.GormDB

	app := cli.NewApp()
	app.Name = "aguisctl"
	app.Usage = "CLI tool for interacting with a running aguis instance."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "database",
			Usage: "Path to aguis database",
			Value: tempFile("agdb.db"),
		},
	}
	app.Before = setup(&db)
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
						return db.SetAdmin(c.Uint64("id"))
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
	db.Close()
}

func setup(db *database.GormDB) cli.BeforeFunc {
	return func(c *cli.Context) error {
		l := logrus.New()
		l.Formatter = logger.NewDevFormatter(l.Formatter)
		tdb, err := database.NewGormDB("sqlite3", c.String("database"), database.Logger{Logger: l})
		if err != nil {
			return err
		}
		*db = *tdb
		return nil
	}
}

func tempFile(name string) string {
	return filepath.Join(os.TempDir(), name)
}
