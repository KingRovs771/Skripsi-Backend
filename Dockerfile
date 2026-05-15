# =============================================
# Stage 1: Builder
# =============================================
FROM golang:1.24-alpine AS builder

# Install dependencies untuk build
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go.mod dan go.sum terlebih dahulu untuk cache layer
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Build binary dengan optimasi ukuran
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o server \
    ./main.go

# =============================================
# Stage 2: Runner (image minimal)
# =============================================
FROM alpine:3.20 AS runner

# Install ca-certificates dan tzdata agar HTTPS & timezone berfungsi
RUN apk add --no-cache ca-certificates tzdata

# Set timezone Asia/Jakarta
ENV TZ=Asia/Jakarta

WORKDIR /app

# Copy binary dari stage builder
COPY --from=builder /app/server .

# Railway secara otomatis meng-inject PORT, default 8080
ENV PORT=8080

EXPOSE 8080

# Jalankan binary
CMD ["./server"]
