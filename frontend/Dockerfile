FROM node:20-alpine as base

WORKDIR /opt/frontend
RUN npm i -g pnpm
COPY . .
RUN pnpm install --production=false
RUN pnpm run build

FROM caddy:2.7.6-alpine as prod

COPY --from=base /opt/frontend/dist /usr/share/caddy

FROM base as dev

CMD ["pnpm", "dev"]