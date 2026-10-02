FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN go run github.com/a-h/templ/cmd/templ@v0.3.960 generate
RUN CGO_ENABLED=0 go build -buildvcs=false -o /app ./cmd

FROM alpine:3.21
COPY --from=build /app /app
CMD ["/app"]
