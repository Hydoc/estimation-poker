FROM node:20-alpine AS base

RUN apk add caddy supervisor



FROM base AS build
RUN npm i -g pnpm

WORKDIR /opt/backend
COPY ./backend /opt/backend
COPY --from=golang:1.24-alpine /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"
RUN go mod download
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server


WORKDIR /opt/build
COPY ./frontend /opt/build

RUN pnpm install --production=false
RUN pnpm run build




FROM base AS live

COPY --from=build /opt/build/dist /opt/frontend
COPY --from=build /server /server

ADD ./deployment/Caddyfile /etc/caddy/Caddyfile
ADD ./deployment/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

EXPOSE 80

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]

