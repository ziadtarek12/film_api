FROM golang:latest

WORKDIR /app

COPY . .

# Remove unused dependencies and update go.mod
RUN go mod tidy

RUN go build -o ./bin/api ./cmd/api 

EXPOSE ${APP_PORT}

# CMD ["./bin/api", "-addr=:${APP_PORT}", "-dsn=${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?parseTime=true"]