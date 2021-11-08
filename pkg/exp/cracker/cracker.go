package cracker

import (
	"bytes"
	"crypto/x509"
	_ "embed"
	"errors"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

//go:embed "top250.txt"
var rawCrackList []byte
var passphrases [][]byte

func init() {
	passphrases = bytes.Split(rawCrackList, []byte("\n"))
}

var (
	ErrUncrackable = errors.New("unable to crack encryption")
)

func Crack(in []byte) (interface{}, error) {
	for _, passphrase := range passphrases {
		parsed, err := ssh.ParseRawPrivateKeyWithPassphrase(in, passphrase)
		if err != nil {
			if errors.Is(err, x509.IncorrectPasswordError) {
				continue
			} else {
				return nil, err
			}
		}
		logrus.Infof("🔓 Cracked key with passphrase: %s", string(passphrase))
		return parsed, nil
	}
	return nil, ErrUncrackable
}
