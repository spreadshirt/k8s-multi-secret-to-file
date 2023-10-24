FROM docker.io/golang:1.21.3-alpine3.18 as builder
RUN apk add --no-cache shadow && useradd --home-dir /dev/null --shell /bin/false scratchuser && apk del shadow

WORKDIR /build
COPY . .
RUN --mount=type=cache,target=/root/.go/pkg --mount=type=cache,target=/root/.cache/go-build \
  go build -v -o k8s-multi-secret-to-file .

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
USER scratchuser
COPY --from=builder /build/k8s-multi-secret-to-file /usr/local/bin/k8s-multi-secret-to-file

ENTRYPOINT ["/usr/local/bin/k8s-multi-secret-to-file"]