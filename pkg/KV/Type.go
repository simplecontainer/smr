package KV

type KV struct {
	Key    string
	Val    []byte
	Node   uint64
	Local  bool
	Replay bool
}
