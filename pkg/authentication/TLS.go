package authentication

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/static"
	"path/filepath"
)

func NewUser(TLSRequest *tls.ConnectionState) *User {
	user := &User{}
	user.ReadTLSFromGinCtx(TLSRequest)

	return user
}

func (user *User) CreateUser(k *keys.Keys, agent string, username string, domain string, externalIP string) (string, error) {
	if user.Username == agent && user.Domain == "localhost" {
		exists := k.ClientExists(static.SMR_SSH_HOME, filepath.Clean(username))
		usernameClean := filepath.Clean(username)

		if exists == nil {
			client := keys.NewClient()

			err := client.Generate(
				k.CA,
				configuration.NewDomains([]string{domain, fmt.Sprintf("%s.%s", static.SMR_ENDPOINT_NAME, static.SMR_LOCAL_DOMAIN)}),
				configuration.NewIPs([]string{externalIP}),
				username,
			)

			if err != nil {
				return "", err
			}

			err = client.Write(static.SMR_SSH_HOME, usernameClean)

			if err != nil {
				return "", err
			}

			err = k.GeneratePemBundle(static.SMR_SSH_HOME, usernameClean, client)

			if err != nil {
				return "", err
			}

			k.AppendClient(usernameClean, client)

			return fmt.Sprintf("%s/%s.pem", static.SMR_SSH_HOME, usernameClean), nil
		} else {
			return "", exists
		}
	} else {
		return "", errors.New("users can only be created by agent user from localhost")
	}
}

func (user *User) ReadTLSFromGinCtx(TLSRequest *tls.ConnectionState) {
	if TLSRequest != nil {
		user.Domain = TLSRequest.ServerName

		if TLSRequest.PeerCertificates[0] != nil {
			user.Username = TLSRequest.PeerCertificates[0].Subject.CommonName
		}
	}
}
