FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN go run github.com/a-h/templ/cmd/templ@v0.3.960 generate
ARG VERSION=dev
ARG REVISION=unknown
RUN CGO_ENABLED=0 go build -buildvcs=false \
    -ldflags "-X bookshop/frontend/internal/services/metrics.version=${VERSION} -X bookshop/frontend/internal/services/metrics.revision=${REVISION}" \
    -o /app ./cmd

FROM alpine:3.21
COPY --from=build /app /app
COPY config/config.yaml /config/config.yaml
CMD ["/app"]
