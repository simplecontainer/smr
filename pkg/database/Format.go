package database

import (
	"fmt"
	"strings"
)

func Format(kind string, group string, identifier string, key string) FormatStructure {
	return FormatStructure{
		Kind:       strings.TrimSpace(kind),
		Group:      strings.TrimSpace(group),
		Identifier: strings.TrimSpace(identifier),
		Key:        strings.TrimSpace(key),
	}
}

func FormatEmpty() *FormatStructure {
	return &FormatStructure{}
}

func (format *FormatStructure) FromString(f string) FormatStructure {
	elems := strings.Split(f, ".")

	if len(elems) > 3 {
		format.Kind = strings.TrimSpace(elems[0])
		format.Group = strings.TrimSpace(elems[1])
		format.Identifier = strings.TrimSpace(elems[2])
		format.Key = strings.TrimSpace(elems[3])
	}

	return *format
}

func (format *FormatStructure) ToString() string {
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

func (format *FormatStructure) ToBytes() []byte {
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
