make controller

mv ./bazel-bin/cmd/controller/linux_amd64_pure_stripped/controller .

cat <<EOF > Dockerfile
FROM alpine:3.10.3

RUN apk add --no-cache ca-certificates && \
addgroup -S certmanager && adduser -S -G certmanager certmanager

ADD controller /usr/bin/cert-manager

USER certmanager
ENTRYPOINT ["/usr/bin/cert-manager"]
ARG VCS_REF
LABEL org.label-schema.vcs-ref=$VCS_REF \
  org.label-schema.vcs-url="https://github.com/jetstack/cert-manager" \
  org.label-schema.license="Apache-2.0"
EOF

cat <<EOF > build_args
{"VCS_REF": "$(git rev-parse HEAD)"}
EOF
