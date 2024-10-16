package gitops

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/simplecontainer/smr/implementations/gitops/shared"
	"github.com/simplecontainer/smr/implementations/gitops/status"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/plugins"
	"net/http"
	"reflect"
)

func (gitops *Gitops) Run(operation string, args ...interface{}) httpcontract.ResponseOperator {
	reflected := reflect.TypeOf(gitops)
	reflectedValue := reflect.ValueOf(gitops)

	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)

		if operation == method.Name {
			inputs := make([]reflect.Value, len(args))

			for i, _ := range args {
				inputs[i] = reflect.ValueOf(args[i])
			}

			returnValue := reflectedValue.MethodByName(operation).Call(inputs)

			return returnValue[0].Interface().(httpcontract.ResponseOperator)
		}
	}

	return httpcontract.ResponseOperator{
		HttpStatus:       400,
		Explanation:      "server doesn't support requested functionality",
		ErrorExplanation: "implementation is missing",
		Error:            true,
		Success:          false,
		Data:             nil,
	}
}

func (gitops *Gitops) ListSupported(args ...interface{}) httpcontract.ResponseOperator {
	reflected := reflect.TypeOf(gitops)

	supportedOperations := map[string]any{}
	supportedOperations["SupportedOperations"] = []string{}

OUTER:
	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)
		for _, forbiddenOperator := range invalidOperators {
			if forbiddenOperator == method.Name {
				continue OUTER
			}
		}

		supportedOperations["SupportedOperations"] = append(supportedOperations["SupportedOperations"].([]string), method.Name)
	}

	return httpcontract.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             supportedOperations,
	}
}

func (gitops *Gitops) List(request httpcontract.RequestOperator) httpcontract.ResponseOperator {
	data := make(map[string]any)

	pl := plugins.GetPlugin(request.Manager.Config.OptRoot, "gitops.so")
	sharedObj := pl.GetShared().(*shared.Shared)

	for key, gitopsInstance := range sharedObj.Watcher.Repositories {
		data[key] = gitopsInstance.Gitops
	}

	return httpcontract.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "list of the gitops objects",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             data,
	}
}

func (gitops *Gitops) Get(request httpcontract.RequestOperator) httpcontract.ResponseOperator {
	if request.Data == nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	format := f.NewFromString(fmt.Sprintf("%s.%s.%s.%s", KIND, request.Data["group"], request.Data["identifier"], "object"))

	obj := objects.New(request.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "gitops definition is not found on the server",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	definitionObject := obj.GetDefinition()

	var definition = make(map[string]any)
	definition["kind"] = KIND
	definition[KIND] = definitionObject

	return httpcontract.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "gitops object is found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             definition,
	}
}

func (gitops *Gitops) Delete(request httpcontract.RequestOperator) httpcontract.ResponseOperator {
	if request.Data == nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", request.Data["group"], request.Data["identifier"])

	format := f.New("gitops", request.Data["group"].(string), request.Data["identifier"].(string), "object")

	obj := objects.New(request.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "object database failed to process request",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	if !obj.Exists() {
		return httpcontract.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "object not found on the server",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
		}
	}

	_, err = obj.Remove(format)

	if err != nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "object removal failed",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	pl := plugins.GetPlugin(request.Manager.Config.OptRoot, "gitops.so")
	sharedObj := pl.GetShared().(*shared.Shared)

	gitopsWatcher := sharedObj.Watcher.Find(GroupIdentifier)

	if gitopsWatcher == nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "gitops definition doesn't exists",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	} else {
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_PENDING_DELETE)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}

	return httpcontract.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "gitops is transitioned to the pending delete state",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}

func (gitops *Gitops) Refresh(request httpcontract.RequestOperator) httpcontract.ResponseOperator {
	if request.Data == nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", request.Data["group"], request.Data["identifier"])

	pl := plugins.GetPlugin(request.Manager.Config.OptRoot, "gitops.so")
	sharedObj := pl.GetShared().(*shared.Shared)

	gitopsWatcher := sharedObj.Watcher.Find(GroupIdentifier)

	if gitopsWatcher == nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "gitops definition doesn't exists",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	} else {
		gitopsWatcher.Gitops.ForcePoll = true
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}

	return httpcontract.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "refresh is triggered manually",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}

func (gitops *Gitops) Sync(request httpcontract.RequestOperator) httpcontract.ResponseOperator {
	if request.Data == nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", request.Data["group"], request.Data["identifier"])

	pl := plugins.GetPlugin(request.Manager.Config.OptRoot, "gitops.so")
	sharedObj := pl.GetShared().(*shared.Shared)

	gitopsWatcher := sharedObj.Watcher.Find(GroupIdentifier)

	if gitopsWatcher == nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "gitops definition doesn't exists",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	} else {
		if gitopsWatcher.Gitops.AutomaticSync == false {
			gitopsWatcher.Gitops.ManualSync = true
		}

		gitopsWatcher.Gitops.ForcePoll = true
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}

	return httpcontract.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "sync is triggered manually",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}
