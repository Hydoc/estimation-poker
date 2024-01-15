FROM node:18-alpine as base

ARG node_env="production"

ENV NODE_ENV=${node_env}
ENV TZ=Europe/Berlin

RUN mkdir -p /opt/frontend

WORKDIR /opt/frontend

FROM base as build
RUN npm i -g pnpm

COPY . .
RUN pnpm install --production=false
RUN pnpm run build


FROM build as dev

CMD ["pnpm", "dev"]