package f

const TYPE_FORMATED = "f"
const TYPE_UNFORMATED = "u"

type Format struct {
	Kind       string
	Group      string
	Identifier string
	Key        string
	Elems      int
	Category   string
	Type       string
}

type Unformated struct {
	Key      string
	Category string
	Type     string
}
