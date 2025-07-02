package f

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestBuild tests the Build function with various input formats
func TestBuild(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		group       string
		expectError bool
		expected    string
	}{
		{
			name:        "Only kind provided",
			arg:         "kind1",
			group:       "default-group",
			expectError: false,
			expected:    fmt.Sprintf("%s/kind/kind1", static.SMR_PREFIX),
		},
		{
			name:        "Kind and name provided",
			arg:         "kind1/name1",
			group:       "flag-group",
			expectError: false,
			expected:    fmt.Sprintf("%s/kind/kind1/flag-group/name1", static.SMR_PREFIX),
		},
		{
			name:        "Kind, group and name provided",
			arg:         "kind1/group1/name1",
			group:       "flag-group", // This should be ignored as group is provided in arg
			expectError: false,
			expected:    fmt.Sprintf("%s/kind/kind1/group1/name1", static.SMR_PREFIX),
		},
		{
			name:        "Category, kind, group and name provided",
			arg:         "category1/kind1/group1/name1",
			group:       "flag-group", // This should be ignored
			expectError: false,
			expected:    fmt.Sprintf("%s/category1/kind1/group1/name1", static.SMR_PREFIX),
		},
		{
			name:        "Prefix, category, kind, group and name provided",
			arg:         "prefix1/category1/kind1/group1/name1",
			group:       "flag-group", // This should be ignored
			expectError: false,
			expected:    "prefix1/category1/kind1/group1/name1",
		},
		{
			name:        "Empty input",
			arg:         "",
			group:       "flag-group",
			expectError: true,
			expected:    "",
		},
		{
			name:        "Real world example",
			arg:         "simplecontainer.io/v1/state/containers/example/busybox/example-busybox-1",
			group:       "flag-group",
			expectError: false,
			expected:    "simplecontainer.io/v1/state/containers/example/busybox/example-busybox-1",
		},
		{
			name:        "Real world example",
			arg:         "simplecontainer.io/v1/kind/containers/example/busybox/",
			group:       "flag-group",
			expectError: false,
			expected:    "simplecontainer.io/v1/kind/containers/example/busybox",
		},
		{
			name:        "Real world example 2",
			arg:         "containers/example/example-busybox-1",
			group:       "flag-group",
			expectError: false,
			expected:    "simplecontainer.io/v1/kind/containers/example/example-busybox-1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			format, err := Build(tc.arg, tc.group)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, format.ToString())
			}
		})
	}
}

// TestNew tests creating a format from elements
func TestNew(t *testing.T) {
	format := New("prefix", "version", "category", "kind", "group", "name")
	assert.Equal(t, "prefix/version/category/kind/group/name", format.ToString())

	// Test with fewer elements
	format = New("prefix", "category", "kind")
	assert.Equal(t, "prefix/category/kind", format.ToString())

	// Test with empty elements
	format = New("prefix", "", "category", "", "group", "name")
	assert.Equal(t, "", format.ToString())
}

// TestNewFromString tests creating a format from a string
func TestNewFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		valid    bool
	}{
		{
			input:    "prefix/version/category/kind/group/name",
			expected: "prefix/version/category/kind/group/name",
			valid:    true,
		},
		{
			input:    "simplecontainer.io/v1/kind/network/internal/cluster",
			expected: "simplecontainer.io/v1/kind/network/internal/cluster",
			valid:    true,
		},
		{
			input:    "prefix/category/kind",
			expected: "prefix/category/kind",
			valid:    true,
		},
		{
			input:    "prefix//category//name", // Empty elements
			expected: "prefix/category/name",
			valid:    false, // Should be invalid due to empty elements
		},
		{
			input:    "",
			expected: "",
			valid:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			format := NewFromString(tc.input)

			if tc.valid {
				assert.Equal(t, tc.expected, format.ToString())
			} else {
				assert.Equal(t, tc.valid, format.IsValid())
			}
		})
	}
}

