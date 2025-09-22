# ---------------------------
# Frontend Builder (Node.js)
# ---------------------------
FROM node:20-bullseye AS ui-builder

WORKDIR /app/frontend

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile

COPY frontend .
RUN pnpm run build

# ---------------------------
# Backend Builder (Go 1.25.1)
# ---------------------------
FROM golang:1.25.1 AS api-builder

WORKDIR /app

# Copy go modules first for caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Copy built frontend
COPY --from=ui-builder /app/frontend/dist ./frontend/dist

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# ---------------------------
# Runtime stage
# ---------------------------
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=api-builder /app/main /app/
COPY --from=api-builder /app/frontend/dist /app/frontend/dist

EXPOSE 8085
CMD ["./main"]
