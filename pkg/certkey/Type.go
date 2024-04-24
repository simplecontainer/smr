package certkey

type CertKey struct {
	Certificate string `json:"certificate"`
	PublicKey   string `json:"publicKey"`
	PrivateKey  string `json:"privateKey"`
}
