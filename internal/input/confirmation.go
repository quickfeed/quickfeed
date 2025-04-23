package input

import (
	"errors"
	"fmt"
	"strings"
)

// AskForConfirmation prompts the user for confirmation with a yes/no question.
func AskForConfirmation(question string) error {
	var answer string
	fmt.Printf("%s (y/n): ", question)
	if _, err := fmt.Scanln(&answer); err != nil {
		return fmt.Errorf("failed to retrieve user input to question: %s, err: %w", question, err)
	}
	if strings.TrimSpace(strings.ToLower(answer)) != "y" {
		return errors.New("aborting operation")
	}
	return nil
}
