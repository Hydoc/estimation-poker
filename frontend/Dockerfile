FROM node:20-alpine AS base

WORKDIR /opt/frontend
RUN npm i -g pnpm
COPY . .
RUN CI=1 pnpm install --production=false
RUN pnpm run build

FROM caddy:2.7.6-alpine AS prod

COPY --from=base /opt/frontend/dist /usr/share/caddy

FROM base AS dev

CMD ["sh", "/opt/frontend/run.sh"]