# Install dependencies only for building
FROM node:20-alpine AS builder
WORKDIR /app

COPY package*.json ./
RUN npm install

ARG NEXT_PUBLIC_API_ORIGIN
ENV NEXT_PUBLIC_API_ORIGIN=$NEXT_PUBLIC_API_ORIGIN

# Copy all source files and build
COPY . .

ARG NEXT_PUBLIC_API_ORIGIN
ARG NEXT_PUBLIC_FRONTEND_URI
ARG API_LOCAL

ENV NEXT_PUBLIC_API_ORIGIN=$NEXT_PUBLIC_API_ORIGIN
ENV NEXT_PUBLIC_FRONTEND_URI=$NEXT_PUBLIC_FRONTEND_URI
ENV API_LOCAL=$API_LOCAL

RUN npm run build

# Production image
FROM node:20-alpine AS runner
WORKDIR /app

# Copy only build artifacts + package.json
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/next.config.ts ./next.config.ts

RUN npm install --production

CMD ["npm", "start"]