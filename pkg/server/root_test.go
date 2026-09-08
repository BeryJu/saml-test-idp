package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/crewjam/saml"
)

const spMetadata = `<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://sp.example.com/metadata">
  <SPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://sp.example.com/acs" index="0"/>
  </SPSSODescriptor>
</EntityDescriptor>`

func TestServer(t *testing.T) {
	sp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(spMetadata))
	}))
	defer sp.Close()
	t.Setenv("IDP_METADATA_URL", sp.URL)

	idp := httptest.NewServer(newServer().h)
	defer idp.Close()

	for _, tc := range []struct{ path, contains string }{
		{"/health", "hello :)"},
		{"/metadata", "IDPSSODescriptor"},
		{"/services/test-app", "https://sp.example.com/metadata"},
		{"/shortcuts/test-app", "https://sp.example.com/metadata"},
		{"/login/test-app", "<form"},
	} {
		res, err := http.Get(idp.URL + tc.path)
		if err != nil {
			t.Fatalf("GET %s: %v", tc.path, err)
		}
		body, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("GET %s: status %d", tc.path, res.StatusCode)
		}
		if !strings.Contains(string(body), tc.contains) {
			t.Errorf("GET %s: body missing %q, got %s", tc.path, tc.contains, body)
		}
	}
}

type staticSessionProvider struct{ session *saml.Session }

func (p staticSessionProvider) GetSession(_ http.ResponseWriter, _ *http.Request, _ *saml.IdpAuthnRequest) *saml.Session {
	return p.session
}

func TestNameIDPolicyProvider(t *testing.T) {
	email := string(saml.EmailAddressNameIDFormat)
	persistent := string(saml.PersistentNameIDFormat)
	for _, tc := range []struct {
		name           string
		format         *string
		wantNameID     string
		wantNameIDForm string
	}{
		{"no policy", nil, "transient-id", ""},
		{"email", &email, "user1@example.com", email},
		{"persistent", &persistent, "user1", persistent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := nameIDPolicyProvider{staticSessionProvider{&saml.Session{
				NameID:    "transient-id",
				UserName:  "user1",
				UserEmail: "user1@example.com",
			}}}
			req := &saml.IdpAuthnRequest{}
			if tc.format != nil {
				req.Request.NameIDPolicy = &saml.NameIDPolicy{Format: tc.format}
			}
			session := p.GetSession(nil, nil, req)
			if session.NameID != tc.wantNameID {
				t.Errorf("NameID = %q, want %q", session.NameID, tc.wantNameID)
			}
			if session.NameIDFormat != tc.wantNameIDForm {
				t.Errorf("NameIDFormat = %q, want %q", session.NameIDFormat, tc.wantNameIDForm)
			}
		})
	}
}
