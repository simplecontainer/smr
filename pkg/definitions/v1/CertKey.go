package v1

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
)

type CertKeyDefinition struct {
	Meta CertKeyMeta `json:"meta" validate:"required"`
	Spec CertKeySpec `json:"spec" validate:"required"`
}

type CertKeyMeta struct {
	Group string `json:"group" validate:"required"`
	Name  string `json:"name" validate:"required"`
}

type CertKeySpec struct {
	Certificate        string `json:"certificate"`
	PublicKey          string `json:"publicKey"`
	PrivateKey         string `json:"privateKey"`
	PrivateKeyPassword string `json:"privateKeyPassword"`
	KeyStore           string `json:"keyStore"`
	KeyStorePassword   string `json:"keyStorePassword"`
	CertStore          string `json:"certStore"`
	CertStorePassword  string `json:"certStorePassword"`
}

func (certkey *CertKeyDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(certkey)
	return string(bytes), err
}

func (certkey *CertKeyDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(certkey)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return false, err
		}
		// from here you can create your own error messages in whatever language you wish
		return false, err
	}

	return true, nil
}
