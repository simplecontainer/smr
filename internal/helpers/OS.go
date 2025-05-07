package helpers

import (
	"fmt"
	"os"
)

func GetRealHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}

	sudoUser := os.Getenv("SUDO_USER")

	if home == "/root" && sudoUser != "" {
		home = fmt.Sprintf("/home/%s", sudoUser)
	}

	return home
}
