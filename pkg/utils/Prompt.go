package utils

import (
	"fmt"
	"log"

	"github.com/manifoldco/promptui"
)

func Confirm(message string) bool {
	ask := promptui.Select{
		Label: fmt.Sprintf("%s [y/n]", message),
		Items: []string{"y", "n"},
	}

	_, result, err := ask.Run()
	if err != nil {
		log.Fatal("Prompt failed", err)
	}

	return result == "y"
}
