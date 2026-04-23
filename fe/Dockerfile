FROM node:24.14.1-alpine AS deps
WORKDIR /app
RUN corepack enable

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

FROM node:24.14.1-alpine AS builder
WORKDIR /app
RUN corepack enable

COPY --from=deps app/node_modules ./node_modules

COPY . .
ARG API_BASE_URL
ENV API_BASE_URL=${API_BASE_URL}
RUN pnpm build

FROM node:24.14.1-alpine AS runner
WORKDIR /app
RUN corepack enable

ENV NODE_ENV=production
ENV PORT=300

COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/pnpm-lock.yaml ./pnpm-lock.yaml
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/next.config.ts ./next.config.ts

EXPOSE 3000

CMD ["pnpm", "start"]