package replicas

type Replicas struct {
	Group          string
	GeneratedIndex int
	Replicas       int
	Changed        bool
}
