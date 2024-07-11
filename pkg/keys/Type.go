package keys

import (
	"bytes"
)

type Keys struct {
	CAPrivateKey *bytes.Buffer
	CAPem        *bytes.Buffer

	CAPrivateKeyPath string
	CAPemPath        string

	ServerPrivateKey *bytes.Buffer
	ServerCertPem    *bytes.Buffer

	ServerPrivateKeyPath string
	ServerCertPemPath    string

	ClientPrivateKey *bytes.Buffer
	ClientCertPem    *bytes.Buffer

	ClientPrivateKeyPath string
	ClientCertPemPath    string

	SerialNumber int64
}
