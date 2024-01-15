FROM golang:1.21.6

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

CMD ["/server"]