package contracts

type Event interface {
	GetType() string
	GetTarget() string
	GetKind() string
	GetGroup() string
	GetName() string
	GetData() []byte
}
