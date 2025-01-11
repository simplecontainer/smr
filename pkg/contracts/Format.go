package contracts

type Format interface {
	GetCategory() string
	GetType() string
	ToString() string
	Full() bool
	IsValid() bool
	ToBytes() []byte
}
