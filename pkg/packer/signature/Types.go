package signature

const (
	SignatureMediaType = "application/vnd.simplecontainer.signature.v1+json"
)

type Signer struct {
	PrivateKeyPath string
	SignerName     string
	SignerEmail    string
}

type Signature struct {
	Algorithm string `json:"algorithm"`
	Signature string `json:"signature"` // base64 encoded
	PublicKey string `json:"publicKey"` // base64 encoded
	Signer    struct {
		Name  string `json:"name,omitempty"`
		Email string `json:"email,omitempty"`
	} `json:"signer"`
}
