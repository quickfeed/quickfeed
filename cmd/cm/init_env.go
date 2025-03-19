package main

import (
	"bufio"
	"cmp"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// initEnv initializes the environment variables for the course.
// The environment variables are saved to a .env file in the root of the git repository.
//
// Run with the following command:
//
//	cm init-env -year 2025 -course dat520 -name "Distributed Systems"
func initEnv(args []string) {
	fs := flag.NewFlagSet(initEnvCmd, flag.ExitOnError)
	var (
		year           int
		course         string
		courseName     string
		discordJoinURL string
		botUser        string
	)
	fs.IntVar(&year, "year", 0, "Course year (required)")
	fs.StringVar(&courseName, "name", "", "Course name (required)")
	fs.StringVar(&course, "course", "", "Course code (default: git repo name)")
	fs.StringVar(&discordJoinURL, "discord-join-url", "", "Discord join URL")
	fs.StringVar(&botUser, "bot-user", "", "Help bot user (default: helpbot)")

	if err := fs.Parse(args); err != nil {
		exitErr(err, "Error parsing flags")
	}

	// if course is not set explicitly, use the git repo name
	if course == "" {
		course = filepath.Base(gitRoot)
	}

	// validate required flags
	if year == 0 || course == "" || courseName == "" {
		fs.Usage()
		os.Exit(1)
	}

	// set environment variables to be saved to .env file
	env := make(map[string]string)
	fs.VisitAll(func(f *flag.Flag) {
		envName := toEnvKey(f.Name)
		envValue := toEnvValue(f.Value.String())
		env[envName] = cmp.Or(envValue, defaultValues[envName])
	})
	if err := saveEnv(env); err != nil {
		exitErr(err, "Error saving environment variables")
	}
}

// toEnvKey converts a flag name to an environment variable key.
func toEnvKey(flagName string) string {
	upper := strings.ToUpper(flagName)
	return strings.ReplaceAll(upper, "-", "_")
}

// toEnvValue converts a flag value to an environment variable value.
func toEnvValue(val string) string {
	// if the provided flag value contains spaces wrap it in quotes
	if strings.Contains(val, " ") {
		return fmt.Sprintf(`"%s"`, val)
	}
	return val
}

// saveEnv saves the env map to the .env file.
// It will not update an existing .env file.
func saveEnv(env map[string]string) error {
	envFile := envFilePath()
	if exists(envFile) {
		return fmt.Errorf("file %s already exists", envFile)
	}
	file, err := os.OpenFile(envFile, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close() // skipcq: GO-S2307

	fmt.Printf("Saving environment variables to %s\n", lastDirFile(envFile))

	for k, v := range env {
		_, err := fmt.Fprintf(file, "%s=%s\n", k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// loadEnv loads environment variables from the .env file.
// Each variable's value is expanded with existing variables from the environment.
// It will not override a variable that already exists in the environment.
func loadEnv() error {
	envFile := envFilePath()
	file, err := os.Open(envFile)
	if err != nil {
		return err
	}
	defer file.Close() // skipcq: GO-S2307

	fmt.Printf("Loading environment variables from %s\n", lastDirFile(envFile))

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if ignore(line) {
			continue
		}
		key, val, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		k := strings.TrimSpace(key)
		if os.Getenv(k) != "" {
			// Ignore .env entries already set in the environment.
			continue
		}
		val = os.ExpandEnv(strings.Trim(strings.TrimSpace(val), `"`))
		os.Setenv(k, val)
	}
	if err := checkRequiredEnv(); err != nil {
		return err
	}
	return scanner.Err()
}

func checkRequiredEnv() error {
	for _, key := range []string{"YEAR", "COURSE", "NAME"} {
		if os.Getenv(key) == "" {
			return fmt.Errorf("required environment variable $%s not set", key)
		}
	}
	return nil
}

func ignore(line string) bool {
	trimmedLine := strings.TrimSpace(line)
	return trimmedLine == "" || strings.HasPrefix(trimmedLine, "#")
}

func envFilePath() string {
	return filepath.Join(gitRoot, ".env")
}

// lastDirFile returns the last directory and file name from a path.
func lastDirFile(path string) string {
	return fmt.Sprintf("%s/%s", filepath.Base(filepath.Dir(path)), filepath.Base(path))
}

func course() string {
	return os.Getenv("COURSE")
}

func name() string {
	return os.Getenv("NAME")
}

func year() string {
	return os.Getenv("YEAR")
}

func courseOrg() string {
	return cmp.Or(os.Getenv("COURSE_ORG"), os.ExpandEnv("$COURSE-$YEAR"))
}

func replacements() map[string]string {
	return map[string]string{
		"COURSE_NAME":      name(),
		"COURSE_ORG":       courseOrg(),
		"BOT_USER":         cmp.Or(os.Getenv("BOT_USER"), defaultValues["BOT_USER"]),
		"DISCORD_JOIN_URL": os.Getenv("DISCORD_JOIN_URL"),
	}
}

var defaultValues = map[string]string{
	"BOT_USER": "helpbot",
}
