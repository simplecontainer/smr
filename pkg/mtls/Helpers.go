package mtls

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/keys"
	"strings"
)

func GeneratePemBundle(keys *keys.Keys) string {
	return fmt.Sprintf("%s\n%s\n%s\n", strings.TrimSpace(keys.ClientPrivateKey.String()), strings.TrimSpace(keys.ClientCertPem.String()), strings.TrimSpace(keys.CAPem.String()))
}
