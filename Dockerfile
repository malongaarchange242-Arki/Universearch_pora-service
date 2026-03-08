# ---------------------------------------------------------
# ÉTAPE 1 : BUILD (Compilation Go)
# ---------------------------------------------------------
    FROM golang:1.21-alpine AS builder

    # Configuration Go
    ENV CGO_ENABLED=0 \
        GO111MODULE=on \
        GOPROXY=https://proxy.golang.org
    
    WORKDIR /app
    
    # ---------------------------------------------------------
    # Dépendances Go
    # ---------------------------------------------------------
    COPY go.mod go.sum ./
    RUN go mod download
    
    # ---------------------------------------------------------
    # Code source
    # ---------------------------------------------------------
    COPY . .
    
    # ---------------------------------------------------------
    # Build du binaire PORA
    # ---------------------------------------------------------
    RUN go build -o /bin/pora-ranking-service .
    
    
    # ---------------------------------------------------------
    # ÉTAPE 2 : IMAGE FINALE (minimaliste & sécurisée)
    # ---------------------------------------------------------
    FROM alpine:3.18
    
    # Certificats nécessaires pour HTTPS (Supabase)
    RUN apk add --no-cache ca-certificates
    
    # Copier le binaire compilé
    COPY --from=builder /bin/pora-ranking-service /bin/pora-ranking-service
    
    # Fuseau horaire (important pour cron & logs)
    ENV TZ=Etc/UTC
    
    # Dossier de travail
    WORKDIR /app
    
    # Port exposé par l’API
    EXPOSE 8080
    
    # Lancement du service PORA
    ENTRYPOINT ["/bin/pora-ranking-service"]
    