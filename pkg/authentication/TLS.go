package authentication

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/static"
	"net"
	"path/filepath"
)

func NewUser(TLSRequest *tls.ConnectionState) *User {
	user := &User{}
	user.ReadTLSFromGinCtx(TLSRequest)

	return user
}

func (user *User) CreateUser(k *keys.Keys, username string, domain string, externalIP string) (string, error) {
	if user.Username == "root" && user.Domain == "localhost" {
		found := k.Exists(static.SMR_SSH_HOME, filepath.Clean(username))
		usernameClean := filepath.Clean(username)

		if found != nil {
			newClient := keys.Client{}
			err := newClient.Generate(
				k.CA,
				[]string{domain, fmt.Sprintf("smr-agent.%s", static.SMR_LOCAL_DOMAIN)},
				[]net.IP{net.ParseIP(externalIP), net.IPv6loopback},
				username,
			)

			if err != nil {
				return "", err
			}

			err = newClient.Write(static.SMR_SSH_HOME, usernameClean)

			if err != nil {
				return "", err
			}

			err = k.GeneratePemBundle(static.SMR_SSH_HOME, usernameClean, &newClient)

			if err != nil {
				return "", err
			}

			k.AppendClient(usernameClean, &newClient)

			return fmt.Sprintf("%s/%s.pem", static.SMR_SSH_HOME, usernameClean), nil
		} else {
			return "", errors.New("credentials for given username already exists")
		}
	} else {
		return "", errors.New("users can only be created by root user from localhost")
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

func (user *User) ToString() string {
	if user != nil {
		str, err := json.Marshal(user)

		if err != nil {
			return ""
		}

		return string(str)
	}

	return ""
}

func (user *User) FromString(str string) error {
	return json.Unmarshal([]byte(str), &user)
}
