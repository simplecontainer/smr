package f

/*
	All database keys should follow the format:
	prefix/version/category/kind/group/identifier/key
*/

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
)

func New(elements ...string) Format {
	builder := ""

	for _, member := range elements {
		builder += member
		builder += "/"
	}

	builder = strings.TrimSuffix(builder, "/")
	return NewFromString(builder)
}

func NewFromString(data string) Format {
	UUID, f := parseUUID(data)

	elements, nonEmptyCount := buildElements(strings.SplitN(f, "/", 6))
	format := Format{
		Elements: elements,
		Elems:    nonEmptyCount,
		UUID:     UUID,
		Type:     TYPE_FORMATED,
	}

	if format.IsValid() {
		return format
	} else {
		return Format{}
	}
}

func buildElements(splitted []string) ([]string, int) {
	elements := make([]string, 6)

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

func (format Format) GetPrefix() string { return format.Elements[0] }

func (format Format) GetVersion() string { return format.Elements[1] }

func (format Format) GetCategory() string {
	return format.Elements[2]
}

func (format Format) GetKind() string {
	return format.Elements[3]
}

func (format Format) GetGroup() string {
	return format.Elements[4]
}

func (format Format) GetName() string {
	return format.Elements[5]
}

func (format Format) GetType() string {
	return format.Type
}

func (format Format) GetUUID() uuid.UUID {
	return format.UUID
}

func (format Format) IsValid() bool {
	split := strings.SplitN(format.ToString(), "/", 6)

	if len(split) > 0 {
		for _, element := range split {
			if element == "" {
				return false
			}
		}

		return true
	} else {
		return false
	}
}

func (format Format) Full() bool {
	return format.Elems == 6
}

func (format Format) ToString() string {
	output := ""

	for _, s := range format.Elements {
		if s == "" {
			break
		}

		output += fmt.Sprintf("%s/", s)
	}

	return strings.TrimSuffix(output, "/")
}

func (format Format) ToStringWithUUID() string {
	output := ""

	for _, s := range format.Elements {
		if s == "" {
			break
		}

		output += fmt.Sprintf("%s/", s)
	}

	return fmt.Sprintf("%s%s", format.UUID, strings.TrimSuffix(output, "/"))
}

func (format Format) ToBytes() []byte {
	output := ""

	for _, s := range format.Elements {
		if s == "" {
			break
		}

		output += fmt.Sprintf("%s/", s)
	}

	return []byte(strings.TrimSuffix(output, "/"))
}
