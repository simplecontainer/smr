package f

import (
	"github.com/go-playground/assert/v2"

	"testing"
)

func TestNew(t *testing.T) {
	type Wanted struct {
		format *Format
	}

	type Parameters struct {
		kind       string
		group      string
		identifier string
		key        string
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
				format: New("container", "mysql", "mysql", "object"),
			},
			Parameters{
				kind:       "container",
				group:      "mysql",
				key:        "mysql",
				identifier: "object",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			format := New(tc.parameters.kind, tc.parameters.group, tc.parameters.key, tc.parameters.identifier)
			assert.Equal(t, tc.wanted.format, format)
		})
	}
}

func TestNewFromString(t *testing.T) {
	type Wanted struct {
		format *Format
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
				format: New("container", "mysql", "mysql", "object"),
			},
			Parameters{
				format: "container.mysql.mysql.object",
			},
		},
		{
			"Valid format missing identifier",
			func() {
			},
			Wanted{
				format: New("container", "mysql", "mysql", ""),
			},
			Parameters{
				format: "container.mysql.mysql",
			},
		},
		{
			"Invalid format",
			func() {
			},
			Wanted{
				format: &Format{},
			},
			Parameters{
				format: "..x.x.x..",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			format := NewFromString(tc.parameters.format)
			assert.Equal(t, tc.wanted.format, format)
		})
	}
}

func TestToString(t *testing.T) {
	type Wanted struct {
		string string
	}

	type Parameters struct {
		format *Format
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
				string: "container.mysql.mysql.object",
			},
			Parameters{
				format: New("container", "mysql", "mysql", "object"),
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
				format: New("", "mysql", "mysql", "object"),
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
		format *Format
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
				bytes: []byte("container.mysql.mysql.object"),
			},
			Parameters{
				format: New("container", "mysql", "mysql", "object"),
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
				format: New("", "mysql", "mysql", "object"),
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
