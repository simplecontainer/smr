package manager

import (
	"smr/pkg/cli"
)

func (mgr *Manager) Apply(jsonData string) {
	cli.SendFile("http://localhost:8080/apply", jsonData)
}
