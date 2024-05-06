package certkey

type CertKey struct {
	Certificate        string `json:"certificate"`
	PublicKey          string `json:"publicKey"`
	PrivateKey         string `json:"privateKey"`
	PrivateKeyPassword string `json:"privateKeyPassword"`
	KeyStore           string `json:"keystore"`
	KeyStorePassword   string `json:"keyStorePassword"`
	CertStore          string `json:"certstore"`
	CertStorePassword  string `json:"certstorePassword"`
}
