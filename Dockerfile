# Multi-stage build
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Go backend stage
FROM golang:1.21-alpine AS backend-builder

RUN apk add --no-cache exiftool

WORKDIR /app/backend
COPY backend/go.mod ./
RUN go mod download

COPY backend/ .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

RUN apk add --no-cache exiftool

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/backend/main .

# Copy frontend build
COPY --from=frontend-builder /app/frontend/build ./static

# Create directories for data and cache
RUN mkdir -p /app/data /app/cache

EXPOSE 3000

CMD ["./main"]