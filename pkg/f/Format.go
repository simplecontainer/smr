package f

import (
	"fmt"
	"github.com/google/uuid"
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

func NewFromString(data string) Format {
	UUID, f := parseUUID(data)

	elements, nonEmptyCount := BuildElements(strings.SplitN(f, ".", 4))
	format := Format{
		Kind:       strings.TrimSpace(elements[0]),
		Group:      strings.TrimSpace(elements[1]),
		Identifier: strings.TrimSpace(elements[2]),
		Key:        strings.TrimSpace(elements[3]),
		Elems:      nonEmptyCount,
		Category:   strings.TrimSpace(elements[3]),
		UUID:       UUID,
		Type:       TYPE_FORMATED,
	}

	if format.IsValid() {
		return format
	} else {
		return Format{}
	}
}

func parseUUID(f string) (uuid.UUID, string) {
	if len(f) > 36 {
		UUID, err := uuid.Parse(f[:36])

		if err != nil {
			UUID = uuid.New()

			//Format didn't start with UUID so return new UUID and f as it was since it only had data
			return UUID, f
		}

		//Format started with valid UUID return UUID and rest of the format
		return UUID, f[36:]
	} else {
		UUID := uuid.New()

		//Format didn't start with UUID so return new UUID and f as it was since it only had data
		return UUID, f
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

func (format Format) GetUUID() uuid.UUID {
	return format.UUID
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
func (format Format) ToStringWithUUID() string {
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

	return fmt.Sprintf("%s%s", format.UUID, output)
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
