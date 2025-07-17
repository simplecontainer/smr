package commonv1

type Meta struct {
	Group   string            `json:"group" validate:"required"`
	Name    string            `json:"name" validate:"required"`
	Labels  map[string]string `json:"labels,omitempty"`
	Runtime *Runtime          `json:"runtime,omitempty"`
}

func (m *Meta) GetGroup() string {
	return m.Group
}

func (m *Meta) SetGroup(group string) {
	m.Group = group
}

func (m *Meta) GetName() string {
	return m.Name
}

func (m *Meta) SetName(name string) {
	m.Name = name
}

func (m *Meta) GetLabels() map[string]string {
	return m.Labels
}

func (m *Meta) SetLabels(labels map[string]string) {
	m.Labels = labels
}

func (m *Meta) GetLabel(key string) (string, bool) {
	if m.Labels == nil {
		return "", false
	}
	value, exists := m.Labels[key]
	return value, exists
}

func (m *Meta) SetLabel(key, value string) {
	if m.Labels == nil {
		m.Labels = make(map[string]string)
	}
	m.Labels[key] = value
}

func (m *Meta) DeleteLabel(key string) {
	if m.Labels != nil {
		delete(m.Labels, key)
	}
}

func (m *Meta) GetRuntime() *Runtime {
	return m.Runtime
}

func (m *Meta) SetRuntime(runtime *Runtime) {
	m.Runtime = runtime
}
