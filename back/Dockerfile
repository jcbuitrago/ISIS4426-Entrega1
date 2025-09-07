# -------- Build stage --------
    FROM golang:1.25.1-alpine3.22 AS build

    WORKDIR /app
    
    # Módulo del backend vive en /back
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    
    # El paquete a compilar se pasa por ARG:
    # - API   -> PKG=.
    # - Worker-> PKG=./cmd/worker
    ARG PKG=.
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/app ${PKG}
    
    # -------- Runtime stage --------
    FROM alpine:3.20
    WORKDIR /app
    RUN apk add --no-cache ca-certificates
    COPY --from=build /app/bin/app /app/app
    
    ENV PORT=8080
    EXPOSE 8080
    ENTRYPOINT ["/app/app"]
    