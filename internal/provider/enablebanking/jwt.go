package enablebanking

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"
)

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("enablebanking private key is not PEM")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("enablebanking private key is not RSA")
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("enablebanking unsupported private key type: %s", block.Type)
	}
}

func createJWT(appID string, key *rsa.PrivateKey, now time.Time) (string, error) {
	header, err := json.Marshal(struct {
		Alg string `json:"alg"`
		Typ string `json:"typ,omitempty"`
		Kid string `json:"kid"`
	}{
		Alg: "RS256",
		Typ: "JWT",
		Kid: appID,
	})
	if err != nil {
		return "", err
	}
	iat := now.Unix()
	claims, err := json.Marshal(struct {
		Iss string `json:"iss"`
		Aud string `json:"aud"`
		Iat int64  `json:"iat"`
		Exp int64  `json:"exp"`
	}{
		Iss: "enablebanking.com",
		Aud: "api.enablebanking.com",
		Iat: iat,
		Exp: iat + 3600,
	})
	if err != nil {
		return "", err
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(claims)
	digest := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}
