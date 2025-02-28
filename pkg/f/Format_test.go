package f

import (
	"github.com/go-playground/assert/v2"
	"github.com/simplecontainer/smr/pkg/contracts/iformat"
	"testing"
)

func TestNew(t *testing.T) {
	type Wanted struct {
		format iformat.Format
	}

	type Parameters struct {
		prefix   string
		version  string
		category string
		kind     string
		group    string
		name     string
	}

	testCases := []struct {
		name       string
		mockFunc   func()
		wanted     Wanted
		parameters Parameters
	}{
		{
			"Valid configuration",
			func() {
			},
			Wanted{
				format: New("simplecontainer.io", "v1", "secret", "secret", "test", "test"),
			},
			Parameters{
				prefix:   "simplecontainer.io",
				version:  "v1",
				category: "secret",
				kind:     "secret",
				group:    "test",
				name:     "test",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			format := New(tc.parameters.prefix, tc.parameters.version, tc.parameters.category, tc.parameters.kind, tc.parameters.group, tc.parameters.name)

			// UUID will be different for two new formats so match them to pass test
			format.UUID = tc.wanted.format.GetUUID()

			assert.Equal(t, tc.wanted.format.ToString(), format.ToString())
			assert.Equal(t, tc.wanted.format.ToBytes(), format.ToBytes())
			assert.Equal(t, tc.wanted.format.ToStringWithUUID(), format.ToStringWithUUID())
		})
	}
}

func TestInverse(t *testing.T) {
	type Wanted struct {
		format iformat.Format
	}

	type Parameters struct {
		format string
	}

	testCases := []struct {
		name       string
		mockFunc   func()
		wanted     Wanted
		parameters Parameters
	}{
		{
			"Valid format",
			func() {
			},
			Wanted{
				format: New("simplecontainer.io", "v1", "secret", "secret", "test", "test"),
			},
			Parameters{
				format: "secret/test/test",
			},
		}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			format := NewFromString(tc.parameters.format).Inverse().(Format)

			assert.Equal(t, tc.wanted.format.GetKind(), format.GetKind())
			assert.Equal(t, tc.wanted.format.GetGroup(), format.GetGroup())
			assert.Equal(t, tc.wanted.format.GetName(), format.GetName())
		})
	}
}

func TestNewFromString(t *testing.T) {
	type Wanted struct {
		format iformat.Format
	}

	type Parameters struct {
		format string
	}

	testCases := []struct {
		name       string
		mockFunc   func()
		wanted     Wanted
		parameters Parameters
	}{
		{
			"Valid format",
			func() {
			},
			Wanted{
				format: New("simplecontainer.io", "v1", "secret", "secret", "test", "test"),
			},
			Parameters{
				format: "simplecontainer.io/v1/secret/secret/test/test",
			},
		},
		{
			"Valid format missing identifier",
			func() {
			},
			Wanted{
				format: New("simplecontainer.io", "v1", "secret", "secret", "test"),
			},
			Parameters{
				format: "simplecontainer.io/v1/secret/secret/test",
			},
		},
		{
			"Invalid format",
			func() {
			},
			Wanted{
				format: NewFromString(""),
			},
			Parameters{
				format: "//x/x/x//",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			format := NewFromString(tc.parameters.format)

			assert.Equal(t, tc.wanted.format.ToString(), format.ToString())
			assert.Equal(t, tc.wanted.format.ToBytes(), format.ToBytes())
		})
	}
}

func TestToString(t *testing.T) {
	type Wanted struct {
		string string
	}

	type Parameters struct {
		format iformat.Format
	}

	testCases := []struct {
		name       string
		mockFunc   func()
		wanted     Wanted
		parameters Parameters
	}{
		{
			"Valid format",
			func() {
			},
			Wanted{
				string: "simplecontainer.io/v1/secret/secret/test/test",
			},
			Parameters{
				format: New("simplecontainer.io", "v1", "secret", "secret", "test", "test"),
			},
		},
		{
			"Format with empty spaces",
			func() {
			},
			Wanted{
				string: "secret/test",
			},
			Parameters{
				format: New("", "secret", "", "test"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			formatToString := tc.parameters.format.ToString()
			assert.Equal(t, tc.wanted.string, formatToString)
		})
	}
}

func TestToBytes(t *testing.T) {
	type Wanted struct {
		bytes []byte
	}

	type Parameters struct {
		format iformat.Format
	}

	testCases := []struct {
		name       string
		mockFunc   func()
		wanted     Wanted
		parameters Parameters
	}{
		{
			"Valid format",
			func() {
			},
			Wanted{
				bytes: []byte("simplecontainer.io/v1/secret/secret/test/test"),
			},
			Parameters{
				format: New("simplecontainer.io", "v1", "secret", "secret", "test", "test"),
			},
		},
		{
			"Format with spaces",
			func() {
			},
			Wanted{
				bytes: []byte("secret/test"),
			},
			Parameters{
				format: New("", "secret", "", "test"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			formatToBytes := tc.parameters.format.ToBytes()
			assert.Equal(t, tc.wanted.bytes, formatToBytes)
		})
	}
}
