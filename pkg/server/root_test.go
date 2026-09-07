package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
