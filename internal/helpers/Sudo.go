package helpers

import (
	"os"
	"os/user"
	"strconv"
)

type RealUser struct {
	Username string
	Uid      int
	Gid      int
}

func IsRunningAsSudo() bool {
	_, hasSudoUser := os.LookupEnv("SUDO_USER")

	euid := os.Geteuid()
	ruid := os.Getuid()

	return hasSudoUser || (euid == 0 && euid != ruid)
}

func GetRealHome() string {
	u, err := GetRealUser()

	if err != nil {
		return os.Getenv("HOME")
	}

	if u.HomeDir != "" {
		return u.HomeDir
	} else {
		return os.Getenv("HOME")
	}
}

func GetRealUser() (*user.User, error) {
	if sudoUser, exists := os.LookupEnv("SUDO_USER"); exists {
		u, err := user.Lookup(sudoUser)

		if err != nil {
			return nil, err
		}

		return u, nil
	} else {
		return user.Current()
	}
}

func Chown(path string, uid, gid string) error {
	UID, err := strconv.Atoi(uid)

	if err != nil {
		return err
	}

	GID, err := strconv.Atoi(gid)

	if err != nil {
		return err
	}

	return os.Chown(path, UID, GID)
}
