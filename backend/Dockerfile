FROM golang:1.24-alpine AS base

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:latest AS prod

COPY --from=base /server /server

CMD ["/server"]

FROM base AS dev

RUN go install github.com/air-verse/air@latest && go install github.com/go-delve/delve/cmd/dlv@latest

CMD ["air", "-c", ".air.toml"]