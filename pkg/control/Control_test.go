package control

import (
	"encoding/json"
	"github.com/simplecontainer/smr/pkg/control/drain"
	"gotest.tools/v3/assert"
	"testing"
)

func Test(t *testing.T) {
	b := NewCommandBatch()

	params := map[string]string{"force": "true"}
	b.AddCommand(drain.NewDrainCommand(params))

	bytes, err := json.Marshal(b)

	if err != nil {
		t.Error(err)
	}

	batchOverWire := &CommandBatch{}

	err = json.Unmarshal(bytes, batchOverWire)

	if err != nil {
		t.Error(err)
	}

	for _, cmd := range batchOverWire.GetCommands() {
		assert.Equal(t, cmd.Name(), "drain")
		assert.DeepEqual(t, cmd.Data(), params)
	}
}
