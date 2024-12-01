FROM golang:1.23.3-bullseye AS build

WORKDIR /app

COPY src/go.mod src/go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go mod download

FROM build AS dev

RUN go install github.com/cosmtrek/air@latest && \
  go install github.com/go-delve/delve/cmd/dlv@latest

COPY src .

CMD ["air", "-c", ".air.toml"]

FROM build AS build-production

RUN useradd -u 1001 appuser

COPY src .

RUN go build \
  -ldflags="-linkmode external -extldflags -static" \
  -tags netgo \
  -o zte-sms-read

# ca-certificates is required for HTTPS
RUN apt-get update && apt-get install -y ca-certificates

FROM scratch

ENV GIN_MODE=release

WORKDIR /

COPY --from=build-production /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-production /etc/passwd /etc/passwd
COPY --from=build-production /app/zte-sms-read zte-sms-read

USER appuser

EXPOSE 3000

CMD ["/zte-sms-read"]
