# ---------------------------------------------------------
# ÉTAPE 1 : BUILD (Compilation Go)
# ---------------------------------------------------------
FROM golang:1.23-alpine AS builder

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPROXY=https://proxy.golang.org

WORKDIR /app

# Dépendances Go
COPY go.mod go.sum ./
RUN go mod download

# Code source
COPY . .

# Build du binaire PORA
RUN go build -o /bin/pora-ranking-service .

# ---------------------------------------------------------
# ÉTAPE 2 : IMAGE FINALE (minimaliste)
# ---------------------------------------------------------
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/pora-ranking-service /bin/pora-ranking-service

ENV TZ=Etc/UTC

WORKDIR /app

EXPOSE 8080

ENTRYPOINT ["/bin/pora-ranking-service"]