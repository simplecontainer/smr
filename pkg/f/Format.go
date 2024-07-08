package f

import (
	"fmt"
	"strings"
)

func New(kind string, group string, identifier string, key string) *Format {
	return &Format{
		Kind:       strings.TrimSpace(kind),
		Group:      strings.TrimSpace(group),
		Identifier: strings.TrimSpace(identifier),
		Key:        strings.TrimSpace(key),
	}
}

func NewFromString(f string) *Format {
	elems := strings.Split(f, ".")
	format := &Format{}

	if len(elems) > 0 {
		format.Kind = strings.TrimSpace(elems[0])
	}

	if len(elems) > 1 {
		format.Group = strings.TrimSpace(elems[1])
	}

	if len(elems) > 2 {
		format.Identifier = strings.TrimSpace(elems[2])
	}

	if len(elems) > 3 {
		format.Key = strings.TrimSpace(elems[3])
	}

	return format
}

func (format *Format) FromString(f string) *Format {
	elems := strings.Split(f, ".")

	if len(elems) > 0 {
		format.Kind = strings.TrimSpace(elems[0])
	}

	if len(elems) > 1 {
		format.Group = strings.TrimSpace(elems[1])
	}

	if len(elems) > 2 {
		format.Identifier = strings.TrimSpace(elems[2])
	}

	if len(elems) > 3 {
		format.Key = strings.TrimSpace(elems[3])
	}

	return format
}

func (format *Format) ToString() string {
	output := ""

	if format.Kind != "" {
		output = fmt.Sprintf("%s", format.Kind)
	}

	if format.Group != "" {
		output = fmt.Sprintf("%s.%s", format.Kind, format.Group)
	}

	if format.Identifier != "" {
		output = fmt.Sprintf("%s.%s.%s", format.Kind, format.Group, format.Identifier)
	}

	if format.Key != "" {
		output = fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key)
	}

	return output
}

func (format *Format) ToBytes() []byte {
	output := ""

	if format.Kind != "" {
		output = fmt.Sprintf("%s", format.Kind)
	}

	if format.Group != "" {
		output = fmt.Sprintf("%s.%s", format.Kind, format.Group)
	}

	if format.Identifier != "" {
		output = fmt.Sprintf("%s.%s.%s", format.Kind, format.Group, format.Identifier)
	}

	if format.Key != "" {
		output = fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key)
	}

	return []byte(output)
}
