package template

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type FunctionManager struct {
	functions map[string]interface{}
}

func NewFunctionManager() *FunctionManager {
	fm := &FunctionManager{
		functions: make(map[string]interface{}),
	}

	// Load default functions
	fm.LoadDefaultFunctions()

	return fm
}

func NewEmptyFunctionManager() *FunctionManager {
	return &FunctionManager{
		functions: make(map[string]interface{}),
	}
}

func (fm *FunctionManager) AddFunction(name string, fn interface{}) error {
	if err := fm.validateFunction(fn); err != nil {
		return fmt.Errorf("invalid function %s: %w", name, err)
	}

	fm.functions[name] = fn
	return nil
}

func (fm *FunctionManager) AddFunctions(functions map[string]interface{}) error {
	for name, fn := range functions {
		if err := fm.AddFunction(name, fn); err != nil {
			return err
		}
	}
	return nil
}

func (fm *FunctionManager) RemoveFunction(name string) {
	delete(fm.functions, name)
}

func (fm *FunctionManager) RemoveFunctions(names ...string) {
	for _, name := range names {
		fm.RemoveFunction(name)
	}
}

func (fm *FunctionManager) GetFunction(name string) (interface{}, bool) {
	fn, exists := fm.functions[name]
	return fn, exists
}

func (fm *FunctionManager) HasFunction(name string) bool {
	_, exists := fm.functions[name]
	return exists
}

func (fm *FunctionManager) ListFunctions() []string {
	var names []string
	for name := range fm.functions {
		names = append(names, name)
	}
	return names
}

func (fm *FunctionManager) Count() int {
	return len(fm.functions)
}

func (fm *FunctionManager) Clear() {
	fm.functions = make(map[string]interface{})
}

func (fm *FunctionManager) Clone() *FunctionManager {
	clone := &FunctionManager{
		functions: make(map[string]interface{}),
	}

	// Copy functions
	for name, fn := range fm.functions {
		clone.functions[name] = fn
	}

	return clone
}

func (fm *FunctionManager) FuncMap() template.FuncMap {
	funcMap := make(template.FuncMap)
	for name, fn := range fm.functions {
		funcMap[name] = fn
	}
	return funcMap
}

func (fm *FunctionManager) validateFunction(fn interface{}) error {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return fmt.Errorf("not a function")
	}
	return nil
}

