package f

/*
	All database keys should follow the format:
	category/kind/group/identifier/key

	eg:


*/

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/simplecontainer/smr/pkg/contracts"
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
	UUID, f := parseUUID(strings.TrimPrefix(data, "/"))

	elements, nonEmptyCount := BuildElements(strings.SplitN(f, "/", 5))
	format := Format{
		Prefix:     strings.TrimSpace(elements[0]),
		Category:   strings.TrimSpace(elements[1]),
		Kind:       strings.TrimSpace(elements[2]),
		Group:      strings.TrimSpace(elements[3]),
		Identifier: strings.TrimSpace(elements[4]),
		Elems:      nonEmptyCount,
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
		fmt.Println("No data")
		UUID := uuid.New()

		//Format didn't start with UUID so return new UUID and f as it was since it only had data
		return UUID, f
	}
}

func BuildElements(splitted []string) ([]string, int) {
	elements := make([]string, 5)

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
	return format.Prefix
}

func (format Format) GetType() string {
	return format.Type
}

func (format Format) GetUUID() uuid.UUID {
	return format.UUID
}

func (format Format) IsValid() bool {
	split := strings.SplitN(strings.TrimPrefix(format.ToString(), "/"), "/", 5)

	for _, element := range split {
		if element == "" {
			return false
		}
	}

	return true
}

func (format Format) Full() bool {
	return format.Elems == 5
}

func (format Format) WithPrefix(prefix string) contracts.Format {
	format.Prefix = prefix
	return format
}

func (format Format) ToString() string {
	output := fmt.Sprintf("/%s/%s/%s/%s/%s", format.Prefix, format.Category, format.Kind, format.Group, format.Identifier)
	return output
}

func (format Format) ToStringWithUUID() string {
	output := fmt.Sprintf("%s/%s/%s/%s/%s", format.Prefix, format.Category, format.Kind, format.Group, format.Identifier)
	return fmt.Sprintf("%s%s", format.UUID, output)
}

func (format Format) ToBytes() []byte {
	output := fmt.Sprintf("/%s/%s/%s/%s/%s", format.Prefix, format.Category, format.Kind, format.Group, format.Identifier)
	return []byte(output)
}
