FROM node:20-alpine as base

WORKDIR /opt/frontend
RUN npm i -g pnpm
COPY . .
RUN pnpm install --production=false
RUN pnpm run build

FROM base as dev

CMD ["pnpm", "dev"]