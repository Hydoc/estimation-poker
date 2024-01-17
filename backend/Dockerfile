FROM golang:1.21.6 as base

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM base as live

CMD ["/server"]

FROM base as dev

RUN go install github.com/cosmtrek/air@latest

CMD ["air", "-c", ".air.toml"]