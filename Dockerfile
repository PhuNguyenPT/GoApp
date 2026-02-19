FROM golang:1.25.5-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main cmd/api/main.go

FROM golang:1.25.5-alpine AS watch
WORKDIR /app

RUN apk add --no-cache nodejs npm

COPY go.mod go.sum ./
RUN go mod download

COPY frontend-template/package*.json ./frontend-template/
RUN cd frontend-template && npm ci

RUN go install github.com/air-verse/air@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN go install github.com/a-h/templ/cmd/templ@latest

CMD ["air", "-c", ".air.docker.toml"]

FROM alpine:3.20.1 AS prod
WORKDIR /app
COPY --from=build /app/main /app/main
COPY --from=build /app/frontend-template /app/frontend-template
EXPOSE ${PORT}
CMD ["./main"]

FROM node:24 AS frontend_builder
WORKDIR /frontend

COPY frontend/package*.json ./
RUN npm ci
COPY frontend/. .
RUN npm run build

FROM node:24-slim AS frontend
RUN npm install -g serve
COPY --from=frontend_builder /frontend/dist /app/dist
EXPOSE 5173
CMD ["serve", "-s", "/app/dist", "-l", "5173"]
