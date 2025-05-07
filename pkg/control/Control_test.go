package control

import (
	"encoding/json"
	"github.com/simplecontainer/smr/pkg/control/controls/drain"
	"gotest.tools/v3/assert"
	"testing"
)

func Test(t *testing.T) {
	batch := NewCommandBatch()

	params := map[string]string{"force": "true"}
	batch.AddCommand(drain.NewDrainCommand(params))

	bytes, err := json.Marshal(batch)

	if err != nil {
		t.Error(err)
	}

	batchOverWire := &CommandBatch{}

	err = json.Unmarshal(bytes, batchOverWire)

	if err != nil {
		t.Error(err)
	}

	for _, cmd := range batchOverWire.Commands {
		assert.Equal(t, cmd.Name(), "drain")
		assert.DeepEqual(t, cmd.Data(), params)
	}
}
