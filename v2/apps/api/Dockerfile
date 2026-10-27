FROM node:22-alpine
WORKDIR /app
COPY package.json package-lock.json* ./
COPY apps/api/package.json apps/api/
RUN npm ci -w apps/api --include-workspace-root
COPY db db
COPY apps/api apps/api
RUN npm run build -w apps/api
ENV NODE_ENV=production PORT=4000
CMD ["sh","-c","node apps/api/dist/migrate.js && node apps/api/dist/server.js"]
