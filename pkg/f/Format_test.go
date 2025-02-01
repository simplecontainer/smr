package f

import (
	"github.com/go-playground/assert/v2"
	"github.com/simplecontainer/smr/pkg/contracts"
	"testing"
)

func TestNew(t *testing.T) {
	type Wanted struct {
		format contracts.Format
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

func TestNewFromString(t *testing.T) {
	type Wanted struct {
		format contracts.Format
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
		format contracts.Format
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
			"Invalid format",
			func() {
			},
			Wanted{
				string: "",
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
		format contracts.Format
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
			"Invalid format",
			func() {
			},
			Wanted{
				bytes: []byte(""),
			},
			Parameters{
				format: New("", "secret", "", "test"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			formatToString := tc.parameters.format.ToBytes()
			assert.Equal(t, tc.wanted.bytes, formatToString)
		})
	}
}
