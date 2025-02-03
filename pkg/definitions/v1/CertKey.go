package v1

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type CertKeyDefinition struct {
	Kind   string          `json:"kind" validate:"required"`
	Prefix string          `json:"prefix" validate:"required"`
	Meta   commonv1.Meta   `json:"meta" validate:"required"`
	Spec   CertKeySpec     `json:"spec" validate:"required"`
	State  *commonv1.State `json:"state"`
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

func (certkey *CertKeyDefinition) SetRuntime(runtime *commonv1.Runtime) {
	certkey.Meta.Runtime = runtime
}

func (certkey *CertKeyDefinition) GetRuntime() *commonv1.Runtime {
	return certkey.Meta.Runtime
}

func (certkey *CertKeyDefinition) GetPrefix() string {
	return certkey.Prefix
}

func (certkey *CertKeyDefinition) GetMeta() commonv1.Meta {
	return certkey.Meta
}

func (certkey *CertKeyDefinition) GetState() *commonv1.State {
	return certkey.State
}

func (certkey *CertKeyDefinition) SetState(state *commonv1.State) {
	certkey.State = state
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
