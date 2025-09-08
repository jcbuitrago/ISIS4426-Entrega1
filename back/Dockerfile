# --- etapa de build ---
    FROM golang:1.25.1-alpine3.22 AS build
    WORKDIR /app
    COPY go.mod go.sum ./
    RUN go mod download
    COPY . .
    
    # variable PKG para indicar el paquete que se compila
    ARG PKG=.
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/app ${PKG}
    
    # --- etapa runtime ---
    FROM alpine:3.20
    WORKDIR /app
    RUN apk add --no-cache ca-certificates ffmpeg
    COPY --from=build --chmod=755 /app/bin/app /app/app
    ENV PORT=8080
    EXPOSE 8080
    ENTRYPOINT ["/app/app"]
    