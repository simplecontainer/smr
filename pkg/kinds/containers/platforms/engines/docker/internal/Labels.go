package internal

import (
	"github.com/simplecontainer/smr/pkg/smaps"
)

type Labels struct {
	Labels *smaps.Smap
}

func NewLabels(definition map[string]string) *Labels {
	return &Labels{
		Labels: smaps.NewFromMap(definition),
	}
}

func (labels *Labels) Add(key string, value string) {
	labels.Labels.Add(key, value)
}

func (labels *Labels) ToMap() map[string]string {
	l := make(map[string]string)

	labels.Labels.Map.Range(func(key any, value any) bool {
		l[key.(string)] = value.(string)
		return true
	})

	return l
}
