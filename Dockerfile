FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/server ./cmd/server/main.go

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /app/server .
USER nonroot
EXPOSE 3000
ENTRYPOINT [ "/app/server" ]
