package formaters

import (
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
)

func GitopsBuilder(objects []json.RawMessage) ([]implementation.Gitops, error) {
	var gitopsObjs = make([]implementation.Gitops, 0)

	for _, obj := range objects {
		gitopsObj := implementation.Gitops{}

		err := json.Unmarshal(obj, &gitopsObj)
		if err != nil {
			fmt.Println(err)
			continue
		}

		gitopsObjs = append(gitopsObjs, gitopsObj)
	}

	return gitopsObjs, nil
}
