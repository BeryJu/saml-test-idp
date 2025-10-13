package helpers

import (
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
	if sign := Env("IDP_SIGN_REQUESTS", "false"); strings.ToLower(sign) == "true" {
		samlOptions.Key = helpers.LoadRSAKey(os.Getenv("IDP_SSL_KEY"))
		samlOptions.Certificate = helpers.LoadCertificate(os.Getenv("IDP_SSL_CERT"))
		log.Debug("Signing requests")
	}
	log.Debugf("Configuration Optons: %+v", samlOptions)
	return samlOptions
}
