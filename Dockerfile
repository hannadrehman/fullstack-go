# a little gopher packs its suitcase in the builder stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
# dependencies first, so the cache stays happy between builds
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

# pack light: only the compiled binary travels to the final stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /server /server
# port 8080 is the front door, knock politely
EXPOSE 8080
ENTRYPOINT ["/server"]