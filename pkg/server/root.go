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
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlidp"
	"github.com/crewjam/saml/samlsp"
	dsig "github.com/russellhaering/goxmldsig"
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

// nameIDPolicyProvider honours the NameIDPolicy requested by the SP. samlidp's own
// session provider never sets NameIDFormat, so the IdP always issues transient NameIDs.
type nameIDPolicyProvider struct {
	saml.SessionProvider
}

func (p nameIDPolicyProvider) GetSession(w http.ResponseWriter, r *http.Request, req *saml.IdpAuthnRequest) *saml.Session {
	session := p.SessionProvider.GetSession(w, r, req)
	if session == nil || req.Request.NameIDPolicy == nil || req.Request.NameIDPolicy.Format == nil {
		return session
	}
	session.NameIDFormat = *req.Request.NameIDPolicy.Format
	switch saml.NameIDFormat(session.NameIDFormat) {
	case saml.EmailAddressNameIDFormat:
		session.NameID = session.UserEmail
	case saml.PersistentNameIDFormat, saml.UnspecifiedNameIDFormat:
		session.NameID = session.UserName
	}
	return session
}

func mustBcrypt(pw string) []byte {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return h
}

func newServer() *Server {
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
	svc := samlidp.Service{
		Name:     "test-app",
		Metadata: *desc,
	}

	err = config.Store.Put("/services/test-app", svc)
	if err != nil {
		panic(err)
	}

	// Required for /login/test-app (IdP-initiated) to resolve
	err = config.Store.Put("/shortcuts/test-app", samlidp.Shortcut{
		Name:                  "test-app",
		ServiceProviderID:     desc.EntityID,
		URISuffixAsRelayState: true,
	})
	if err != nil {
		panic(err)
	}

	idp, err := samlidp.New(config)
	if err != nil {
		panic(err)
	}
	// https://github.com/crewjam/saml/issues/613
	idp.IDP.LoginURL = idp.IDP.SSOURL
	idp.IDP.SessionProvider = nameIDPolicyProvider{idp}
	// crewjam defaults to RSA-SHA1, which most SPs reject
	idp.IDP.SignatureMethod = dsig.RSASHA256SignatureMethod
	server := &Server{
		idp: idp,
		h:   http.NewServeMux(),
		l:   log.WithField("component", "server"),
		b:   helpers.Env("IDP_BIND", "localhost:9009"),
	}
	server.h.HandleFunc("/health", server.health)
	server.h.Handle("/", server.idp)
	return server
}

func RunServer() {
	server := newServer()
	server.l.Infof("Server listening on '%s'", server.b)

	if _, set := os.LookupEnv("IDP_SSL_CERT"); set {
		server.l.Info("SSL enabled")
		// IDP_SSL_CERT set, so we run SSL mode
		err := http.ListenAndServeTLS(server.b, os.Getenv("IDP_SSL_CERT"), os.Getenv("IDP_SSL_KEY"), server.logRequest(server.h))
		if err != nil {
			panic(err)
		}
	} else {
		err := http.ListenAndServe(server.b, server.logRequest(server.h))
		if err != nil {
			panic(err)
		}
	}
}
