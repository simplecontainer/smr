package exec

import (
	"bytes"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/types"
	"io/ioutil"
)

func (s *Session) Output(container platforms.IContainer) (types.ExecResult, error) {
	var result types.ExecResult
	var err error
	var stdoutBuffer, stderrBuffer bytes.Buffer

	outputDone := make(chan error)

	go func() {
		// TODO: Here is used docker stdcopy handle this in general way
		_, err = stdcopy.StdCopy(&stdoutBuffer, &stderrBuffer, s.Reader)
		outputDone <- err
	}()

	select {
	case err = <-outputDone:
		if err != nil {
			return result, nil
		}
		break

	case <-s.context.Done():
		return result, nil
	}

	stdout, err := ioutil.ReadAll(&stdoutBuffer)
	if err != nil {
		return result, nil
	}

	stderr, err := ioutil.ReadAll(&stderrBuffer)
	if err != nil {
		return result, nil
	}

	res, err := container.ExecInspect(s.ID)
	if err != nil {
		return result, nil
	}

	result.Exit = res
	result.Stdout = string(stdout)
	result.Stderr = string(stderr)

	return result, nil
}
