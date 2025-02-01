package f

import "github.com/google/uuid"

const TYPE_FORMATED = "f"
const TYPE_UNFORMATED = "u"

type Format struct {
	UUID     uuid.UUID
	Elements []string
	Elems    int
	Type     string
}

type Unformated struct {
	UUID     uuid.UUID
	Prefix   string
	Version  string
	Key      string
	Category string
	Type     string
}
