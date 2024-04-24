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

func (format *FormatStructure) ToString() string {
	return fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key)
}
