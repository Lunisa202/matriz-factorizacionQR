# ---- Build stage: compila binario estático ----
FROM golang:1.22-alpine AS builder
WORKDIR /app
# Cache de dependencias
COPY go.mod ./
RUN go mod download
COPY . .
# Binario estático y sin símbolos para imagen mínima
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server .

# ---- Runtime stage: imagen mínima ----
FROM alpine:3.20
WORKDIR /app
# Certificados para llamadas HTTPS salientes (Render, etc.)
RUN apk add --no-cache ca-certificates
COPY --from=builder /server /app/server
EXPOSE 3001
ENV PORT=3001
CMD ["/app/server"]
