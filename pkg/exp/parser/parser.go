package parser

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/sirupsen/logrus"
	"github.com/trufflesecurity/driftwood/pkg/exp/cracker"
)

var (
	ErrNotSupported = errors.New("key type not supported")
	ErrEncryptedKey = errors.New("key is encrypted")
)

func PublicKey(privateKey []byte) ([]byte, error) {
	parsedKey, err := ssh.ParseRawPrivateKey(privateKey)
	if err != nil && strings.Contains(err.Error(), "private key is passphrase protected") {
		logrus.Info("🔒 Encrypted key detected, attempting to crack it")
		parsedKey, err = cracker.Crack(privateKey)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, errors.New(strings.TrimPrefix(err.Error(), "ssh: "))
	}

	var pubKey interface{}
	switch privateKey := parsedKey.(type) {
	case *rsa.PrivateKey:
		pubKey = &privateKey.PublicKey
	case *ecdsa.PrivateKey:
		pubKey = &privateKey.PublicKey
	case ed25519.PrivateKey:
		pubKey = privateKey.Public()
	default:
		return nil, ErrNotSupported
	}

	return x509.MarshalPKIXPublicKey(pubKey)
}
