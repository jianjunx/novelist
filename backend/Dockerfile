# ---- Build stage ----
FROM golang:1.25-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server

# ---- Run stage ----
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/server /server

EXPOSE 8080
CMD ["/server"]
