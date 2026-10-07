FROM node:22-bookworm-slim AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26.5-bookworm AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /go-flow ./cmd/go-flow

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=backend /go-flow /app/go-flow
COPY --from=frontend /app/web/dist /app/web/dist
ENV GIN_MODE=release STATIC_DIR=/app/web/dist PORT=8080
USER 65532:65532
EXPOSE 8080
CMD ["/app/go-flow"]
