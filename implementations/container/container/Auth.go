package container

import (
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/runtime"
	"os"
	"strings"
)

func GetAuth(image string, runtime *runtime.Runtime) string {
	dockerConfig := fmt.Sprintf("%s/%s", runtime.HOMEDIR, ".docker/config.json")
	if _, err := os.Stat(dockerConfig); err == nil {
		body, err := os.ReadFile(dockerConfig)
		if err != nil {
			panic("Unable to read docker auth file")
		}

		config := map[string]map[string]map[string]string{}
		err = json.Unmarshal(body, &config)

		if err != nil {
			panic(err)
		}

		for registry, auth := range config["auths"] {
			if strings.Contains(image, registry) {
				return auth["auth"]
			}
		}

		return ""
	} else {
		return ""
	}
}