func (fm *FunctionManager) LoadDefaultFunctions() {
	// String functions
	fm.functions["upper"] = strings.ToUpper
	fm.functions["lower"] = strings.ToLower
	fm.functions["title"] = strings.Title
	fm.functions["trim"] = strings.TrimSpace
	fm.functions["trimSuffix"] = strings.TrimSuffix
	fm.functions["trimPrefix"] = strings.TrimPrefix
	fm.functions["contains"] = strings.Contains
	fm.functions["hasPrefix"] = strings.HasPrefix
	fm.functions["hasSuffix"] = strings.HasSuffix
	fm.functions["split"] = func(sep, s string) []string {
		return strings.Split(s, sep)
	}
	fm.functions["join"] = func(sep string, elems []string) string {
		return strings.Join(elems, sep)
	}
	fm.functions["replace"] = func(old, new string, n int, s string) string {
		return strings.Replace(s, old, new, n)
	}
	fm.functions["replaceAll"] = func(old, new, s string) string {
		return strings.ReplaceAll(s, old, new)
	}
	fm.functions["quote"] = func(s string) string {
		return fmt.Sprintf("%q", s)
	}
	fm.functions["squote"] = func(s string) string {
		return fmt.Sprintf("'%s'", s)
	}
	fm.functions["substr"] = func(start, length int, s string) string {
		if start < 0 || start >= len(s) {
			return ""
		}
		end := start + length
		if end > len(s) {
			end = len(s)
		}
		return s[start:end]
	}
	fm.functions["indent"] = func(spaces int, s string) string {
		pad := strings.Repeat(" ", spaces)
		return pad + strings.Replace(s, "\n", "\n"+pad, -1)
	}
	fm.functions["nindent"] = func(spaces int, s string) string {
		pad := strings.Repeat(" ", spaces)
		return "\n" + pad + strings.Replace(s, "\n", "\n"+pad, -1)
	}

	// Type conversion functions
	fm.functions["toString"] = func(v interface{}) string {
		return fmt.Sprintf("%v", v)
	}
	fm.functions["toStrings"] = func(list []interface{}) []string {
		result := make([]string, len(list))
		for i, v := range list {
			result[i] = fmt.Sprintf("%v", v)
		}
		return result
	}
	fm.functions["toInt"] = func(v interface{}) (int, error) {
		switch val := v.(type) {
		case int:
			return val, nil
		case string:
			return strconv.Atoi(val)
		case float64:
			return int(val), nil
		default:
			return 0, fmt.Errorf("cannot convert %T to int", v)
		}
	}
	fm.functions["toInt64"] = func(v interface{}) (int64, error) {
		switch val := v.(type) {
		case int64:
			return val, nil
		case int:
			return int64(val), nil
		case string:
			return strconv.ParseInt(val, 10, 64)
		case float64:
			return int64(val), nil
		default:
			return 0, fmt.Errorf("cannot convert %T to int64", v)
		}
	}
	fm.functions["toFloat64"] = func(v interface{}) (float64, error) {
		switch val := v.(type) {
		case float64:
			return val, nil
		case int:
			return float64(val), nil
		case string:
			return strconv.ParseFloat(val, 64)
		default:
			return 0, fmt.Errorf("cannot convert %T to float64", v)
		}
	}

	// JSON functions
	fm.functions["toJson"] = func(v interface{}) (string, error) {
		data, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	fm.functions["toPrettyJson"] = func(v interface{}) (string, error) {
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	fm.functions["fromJson"] = func(s string) (interface{}, error) {
		var result interface{}
		err := json.Unmarshal([]byte(s), &result)
		return result, err
	}

	// List functions
	fm.functions["list"] = func(items ...interface{}) []interface{} {
		return items
	}
	fm.functions["append"] = func(list []interface{}, item interface{}) []interface{} {
		return append(list, item)
	}
	fm.functions["prepend"] = func(list []interface{}, item interface{}) []interface{} {
		return append([]interface{}{item}, list...)
	}
	fm.functions["first"] = func(list []interface{}) interface{} {
		if len(list) == 0 {
			return nil
		}
		return list[0]
	}
	fm.functions["last"] = func(list []interface{}) interface{} {
		if len(list) == 0 {
			return nil
		}
		return list[len(list)-1]
	}
	fm.functions["rest"] = func(list []interface{}) []interface{} {
		if len(list) <= 1 {
			return []interface{}{}
		}
		return list[1:]
	}
	fm.functions["initial"] = func(list []interface{}) []interface{} {
		if len(list) <= 1 {
			return []interface{}{}
		}
		return list[:len(list)-1]
	}
	fm.functions["reverse"] = func(list []interface{}) []interface{} {
		result := make([]interface{}, len(list))
		for i, v := range list {
			result[len(list)-1-i] = v
		}
		return result
	}
	fm.functions["uniq"] = func(list []interface{}) []interface{} {
		seen := make(map[interface{}]bool)
		var result []interface{}
		for _, item := range list {
			if !seen[item] {
				seen[item] = true
				result = append(result, item)
			}
		}
		return result
	}
	fm.functions["compact"] = func(list []interface{}) []interface{} {
		var result []interface{}
		for _, item := range list {
			if item != nil && item != "" {
				result = append(result, item)
			}
		}
		return result
	}

	// Dictionary/Map functions
	fm.functions["dict"] = func(pairs ...interface{}) (map[string]interface{}, error) {
		if len(pairs)%2 != 0 {
			return nil, fmt.Errorf("dict requires an even number of arguments")
		}
		result := make(map[string]interface{})
		for i := 0; i < len(pairs); i += 2 {
			key, ok := pairs[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			result[key] = pairs[i+1]
		}
		return result, nil
	}
	fm.functions["get"] = func(dict map[string]interface{}, key string) interface{} {
		return dict[key]
	}
	fm.functions["set"] = func(dict map[string]interface{}, key string, value interface{}) map[string]interface{} {
		result := make(map[string]interface{})
		for k, v := range dict {
			result[k] = v
		}
		result[key] = value
		return result
	}
	fm.functions["unset"] = func(dict map[string]interface{}, key string) map[string]interface{} {
		result := make(map[string]interface{})
		for k, v := range dict {
			if k != key {
				result[k] = v
			}
		}
		return result
	}
	fm.functions["hasKey"] = func(dict map[string]interface{}, key string) bool {
		_, exists := dict[key]
		return exists
	}
	fm.functions["keys"] = func(dict map[string]interface{}) []string {
		keys := make([]string, 0, len(dict))
		for k := range dict {
			keys = append(keys, k)
		}
		return keys
	}
	fm.functions["values"] = func(dict map[string]interface{}) []interface{} {
		values := make([]interface{}, 0, len(dict))
		for _, v := range dict {
			values = append(values, v)
		}
		return values
	}

	// Math functions
	fm.functions["add"] = func(a, b int) int { return a + b }
	fm.functions["sub"] = func(a, b int) int { return a - b }
	fm.functions["mul"] = func(a, b int) int { return a * b }
	fm.functions["div"] = func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a / b
	}
	fm.functions["mod"] = func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a % b
	}
	fm.functions["max"] = func(a, b int) int {
		if a > b {
			return a
		}
		return b
	}
	fm.functions["min"] = func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}

	// Date functions
	fm.functions["now"] = func() time.Time {
		return time.Now()
	}
	fm.functions["date"] = func(format string, t time.Time) string {
		return t.Format(format)
	}
	fm.functions["dateInZone"] = func(format, zone string, t time.Time) (string, error) {
		loc, err := time.LoadLocation(zone)
		if err != nil {
			return "", err
		}
		return t.In(loc).Format(format), nil
	}
	fm.functions["unixEpoch"] = func(t time.Time) int64 {
		return t.Unix()
	}

	// Path functions
	fm.functions["base"] = filepath.Base
	fm.functions["dir"] = filepath.Dir
	fm.functions["ext"] = filepath.Ext
	fm.functions["clean"] = filepath.Clean
	fm.functions["isAbs"] = filepath.IsAbs

	// Conditional functions
	fm.functions["default"] = func(defaultValue, value interface{}) interface{} {
		if value == nil || value == "" {
			return defaultValue
		}
		return value
	}
	fm.functions["empty"] = func(value interface{}) bool {
		if value == nil {
			return true
		}
		switch v := value.(type) {
		case string:
			return v == ""
		case []interface{}:
			return len(v) == 0
		case map[string]interface{}:
			return len(v) == 0
		default:
			return false
		}
	}
	fm.functions["coalesce"] = func(values ...interface{}) interface{} {
		for _, value := range values {
			if value != nil && value != "" {
				return value
			}
		}
		return nil
	}
	fm.functions["ternary"] = func(condition bool, trueValue, falseValue interface{}) interface{} {
		if condition {
			return trueValue
		}
		return falseValue
	}

	// Regular expression functions
	fm.functions["regexMatch"] = func(pattern, s string) (bool, error) {
		return regexp.MatchString(pattern, s)
	}
	fm.functions["regexFindAll"] = func(pattern, s string, n int) ([]string, error) {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		return re.FindAllString(s, n), nil
	}
	fm.functions["regexReplaceAll"] = func(pattern, repl, s string) (string, error) {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return "", err
		}
		return re.ReplaceAllString(s, repl), nil
	}
	fm.functions["regexSplit"] = func(pattern, s string, n int) ([]string, error) {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		return re.Split(s, n), nil
	}

	// Encoding functions
	fm.functions["base64encode"] = func(input string) string {
		return base64.StdEncoding.EncodeToString([]byte(input))
	}
	fm.functions["base64decode"] = func(input string) (string, error) {
		decoded, err := base64.StdEncoding.DecodeString(input)
		if err != nil {
			return input, err
		}
		return string(decoded), nil
	}
	fm.functions["b64enc"] = func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}
	fm.functions["b64dec"] = func(s string) (string, error) {
		decoded, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	}
	fm.functions["urlquery"] = func(s string) string {
		return strings.ReplaceAll(s, " ", "%20")
	}
}

type Builder struct {
	fm *FunctionManager
}

func NewBuilder() *Builder {
	return &Builder{
		fm: NewFunctionManager(),
	}
}

func NewBuilderEmpty() *Builder {
	return &Builder{
		fm: NewEmptyFunctionManager(),
	}
}

func (b *Builder) WithFunction(name string, fn interface{}) *Builder {
	b.fm.AddFunction(name, fn)
	return b
}

func (b *Builder) WithFunctions(functions map[string]interface{}) *Builder {
	b.fm.AddFunctions(functions)
	return b
}

func (b *Builder) WithoutFunction(name string) *Builder {
	b.fm.RemoveFunction(name)
	return b
}

func (b *Builder) WithoutFunctions(names ...string) *Builder {
	b.fm.RemoveFunctions(names...)
	return b
}

func (b *Builder) Build() *FunctionManager {
	return b.fm
}
