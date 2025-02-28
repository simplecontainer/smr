package keys

import (
	"bytes"
	"encoding/pem"
)

const PRIVATE_KEY = "EC PRIVATE KEY"
const CERTIFICATE = "CERTIFICATE"

func PEMEncode(FileType string, ContentBytes []byte) ([]byte, error) {
	PEMEncoded := new(bytes.Buffer)

	err := pem.Encode(PEMEncoded, &pem.Block{
		Type:  FileType,
		Bytes: ContentBytes,
	})

	if err != nil {
		return nil, err
	}

	return PEMEncoded.Bytes(), nil
}

func PEMDecode(ContentBytes []byte) []byte {
	block, _ := pem.Decode(ContentBytes)
	return block.Bytes
}
