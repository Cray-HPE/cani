package tui

import (
	"github.com/manifoldco/promptui"
)

// CustomConfirmation prompts the user for confirmation
func CustomConfirmation(message string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     message,
		IsConfirm: true,
	}

	_, err := prompt.Run()

	if err != nil {
		if err == promptui.ErrAbort {
			// User chose not to accept the suggestions
			return false, nil
		}
		// An error occurred
		return false, err
	}

	// User chose to continue
	return true, nil
}
