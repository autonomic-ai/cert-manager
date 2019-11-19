BINS=(
  "controller"
  "webhook"
  "cainjector"
)

for bin in "${BINS[@]}"; do
  make $bin

  mv ./bazel-bin/cmd/${bin}/linux_amd64_pure_stripped/${bin} .

  cat <<EOF > Dockerfile.${bin}
FROM alpine:3.10.3

RUN apk add --no-cache ca-certificates && \
addgroup -S certmanager && adduser -S -G certmanager certmanager

ADD $bin /usr/bin/${bin}

USER certmanager
ENTRYPOINT [ "/usr/bin/${bin}" ]
ARG VCS_REF
LABEL org.label-schema.vcs-ref=$VCS_REF \
  org.label-schema.vcs-url="https://github.com/jetstack/cert-manager" \
  org.label-schema.license="Apache-2.0"
EOF
done

cat <<EOF > build_args
{"VCS_REF": "$(git rev-parse HEAD)"}
EOF
