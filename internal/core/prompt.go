package core

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func PromptForConfirmation(message string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s [y/n]: ", message)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	// Trim whitespace and convert to lowercase
	response = strings.ToLower(strings.TrimSpace(response))

	// Match common affirmative responses
	yesPattern := regexp.MustCompile(`^(y|yes|yeah|yep|sure|ok|okay|true|t|aye)$`)
	if yesPattern.MatchString(response) {
		return true, nil
	}

	// Match common negative responses
	noPattern := regexp.MustCompile(`^(n|no|nope|nah|false|f|cancel)$`)
	if noPattern.MatchString(response) {
		return false, nil
	}

	// If no match, default to "no" for safety
	return false, fmt.Errorf("response not recognized, interpreting as 'no': %s", response)
}