// TestFormatGetters tests the getter methods
func TestFormatGetters(t *testing.T) {
	format := NewFromString("prefix/version/category/kind/group/name/field")

	assert.Equal(t, "prefix", format.GetPrefix())
	assert.Equal(t, "version", format.GetVersion())
	assert.Equal(t, "category", format.GetCategory())
	assert.Equal(t, "kind", format.GetKind())
	assert.Equal(t, "group", format.GetGroup())
	assert.Equal(t, "name", format.GetName())
	assert.Equal(t, "field", format.GetField())
	assert.Equal(t, TYPE_FORMATED, format.GetType())
}

// TestShift tests the Shift function
func TestShift(t *testing.T) {
	format := NewFromString("a/b/c/d")
	shifted := format.Shift()

	// Expected result would be a -> kind, b -> group, c -> name

	assert.Equal(t, "a", shifted.GetKind())
	assert.Equal(t, "b", shifted.GetGroup())
	assert.Equal(t, "c", shifted.GetName())
	assert.Equal(t, "d", shifted.GetField())
}

// TestUUID tests UUID handling
func TestUUID(t *testing.T) {
	// Generate a UUID
	id := uuid.New()

	// Create a format with UUID
	formatStr := fmt.Sprintf("%s%s", id.String(), "prefix/category/kind")
	format := NewFromString(formatStr)

	// Check if UUID is parsed correctly
	assert.Equal(t, id, format.GetUUID())

	// Check if UUID is preserved in ToStringWithUUID
	assert.Equal(t, formatStr, format.ToStringWithUUID())

	// Check that ToString does not include UUID
	assert.Equal(t, "prefix/category/kind", format.ToString())
}

// TestIsValid tests the IsValid function
func TestIsValid(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"prefix/category/kind", true},
		{"prefix//kind", false},
		{"", false},
		{"prefix/version/category/kind/group/name", true},
		{"prefix/version/category/kind/group/", true},
		{"/version/category/kind/group/name", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			format := NewFromString(tc.input)
			assert.Equal(t, tc.valid, format.IsValid())
		})
	}
}

// TestCompliant tests the Compliant function
func TestCompliant(t *testing.T) {
	tests := []struct {
		input     string
		compliant bool
	}{
		{"prefix/version/category/kind/group/name", true},
		{"prefix/category/kind", false},
		{"prefix/version/category", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			format := NewFromString(tc.input)
			assert.Equal(t, tc.compliant, format.Compliant())
		})
	}
}

// TestToBytes tests the ToBytes function
func TestToBytes(t *testing.T) {
	format := NewFromString("prefix/category/kind")
	expected := []byte("prefix/category/kind")
	assert.Equal(t, expected, format.ToBytes())
}

// TestBuildElements tests the internal buildElements function
func TestBuildElements(t *testing.T) {
	elements := []string{"a", "b", "", "d", "", "f"}
	_, _, err := buildElements(elements)

	assert.Error(t, err, errors.New("invalid format"))

	elements = []string{"a", "b", "c"}
	_, _, err = buildElements(elements)

	assert.NoError(t, err)
}

// TestParseUUID tests the internal parseUUID function
func TestParseUUID(t *testing.T) {
	// Test with valid UUID
	validUUID := uuid.New()
	input := validUUID.String() + "prefix/category/kind"
	parsedUUID, remainder := parseUUID(input)

	assert.Equal(t, validUUID, parsedUUID)
	assert.Equal(t, "prefix/category/kind", remainder)

	// Test with invalid UUID at the start
	input = "not-a-uuid-prefix/category/kind"
	parsedUUID, remainder = parseUUID(input)

	assert.NotEqual(t, "not-a-uuid", parsedUUID.String())
	assert.Equal(t, input, remainder)

	// Test with short input
	input = "short"
	parsedUUID, remainder = parseUUID(input)

	assert.NotNil(t, parsedUUID)
	assert.Equal(t, input, remainder)
}

// TestEdgeCases tests some edge cases
func TestEdgeCases(t *testing.T) {
	// Test with all empty elements
	format := New("", "", "", "", "", "")
	assert.Equal(t, "", format.ToString())
	assert.False(t, format.IsValid())

	// Test with special characters
	format = NewFromString("pre-fix/ver_sion/cate.gory/k!nd/gr@up/n&me")
	assert.Equal(t, "pre-fix/ver_sion/cate.gory/k!nd/gr@up/n&me", format.ToString())
	assert.True(t, format.IsValid())
}
