package keys

import (
	"bytes"
	"fmt"
	"github.com/simplecontainer/smr/pkg/logger"
	"os"
)

func NewKeys(directory string) *Keys {
	dirCAPrivateKey := fmt.Sprintf("%s/caprivate.key", directory)
	dirCACertpem := fmt.Sprintf("%s/cacert.pem", directory)

	dirServerPrivateKey := fmt.Sprintf("%s/serverprivate.key", directory)
	dirServerCertPem := fmt.Sprintf("%s/servercert.pem", directory)

	dirClientPrivateKey := fmt.Sprintf("%s/clientprivate.key", directory)
	dirClientCertPem := fmt.Sprintf("%s/clientcert.pem", directory)

	caPrivateKey, err := os.ReadFile(dirCAPrivateKey)
	if err != nil {
		caPrivateKey = nil
	}

	serverPrivateKey, err := os.ReadFile(dirServerPrivateKey)
	if err != nil {
		serverPrivateKey = nil
	}

	clientPrivateKey, err := os.ReadFile(dirClientPrivateKey)
	if err != nil {
		clientPrivateKey = nil
	}

	caCertPem, err := os.ReadFile(dirCACertpem)
	if err != nil {
		caCertPem = nil
	}

	serverCertPem, err := os.ReadFile(dirServerCertPem)
	if err != nil {
		serverCertPem = nil
	}

	clientCertPem, err := os.ReadFile(dirClientCertPem)
	if err != nil {
		clientCertPem = nil
	}

	return &Keys{
		CAPrivateKey:         bytes.NewBuffer(caPrivateKey),
		CAPem:                bytes.NewBuffer(caCertPem),
		CAPrivateKeyPath:     dirCAPrivateKey,
		CAPemPath:            dirCACertpem,
		ServerPrivateKey:     bytes.NewBuffer(serverPrivateKey),
		ServerCertPem:        bytes.NewBuffer(serverCertPem),
		ServerPrivateKeyPath: dirServerPrivateKey,
		ServerCertPemPath:    dirServerCertPem,
		ClientPrivateKey:     bytes.NewBuffer(clientPrivateKey),
		ClientCertPem:        bytes.NewBuffer(clientCertPem),
		ClientPrivateKeyPath: dirClientPrivateKey,
		ClientCertPemPath:    dirClientCertPem,
	}
}

func (keys *Keys) GenerateIfNoKeysFound() (bool, error) {
	if keys.ClientPrivateKey.Len() == 0 {
		logger.Log.Info("generating mtls ca, server certificate pem and client certificate pem")
		err := keys.GenerateKeys()

		if err != nil {
			return true, err
		}

		return false, nil
	}

	logger.Log.Info("reading mtls ca, server certificate pem and client certificate pem, since they already exists")
	return true, nil
}

func (keys *Keys) SaveToDirectory() error {
	err := os.WriteFile(keys.CAPrivateKeyPath, keys.CAPrivateKey.Bytes(), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(keys.CAPemPath, keys.CAPem.Bytes(), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(keys.ServerPrivateKeyPath, keys.ServerPrivateKey.Bytes(), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(keys.ServerCertPemPath, keys.ServerCertPem.Bytes(), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(keys.ClientPrivateKeyPath, keys.ClientPrivateKey.Bytes(), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(keys.ClientCertPemPath, keys.ClientCertPem.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
