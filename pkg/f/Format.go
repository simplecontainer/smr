package f

/*
	All database keys should follow the format:
	prefix/version/category/kind/group/identifier/key
*/

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/simplecontainer/smr/pkg/contracts/iformat"
	"github.com/simplecontainer/smr/pkg/static"
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

	elements, nonEmptyCount, err := buildElements(strings.SplitN(f, "/", 7))
	format := Format{
		Elements: elements,
		Elems:    nonEmptyCount,
		UUID:     UUID,
		Type:     TYPE_FORMATED,
	}

	if err == nil && format.IsValid() {
		return format
	} else {
		return Format{}
	}
}

func Build(arg string, group string) (iformat.Format, error) {
	// Build proper format from arg based on info provided
	// Default to prefix=simplecontainer.io, category=kind if missing
	// Group argument is used only for case 2 - ignore in others (it can be set with flag --g)

	var format iformat.Format
	var err error = nil

	split := strings.Split(arg, "/")

	switch len(split) {
	case 1:
		// kind
		format = NewFromString(fmt.Sprintf("%s/%s/%s", static.SMR_PREFIX, "kind", split[0]))
		break
	case 2:
		// kind/name -> read group from flag!
		format = NewFromString(fmt.Sprintf("%s/%s/%s/%s/%s", static.SMR_PREFIX, "kind", split[0], group, split[1]))
		break
	case 3:
		// kind/group/name -> read group from arg
		format = NewFromString(fmt.Sprintf("%s/%s/%s/%s/%s", static.SMR_PREFIX, "kind", split[0], split[1], split[2]))
		break
	case 4:
		// category/kind/group/name
		format = NewFromString(fmt.Sprintf("%s/%s/%s/%s/%s", static.SMR_PREFIX, split[0], split[1], split[2], split[3]))
		break
	case 5:
		// version/category/kind/group/name
		format = NewFromString(fmt.Sprintf("%s/%s/%s/%s/%s", split[0], split[1], split[2], split[3], split[4]))
		break
	case 6:
		// prefix/version/category/kind/group/name
		format = NewFromString(fmt.Sprintf("%s/%s/%s/%s/%s/%s", split[0], split[1], split[2], split[3], split[4], split[5]))
		break
	case 7:
		// prefix/category/kind/group/name/field
		format = NewFromString(fmt.Sprintf("%s/%s/%s/%s/%s/%s/%s", split[0], split[1], split[2], split[3], split[4], split[5], split[6]))
		break
	default:
		err = errors.New("valid formats are: [prefix/category/kind/group/name/field, prefix/category/kind/group/name, category/kind/group/name, kind/group/name, kind/name, kind]")
	}

	if arg == "" {
		err = errors.New("empty input")
	}

	return format, err
}

func buildElements(splitted []string) ([]string, int, error) {
	var size = 7

	elements := make([]string, size)

	nonempty := 0
	for k, v := range splitted {
		if strings.TrimSpace(v) != "" {
			elements[k] = v
			nonempty++
		} else {
			if k+1 <= len(splitted)-1 {
				if elements[k+1] != "" {
					return nil, 0, errors.New("invalid format")
				}
			}

			if k-1 >= 0 && k != len(splitted)-1 {
				if elements[k-1] != "" {
					return nil, 0, errors.New("invalid format")
				}
			}

			elements[k] = ""
		}
	}

	return elements, nonempty, nil
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

func (format Format) GetField() string {
	return format.Elements[6]
}

func (format Format) GetType() string {
	return format.Type
}

func (format Format) Shift() iformat.Format {
	size := len(format.Elements)

	count := 0
	for _, el := range format.Elements {
		if el != "_" {
			count++
		}
	}

	result := make([]string, size)
	for i := range result {
		result[i] = ""
	}

	index := size - 1
	for i := size - 1; i >= 0; i-- {
		if format.Elements[i] != "" {
			result[index] = format.Elements[i]
			index--
		}
	}

	format.Elements = result
	return format
}

func (format Format) GetUUID() uuid.UUID {
	return format.UUID
}

func (format Format) IsValid() bool {
	split := strings.SplitN(format.ToString(), "/", 7)

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

func (format Format) Compliant() bool {
	return format.Elems >= 6 && format.Elems <= 7
}

func (format Format) ToString() string {
	output := ""

	for _, s := range format.Elements {
		if s == "" {
			continue
		}

		output += fmt.Sprintf("%s/", s)
	}

	return strings.TrimSuffix(output, "/")
}

func DefaultToStringOpts() *iformat.ToStringOpts {
	return &iformat.ToStringOpts{}
}

func (format Format) ToStringWithOpts(opts *iformat.ToStringOpts) string {
	var builder strings.Builder

	builder.Grow(64)

	for i, element := range format.Elements {
		if element == "" {
			continue
		}

		if i == 2 && opts.ExcludeCategory {
			continue
		}

		builder.WriteString(element)
		builder.WriteByte('/')
	}

	path := strings.TrimSuffix(builder.String(), "/")

	if opts.IncludeUUID {
		path = format.UUID.String() + path
	}

	if opts.AddPrefixSlash {
		path = "/" + path
	}

	if opts.AddTrailingSlash {
		path = path + "/"
	}

	return path
}

func (format Format) ToStringWithUUID() string {
	output := ""

	for _, s := range format.Elements {
		if s == "" {
			continue
		}

		output += fmt.Sprintf("%s/", s)
	}

	return fmt.Sprintf("%s%s", format.UUID, strings.TrimSuffix(output, "/"))
}

func (format Format) ToBytes() []byte {
	output := ""

	for _, s := range format.Elements {
		if s == "" {
			continue
		}

		output += fmt.Sprintf("%s/", s)
	}

	return []byte(strings.TrimSuffix(output, "/"))
}
