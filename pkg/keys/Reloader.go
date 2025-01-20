package keys

import (
	"crypto/tls"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"os"
	"sync"
)

type keypairReloader struct {
	ReloadC  chan os.Signal
	certMu   sync.RWMutex
	cert     *tls.Certificate
	certPath string
	keyPath  string
}

func NewKeypairReloader(certPath, keyPath string) (*keypairReloader, error) {
	result := &keypairReloader{
		certPath: certPath,
		keyPath:  keyPath,
		ReloadC:  make(chan os.Signal, 1),
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	result.cert = &cert
	go func() {
		for {
			select {
			case <-result.ReloadC:

				if err = result.maybeReload(); err != nil {
					logger.Log.Error("Keeping old TLS certificate because the new one could not be loaded: %v", zap.Error(err))
				}

				logger.Log.Info("Reloaded TLS certificates")
			}
		}
	}()
	return result, nil
}

func (kpr *keypairReloader) maybeReload() error {
	newCert, err := tls.LoadX509KeyPair(kpr.certPath, kpr.keyPath)
	if err != nil {
		return err
	}
	kpr.certMu.Lock()
	defer kpr.certMu.Unlock()
	kpr.cert = &newCert
	return nil
}

func (kpr *keypairReloader) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		kpr.certMu.RLock()
		defer kpr.certMu.RUnlock()
		return kpr.cert, nil
	}
}
