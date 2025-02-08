FROM node:22 AS base

FROM base AS deps
WORKDIR /app

# Install dependencies based on the preferred package manager
COPY ui/package.json ./ui/package.json
COPY package.json package-lock.json ./

RUN npm ci --legacy-peer-deps

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY ui ./ui
COPY package.json package-lock.json ./
COPY --from=deps /app/node_modules ./node_modules

# Next.js collects completely anonymous telemetry data about general usage.
# Learn more here: https://nextjs.org/telemetry
# Uncomment the following line in case you want to disable telemetry during the build.
# ENV NEXT_TELEMETRY_DISABLED=1

RUN npm run -w ui build; 

# Production image, copy all the files and run next
FROM gcr.io/distroless/nodejs22-debian12:nonroot AS runner
WORKDIR /app

ENV NODE_ENV=production
# Uncomment the following line in case you want to disable telemetry during runtime.
# ENV NEXT_TELEMETRY_DISABLED=1

# Use non-root user instead of create a new one like the official example
# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=builder --chown=65532:65532 /app/ui/.next/standalone/ ./
COPY --from=builder --chown=65532:65532 /app/ui/public ./ui/public
COPY --from=builder --chown=65532:65532 /app/ui/.next/static ./ui/.next/static

# server.js is created by next build from the standalone output
# https://nextjs.org/docs/pages/api-reference/config/next-config-js/output
CMD ["./ui/server.js"]