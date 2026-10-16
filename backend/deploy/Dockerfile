FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -buildvcs=false -o /app ./cmd

FROM alpine:3.21
COPY --from=build /app /app
COPY config/config.yaml /config/config.yaml
CMD ["/app"]
