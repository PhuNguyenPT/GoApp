FROM golang:1.25.5-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main cmd/api/main.go

FROM golang:1.25.5-alpine AS watch
WORKDIR /app
RUN apk add --no-cache nodejs npm wget && \
    go install github.com/air-verse/air@latest && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && \
    go install github.com/a-h/templ/cmd/templ@latest && \
    go clean -cache -modcache -testcache
COPY go.mod go.sum ./
RUN go mod download
COPY frontend-template/package*.json ./frontend-template/
RUN cd frontend-template && npm ci && npm cache clean --force
CMD ["air", "-c", ".air.docker.toml"]

FROM alpine:3.23 AS prod
RUN apk add --no-cache wget
WORKDIR /app
COPY --from=build /app/main /app/main
COPY --from=build /app/frontend-template /app/frontend-template
EXPOSE ${PORT}
EXPOSE ${TLS_PORT}
CMD ["/app/main"]

FROM node:24-alpine AS frontend_builder
WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/. .
RUN npm run build

FROM nginx:alpine AS frontend
RUN rm /etc/nginx/conf.d/default.conf
COPY --from=frontend_builder /frontend/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]