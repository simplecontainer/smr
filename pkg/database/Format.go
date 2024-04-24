package database

import (
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
