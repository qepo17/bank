FROM golang:1.24-bookworm AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o web ./cmd/web/main.go

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app

COPY --from=build /app/web /app/web

USER nonroot:nonroot

EXPOSE 80

CMD ["/app/web"]