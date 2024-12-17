package replicas

func (sr *ScopedReplicas) Add(group string, name string) {
	sr.Create = append(sr.Create, R{
		Group: group,
		Name:  name,
	})
}
func (sr *ScopedReplicas) AddExisting(group string, name string) {
	sr.Existing = append(sr.Existing, R{
		Group: group,
		Name:  name,
	})
}
func (sr *ScopedReplicas) Delete(group string, name string) {
	sr.Remove = append(sr.Remove, R{
		Group: group,
		Name:  name,
	})
}
