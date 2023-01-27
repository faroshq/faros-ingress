package utiltls

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func CertAsBytes(certs ...*x509.Certificate) (b []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			b, err = nil, fmt.Errorf("CertAsBytes: %v", r)
		}
	}()

	buf := &bytes.Buffer{}
	for _, cert := range certs {
		err = pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func PrivateKeyAsBytes(key *rsa.PrivateKey) (b []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			b, err = nil, fmt.Errorf("PrivateKeyAsBytes: %v", r)
		}
	}()

	buf := &bytes.Buffer{}

	err = pem.Encode(buf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func CertificatePairFromBytes(certBytes, keyBytes []byte) (cert *x509.Certificate, key *rsa.PrivateKey, err error) {
	cert, err = CertificateFromBytes(certBytes)
	if err != nil {
		return nil, nil, err
	}

	key, err = PrivateKeyFromBytes(keyBytes)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

func CertificateFromBytes(b []byte) (cert *x509.Certificate, err error) {
	cpb, _ := pem.Decode(b)

	cert, err = x509.ParseCertificate(cpb.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func PrivateKeyFromBytes(b []byte) (key *rsa.PrivateKey, err error) {
	kpb, _ := pem.Decode(b)

	key, err = x509.ParsePKCS1PrivateKey(kpb.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}
