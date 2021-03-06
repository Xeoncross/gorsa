package gorsa

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

// LoadPrivateKey from a PEM encoded private (or public) key
func LoadPrivateKey(pembytes []byte, password string) (key rsa.PrivateKey, err error) {
	pembytes = bytes.TrimSpace(pembytes)

	var block *pem.Block
	block, _ = pem.Decode(pembytes)
	if block == nil {
		err = errors.New("Invalid PEM key file")
		return
	}

	// Often needed for encrypted keys (i.e. SSH keys)
	if x509.IsEncryptedPEMBlock(block) {
		block, err = DecryptPEMBlock(block, password)
		if err != nil {
			err = errors.New("Error decrypting PEM block: " + err.Error())
			return
		}
	}

	// "BEGIN RSA (PUBLIC|PRIVATE) KEY" is PKCS#1, which can only contain RSA keys.
	// "BEGIN (PUBLIC|PRIVATE) KEY" is PKCS#8, which can contain a variety of formats.
	// "BEGIN ENCRYPTED PRIVATE KEY" is encrypted PKCS#8.

	switch block.Type {
	case "PRIVATE KEY", "ENCRYPTED PRIVATE KEY", "RSA PRIVATE KEY":

		// PEM keys could be PKCS #1-#15 or another type
		var generalKey interface{}
		generalKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			generalKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				// Must be an older system that made this key...
				var x rsa.PrivateKey
				_, err = asn1.Unmarshal(block.Bytes, &x)

				// uncle
				if err != nil {
					return
				}

				generalKey = &x
			}
		}

		// We only support RSA keys
		switch k := generalKey.(type) {
		case *rsa.PrivateKey:
			key = *k
		default:
			err = fmt.Errorf("Unsupported key type %T", generalKey)
		}
	default:
		err = fmt.Errorf("Unsupported PEM block type %q", block.Type)
	}

	return
}

// LoadPrivateKeyFromFile given (expecting PEM format)
func LoadPrivateKeyFromFile(filename string, password string) (key rsa.PrivateKey, err error) {
	var b []byte
	b, err = ioutil.ReadFile(filename)

	b = bytes.TrimSpace(b)

	return LoadPrivateKey(b, password)
}

// LoadPublicKey from a PEM encoded private (or public) key
func LoadPublicKey(pembytes []byte, password string) (pubkey rsa.PublicKey, err error) {
	pembytes = bytes.TrimSpace(pembytes)

	var block *pem.Block
	block, _ = pem.Decode(pembytes)
	if block == nil {
		err = errors.New("Invalid PEM key file")
		return
	}

	// Often needed for encrypted keys (i.e. SSH keys)
	if x509.IsEncryptedPEMBlock(block) {
		block, err = DecryptPEMBlock(block, password)
		if err != nil {
			err = errors.New("Error decrypting PEM block: " + err.Error())
			return
		}
	}

	// "BEGIN RSA (PUBLIC|PRIVATE) KEY" is PKCS#1, which can only contain RSA keys.
	// "BEGIN (PUBLIC|PRIVATE) KEY" is PKCS#8, which can contain a variety of formats.
	// "BEGIN ENCRYPTED PRIVATE KEY" is encrypted PKCS#8.

	switch block.Type {
	case "PUBLIC KEY", "PRIVATE KEY",
		"ENCRYPTED PRIVATE KEY",
		"RSA PUBLIC KEY", "RSA PRIVATE KEY":

		// PEM keys could be PKCS #1-#15, PKIX, elliptic curve, or another type
		var generalKey interface{}
		generalKey, err = x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			generalKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				generalKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
				if err != nil {
					// Must be an older system that made this key...
					var x rsa.PublicKey
					_, err = asn1.Unmarshal(block.Bytes, &x)

					// uncle
					if err != nil {
						return
					}

					generalKey = &x
				}
			}
		}

		// We only support RSA keys
		switch k := generalKey.(type) {
		case *rsa.PublicKey:
			pubkey = *k
		case *rsa.PrivateKey:
			pubkey = k.PublicKey
		// This also works with DSA and ECDSA
		// case *dsa.PublicKey:
		// case *ecdsa.PublicKey:
		default:
			err = fmt.Errorf("Unsupported key type %T", generalKey)
		}
	default:
		err = fmt.Errorf("Unsupported PEM block type %q", block.Type)
	}

	return
}

// LoadPublicKeyFromFile given (expecting PEM format)
func LoadPublicKeyFromFile(filename string, password string) (pubkey rsa.PublicKey, err error) {
	var b []byte
	b, err = ioutil.ReadFile(filename)

	b = bytes.TrimSpace(b)

	return LoadPublicKey(b, password)
}
