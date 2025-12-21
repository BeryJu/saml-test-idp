FROM docker.io/library/debian:13-slim
ARG TARGETPLATFORM

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates wget && \
    apt-get clean
COPY $TARGETPLATFORM/saml-test-idp /
EXPOSE 9009
WORKDIR /web-root
ENV IDP_BIND=0.0.0.0:9009
HEALTHCHECK --interval=5s --start-period=1s CMD [ "wget", "--spider", "http://localhost:9009/health" ]
ENTRYPOINT [ "/saml-test-idp" ]
