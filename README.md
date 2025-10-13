# saml-test-idp

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/beryju/saml-test-idp/ci-build.yml?branch=main&style=for-the-badge)

This is a small, golang-based SAML Service Provider, to be used in End-to-end or other testing. It uses the https://github.com/crewjam/saml Library for the actual SAML Logic.

saml-test-idp supports IdP-initiated Login flows, *however* RelayState has to be empty for this to work.

This tool is full configured using environment variables.

## URLs

- `http://localhost:9009/health`: Healthcheck URL, used by the docker healtcheck.
- `http://localhost:9009/login/test-app`: Start IDP-initiated login.
- `http://localhost:9009/sso`: SAML SSO URL, needed to configure your SP.
- `http://localhost:9009/metadata`: SAML Metadata URL, needed to configure your SP.
- `http://localhost:9009/`: Test URL, redirects to SAML SSO URL.

## Configuration

- `IDP_BIND`: Which address and port to bind to. Defaults to `0.0.0.0:9009`.
- `IDP_ROOT_URL`: Root URL you're using to access the IDP. Defaults to `http://localhost:9009`.
<!-- - `IDP_ENTITY_ID`: SAML EntityID, defaults to `saml-test-idp`. -->
- `IDP_METADATA_URL`: Optional URL that metadata is fetched from. The metadata is fetched on the first request to `/`.
<!-- - `IDP_COOKIE_NAME`: Custom name for the session cookie. Defaults to `token`. Use this to avoid cookie name conflicts with other applications. -->
<!-- --- -->
<!-- - `IDP_SSO_URL`: If the metadata URL is not configured, use these options to configure it manually. -->
<!-- - `IDP_SSO_BINDING`: Binding Type used for the IdP, defaults to POST. Allowed values: `urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST` and `urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect` -->
<!-- - `IDP_SIGNING_CERT`: PEM-encoded Certificate used for signing, with the PEM Header and all newlines removed. -->
---
Optionally, if you want to use SSL, set these variables
- `IDP_SSL_CERT`: Path to the SSL Certificate the server should use.
- `IDP_SSL_KEY`: Path to the SSL Key the server should use.
- `IDP_SIGN_REQUESTS`: Enable signing of requests.

Note: If you're manually setting `IDP_ROOT_URL`, ensure that you prefix that URL with https.

## Running

This service is intended to run in a docker container

```
# beryju.org is a vanity URL for ghcr.io/beryju
docker pull beryju.io/saml-test-idp
docker run -d --rm \
    -p 9009:9009 \
    -e IDP_ENTITY_ID=saml-test-idp \
    -e IDP_SSO_URL=http://id.beryju.io/... \
    beryju.io/saml-test-idp
```

Or if you want to use docker-compose, use this in your `docker-compose.yaml`.

```yaml
version: '3.5'

services:
  saml-test-idp:
    image: beryju.io/saml-test-idp
    ports:
      - 9009:9009
    environment:
      IDP_METADATA_URL: http://some.site.tld/saml/metadata
    # If you don't want SSL, cut here
      IDP_SSL_CERT: /fullchain.pem
      IDP_SSL_KEY: /privkey.pem
    volumes:
      - ./fullchain.pem:/fullchain.pem
      - ./privkey.pem:/privkey.pem
```
