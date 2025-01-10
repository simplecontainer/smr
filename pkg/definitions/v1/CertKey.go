package v1

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type CertKeyDefinition struct {
	Meta CertKeyMeta `json:"meta" validate:"required"`
	Spec CertKeySpec `json:"spec" validate:"required"`
}

type CertKeyMeta struct {
	Group string         `json:"group" validate:"required"`
	Name  string         `json:"name" validate:"required"`
	Owner commonv1.Owner `json:"-"`
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

func (certkey *CertKeyDefinition) SetOwner(kind string, group string, name string) {
	certkey.Meta.Owner.Kind = kind
	certkey.Meta.Owner.Group = group
	certkey.Meta.Owner.Name = name
}

func (certkey *CertKeyDefinition) GetOwner() commonv1.Owner {
	return certkey.Meta.Owner
}

func (certkey *CertKeyDefinition) GetKind() string {
	return static.KIND_CERTKEY
}

func (certkey *CertKeyDefinition) ResolveReferences(obj contracts.ObjectInterface) ([]contracts.IDefinition, error) {
	return nil, nil
}

func (certkey *CertKeyDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, certkey)
}

func (certkey *CertKeyDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(certkey)
	return bytes, err
}

func (certkey *CertKeyDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(certkey)
	return string(bytes), err
}

func (certkey *CertKeyDefinition) ToJsonStringWithKind() (string, error) {
	bytes, err := json.Marshal(certkey)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return "", err
	}

	definition["kind"] = "certkey"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return "", err
	}

	return string(marshalled), err
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
