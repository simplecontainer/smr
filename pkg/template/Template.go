package template

import (
	"bytes"
	"html/template"
)

func New(name string, tmpl string, values Variables, functions template.FuncMap) Template {
	return Template{
		Name:      name,
		Templated: tmpl,
		Values:    values,
		Functions: functions,
	}
}

func (tmpl Template) Parse(left string, right string) (string, error) {
	t, err := template.New(tmpl.Name).Funcs(tmpl.Functions).Delims(left, right).Parse(tmpl.Templated)

	if err != nil {
		return tmpl.Templated, err
	}

	var buf bytes.Buffer

	err = t.Execute(&buf, tmpl.Values.Values)
	if err != nil {
		return tmpl.Templated, err
	}

	return buf.String(), nil
}
