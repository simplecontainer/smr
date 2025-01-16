package f

import "github.com/google/uuid"

const TYPE_FORMATED = "f"
const TYPE_UNFORMATED = "u"

type Format struct {
	UUID       uuid.UUID
	Kind       string
	Group      string
	Identifier string
	Key        string
	Elems      int
	Category   string
	Type       string
}

type Unformated struct {
	UUID     uuid.UUID
	Key      string
	Category string
	Type     string
}
