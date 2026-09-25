# ---- Deps stage: instala solo producción ----
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install --omit=dev

# ---- Runtime stage ----
FROM node:20-alpine
WORKDIR /app
ENV NODE_ENV=production
# Usuario no-root por seguridad
RUN addgroup -S app && adduser -S app -G app
COPY --from=deps /app/node_modules ./node_modules
COPY src ./src
COPY package.json ./
USER app
EXPOSE 3002
ENV PORT=3002
CMD ["node", "src/index.js"]
