package helpers

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"beryju.io/saml-test-sp/pkg/helpers"
	"github.com/crewjam/saml/samlidp"
)

func Env(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return fallback
}

// readPEM returns the PEM in value, which is either the PEM itself or a path to a file holding it.
func readPEM(value string) []byte {
	if strings.HasPrefix(strings.TrimSpace(value), "-----BEGIN") {
		return []byte(value)
	}
	data, err := os.ReadFile(value)
	if err != nil {
		panic(err)
	}
	return data
}

// SigningKeypair parses a PEM certificate and key, given inline or as file paths.
func SigningKeypair(certPEM string, keyPEM string) (*rsa.PrivateKey, *x509.Certificate) {
	keypair, err := tls.X509KeyPair(readPEM(certPEM), readPEM(keyPEM))
	if err != nil {
		panic(err)
	}
	cert, err := x509.ParseCertificate(keypair.Certificate[0])
	if err != nil {
		panic(err)
	}
	key, ok := keypair.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		panic("signing key must be an RSA key")
	}
	return key, cert
}

func LoadConfig() samlidp.Options {
	samlOptions := samlidp.Options{
		Logger: log.WithField("component", "saml"),
		Store:  &samlidp.MemoryStore{},
	}

	defaultURL := "http://localhost:9009"
	if _, ok := os.LookupEnv("IDP_SSL_CERT"); ok {
		defaultURL = "https://localhost:9009"
	}
	rootURL := Env("IDP_ROOT_URL", defaultURL)
	url, err := url.Parse(rootURL)
	if err != nil {
		panic(err)
	}
	samlOptions.URL = *url

	priv, pub := helpers.Generate(fmt.Sprintf("localhost,%s", url.Hostname()))
	samlOptions.Key = priv
	samlOptions.Certificate = pub
	if cert, ok := os.LookupEnv("IDP_SIGNING_CERT"); ok {
		samlOptions.Key, samlOptions.Certificate = SigningKeypair(cert, Env("IDP_SIGNING_KEY", ""))
		log.Debug("Signing with the configured keypair")
	} else if sign := Env("IDP_SIGN_REQUESTS", "false"); strings.ToLower(sign) == "true" {
		samlOptions.Key = helpers.LoadRSAKey(os.Getenv("IDP_SSL_KEY"))
		samlOptions.Certificate = helpers.LoadCertificate(os.Getenv("IDP_SSL_CERT"))
		log.Debug("Signing requests")
	}
	log.Debugf("Configuration Optons: %+v", samlOptions)
	return samlOptions
}
