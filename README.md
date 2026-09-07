# saml-test-idp

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/beryju/saml-test-idp/ci-build.yml?branch=main&style=for-the-badge)

This is a small, golang-based SAML Identity Provider, to be used in End-to-end or other testing. It uses the https://github.com/crewjam/saml Library for the actual SAML Logic.

saml-test-idp supports IdP-initiated Login flows. `/login/test-app/<suffix>` passes `/<suffix>` as RelayState.

This tool is full configured using environment variables.

## URLs

- `http://localhost:9009/health`: Healthcheck URL, used by the docker healtcheck.
- `http://localhost:9009/login/test-app`: Start IDP-initiated login.
- `http://localhost:9009/sso`: SAML SSO URL, needed to configure your SP.
- `http://localhost:9009/metadata`: SAML Metadata URL, needed to configure your SP.
- `http://localhost:9009/`: Test URL, redirects to SAML SSO URL.

## Users

Two users are created on startup:

| Username | Password    | Groups                 |
| -------- | ----------- | ---------------------- |
| `user1`  | `user1pass` | Administrators, Users  |
| `user2`  | `user2pass` | Users                  |

## Configuration

- `IDP_BIND`: Which address and port to bind to. Defaults to `localhost:9009` (the docker image sets `0.0.0.0:9009`).
- `IDP_ROOT_URL`: Root URL you're using to access the IDP. Defaults to `http://localhost:9009`, or `https://localhost:9009` when `IDP_SSL_CERT` is set.
- `IDP_METADATA_URL`: **Required.** URL the SP metadata is fetched from, on startup.

---

Optionally, if you want to use SSL, set these variables

- `IDP_SSL_CERT`: Path to the SSL Certificate the server should use.
- `IDP_SSL_KEY`: Path to the SSL Key the server should use.
- `IDP_SIGN_REQUESTS`: Set to `true` to sign responses with the SSL cert/key instead of a generated self-signed one.

Note: If you're manually setting `IDP_ROOT_URL`, ensure that you prefix that URL with https.

## Running

This service is intended to run in a docker container

```
# beryju.org is a vanity URL for ghcr.io/beryju
docker pull beryju.io/saml-test-idp
docker run -d --rm \
    -p 9009:9009 \
    -e IDP_METADATA_URL=http://some.site.tld/saml/metadata \
    beryju.io/saml-test-idp
```

Or if you want to use docker-compose, use this in your `docker-compose.yaml`.

```yaml
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
