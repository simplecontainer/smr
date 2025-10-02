package ocicredentials

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func NewInputReader() *InputReader {
	return &InputReader{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (ir *InputReader) PromptForCredentials(registry string) (*Credentials, error) {
	credentials := New()
	credentials.Registry = registry

	username, err := ir.promptString("Username", "")
	if err != nil {
		return nil, err
	}
	credentials.Username = username

	if credentials.Username != "" {
		password, err := ir.promptPassword("Password")
		if err != nil {
			return nil, err
		}
		credentials.Password = password
	}

	return credentials, nil
}

func (ir *InputReader) promptString(prompt, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s: ", prompt)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := ir.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}

	return input, nil
}

func (ir *InputReader) promptPassword(prompt string) (string, error) {
	fmt.Printf("%s: ", prompt)

	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	fmt.Println()
	return string(passwordBytes), nil
}

func (ir *InputReader) ReadFromStdin(registry string) (*Credentials, error) {
	credentials := New()
	credentials.Registry = registry

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if err := ir.parseKeyValue(credentials, line); err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return credentials, nil
}

func (ir *InputReader) parseKeyValue(credentials *Credentials, line string) error {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return ErrInvalidInputFormat
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch strings.ToLower(key) {
	case "username", "user":
		credentials.Username = value
	case "password", "pass":
		credentials.Password = value
	default:
		return ErrUnknownConfigKey
	}

	return nil
}
