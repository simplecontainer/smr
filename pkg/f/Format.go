package f

import (
	"fmt"
	"strings"
)

func New(elements ...string) Format {
	builder := ""

	for _, member := range elements {
		builder += member
		builder += "."
	}

	builder = strings.TrimSuffix(builder, ".")
	return NewFromString(builder)
}

func NewFromString(f string) Format {
	elements, nonEmptyCount := BuildElements(strings.SplitN(f, ".", 4))
	format := Format{
		Kind:       strings.TrimSpace(elements[0]),
		Group:      strings.TrimSpace(elements[1]),
		Identifier: strings.TrimSpace(elements[2]),
		Key:        strings.TrimSpace(elements[3]),
		Elems:      nonEmptyCount,
		Category:   strings.TrimSpace(elements[3]),
		Type:       TYPE_FORMATED,
	}

	if format.IsValid() {
		return format
	} else {
		return Format{}
	}
}

func BuildElements(splitted []string) ([]string, int) {
	elements := make([]string, 4)

	lengthSplitted := len(splitted)
	nonEmptyCount := 0

	for k, _ := range elements {
		if k < lengthSplitted {
			elements[k] = splitted[k]

			if splitted[k] != "" {
				nonEmptyCount += 1
			}
		} else {
			elements[k] = ""
		}
	}

	return elements, nonEmptyCount
}

func (format Format) GetCategory() string {
	return format.Category
}

func (format Format) GetType() string {
	return format.Type
}

func (format Format) IsValid() bool {
	split := strings.SplitN(format.ToString(), ".", 4)

	for _, element := range split {
		if element == "" {
			return false
		}
	}

	return true
}

func (format Format) Full() bool {
	return format.Elems == 4
}

func (format Format) ToString() string {
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

func (format Format) ToBytes() []byte {
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
