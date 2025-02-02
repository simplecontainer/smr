package template

import "html/template"

type Template struct {
	Name      string
	Templated string
	Values    Variables
	Functions template.FuncMap
}

type Variables struct {
	Values map[string]interface{}
}
