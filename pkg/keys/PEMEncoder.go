package keys

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"strings"
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

func PEMParse(bundle string) ([]string, error) {
	var blocks []string

	pemData := []byte(bundle)

	for len(pemData) > 0 {
		var block *pem.Block
		block, pemData = pem.Decode(pemData)

		if block == nil {
			if len(strings.TrimSpace(string(pemData))) > 0 {
				return blocks, fmt.Errorf("failed to decode PEM block, invalid data: %s", string(pemData)[:min(len(pemData), 20)])
			}

			break
		}

		var buf bytes.Buffer
		if err := pem.Encode(&buf, block); err != nil {
			return blocks, fmt.Errorf("failed to re-encode PEM block: %w", err)
		}

		blocks = append(blocks, buf.String())
	}

	return blocks, nil
}
