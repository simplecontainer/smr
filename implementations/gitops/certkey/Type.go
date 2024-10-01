package certkey

import v1 "github.com/simplecontainer/smr/pkg/definitions/v1"

type CertKey struct {
	Certificate        string `json:"certificate"`
	PublicKey          string `json:"publicKey"`
	PrivateKey         string `json:"privateKey"`
	PrivateKeyPassword string `json:"privateKeyPassword"`
	KeyStore           string `json:"keystore"`
	KeyStorePassword   string `json:"keyStorePassword"`
	CertStore          string `json:"certstore"`
	CertStorePassword  string `json:"certstorePassword"`
	Definition         v1.CertKeyDefinition
}
