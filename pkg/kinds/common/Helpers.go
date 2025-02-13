package common

import "fmt"

func GroupIdentifier(group string, name string) string {
	return fmt.Sprintf("%s-%s", group, name)
}
