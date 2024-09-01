package secure

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

const validTime = 5 * 365 * 24 * time.Hour
const organization = "Fada Control"

func GenerateX509Cert() (certPEM, keyPEM []byte, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{organization},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(validTime),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return certPEM, keyPEM, nil
}
func LoadBaseX509KeyPair(certPEM, keyPEM string) (tls.Certificate, error) {
	cert, err := base64.StdEncoding.DecodeString(certPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	key, err := base64.StdEncoding.DecodeString(keyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}
	return LoadX509KeyPairFromMemory(cert, key)
}
func LoadX509KeyPairFromMemory(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) {

	certBlock, _ := pem.Decode(certPEMBlock)
	if certBlock == nil {
		return tls.Certificate{}, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to parse certificate: %v", err)
	}

	keyBlock, _ := pem.Decode(keyPEMBlock)
	if keyBlock == nil {
		return tls.Certificate{}, fmt.Errorf("failed to parse private key PEM")
	}

	privKey, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to parse private key: %v", err)
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  privKey,
		Leaf:        cert,
	}

	return tlsCert, nil
}
