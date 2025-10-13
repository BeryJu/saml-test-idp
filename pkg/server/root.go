package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"beryju.io/saml-test-idp/pkg/helpers"
	"github.com/crewjam/saml/samlidp"
	"github.com/crewjam/saml/samlsp"
)

type Server struct {
	idp *samlidp.Server
	h   *http.ServeMux
	l   *log.Entry
	b   string
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	_, _ = fmt.Fprint(w, "hello :)")
}

func (s *Server) logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.l.WithField("remoteAddr", r.RemoteAddr).WithField("method", r.Method).Info(r.URL)
		handler.ServeHTTP(w, r)
	})
}

func mustBcrypt(pw string) []byte {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return h
}

func RunServer() {
	config := helpers.LoadConfig()
	err := config.Store.Put("/users/user1", samlidp.User{
		Name:           "user1",
		HashedPassword: mustBcrypt("user1pass"),
		Groups:         []string{"Administrators", "Users"},
		Email:          "user1@example.com",
		CommonName:     "Alice Smith",
		Surname:        "Smith",
		GivenName:      "Alice",
	})
	if err != nil {
		panic(err)
	}

	err = config.Store.Put("/users/user2", samlidp.User{
		Name:           "user2",
		HashedPassword: mustBcrypt("user2pass"),
		Groups:         []string{"Users"},
		Email:          "user2@example.com",
		CommonName:     "user2 Smith",
		Surname:        "Smith",
		GivenName:      "Bob",
	})
	if err != nil {
		panic(err)
	}

	metadata := helpers.Env("IDP_METADATA_URL", "")
	var svc samlidp.Service
	if metadata == "" {
		panic("Metadata required")
	}
	u, err := url.Parse(metadata)
	if err != nil {
		panic(err)
	}
	desc, err := samlsp.FetchMetadata(context.TODO(), http.DefaultClient, *u)
	if err != nil {
		panic(err)
	}
	svc = samlidp.Service{
		Name:     "test-app",
		Metadata: *desc,
	}

	err = config.Store.Put("/services/test-app", svc)
	if err != nil {
		panic(err)
	}

	idp, err := samlidp.New(config)
	if err != nil {
		panic(err)
	}
	// https://github.com/crewjam/saml/issues/613
	idp.IDP.LoginURL = idp.IDP.SSOURL
	server := Server{
		idp: idp,
		h:   http.NewServeMux(),
		l:   log.WithField("component", "server"),
	}
	server.h.HandleFunc("/health", server.health)
	server.h.Handle("/", server.idp)

	listen := helpers.Env("IDP_BIND", "localhost:9009")
	server.b = listen
	server.l.Infof("Server listening on '%s'", listen)

	if _, set := os.LookupEnv("IDP_SSL_CERT"); set {
		server.l.Info("SSL enabled")
		// IDP_SSL_CERT set, so we run SSL mode
		err := http.ListenAndServeTLS(listen, os.Getenv("IDP_SSL_CERT"), os.Getenv("IDP_SSL_KEY"), server.logRequest(server.h))
		if err != nil {
			panic(err)
		}
	} else {
		err = http.ListenAndServe(listen, server.logRequest(server.h))
		if err != nil {
			panic(err)
		}
	}
}
