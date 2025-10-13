#!/bin/bash
openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 -nodes \
  -keyout saml-idp.key -out saml-idp.pem -subj "/CN=localhost"
export IDP_SSL_CERT=./saml-idp.pem
export IDP_SSL_KEY=./saml-idp.key
