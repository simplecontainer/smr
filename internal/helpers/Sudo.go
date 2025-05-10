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

func GetRealUser() RealUser {
	result := RealUser{
		Username: "",
		Uid:      os.Getuid(),
		Gid:      os.Getgid(),
	}

	if sudoUser, exists := os.LookupEnv("SUDO_USER"); exists {
		result.Username = sudoUser

		if u, err := user.Lookup(sudoUser); err == nil {
			if uid, err := strconv.Atoi(u.Uid); err == nil {
				result.Uid = uid
			}
			if gid, err := strconv.Atoi(u.Gid); err == nil {
				result.Gid = gid
			}
		}
		return result
	}

	if sudoUid, exists := os.LookupEnv("SUDO_UID"); exists {
		if uid, err := strconv.Atoi(sudoUid); err == nil {
			result.Uid = uid
		}
	}

	if sudoGid, exists := os.LookupEnv("SUDO_GID"); exists {
		if gid, err := strconv.Atoi(sudoGid); err == nil {
			result.Gid = gid
		}
	}

	if result.Username == "" && result.Uid > 0 {
		if u, err := user.LookupId(strconv.Itoa(result.Uid)); err == nil {
			result.Username = u.Username
		}
	}

	return result
}

func Chown(path string, uid, gid int) error {
	return os.Chown(path, uid, gid)
}
