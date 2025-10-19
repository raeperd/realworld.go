FROM golang:1.24-bullseye as builder
ARG VERSION=local

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 go build -o app -ldflags '-w -X main.Version=${VERSION}' .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /src/app /bin/app

EXPOSE 8080
ENTRYPOINT ["app"]
CMD ["--port=8080"]
