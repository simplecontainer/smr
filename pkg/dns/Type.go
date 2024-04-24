package dns

type Records struct {
	ARecords map[string]ARecord
}

type ARecord struct {
	Domain map[string][]string
}
