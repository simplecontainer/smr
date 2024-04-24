package definitions

type CertKey struct {
	Meta CertKeyMeta `mapstructure:"meta"`
	Spec CertKeySpec `mapstructure:"spec"`
}

type CertKeyMeta struct {
	Group      string `json:"group"`
	Identifier string `json:"identifier"`
}

type CertKeySpec struct {
	Certificate string `json:"certificate"`
	PublicKey   string `json:"publicKey"`
	PrivateKey  string `json:"privateKey"`
}
