package f

import "github.com/google/uuid"

const TYPE_FORMATED = "f"
const TYPE_UNFORMATED = "u"

type Format struct {
	UUID       uuid.UUID
	Prefix     string
	Category   string
	Kind       string
	Group      string
	Identifier string
	//Kind       string
	//Group      string
	//Identifier string
	//Key        string
	Elems int
	Type  string
}

type Unformated struct {
	UUID     uuid.UUID
	Prefix   string
	Key      string
	Category string
	Type     string
}
