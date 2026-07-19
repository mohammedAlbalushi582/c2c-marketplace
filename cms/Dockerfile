# Strapi CMS (static/marketing content). Uses its own Postgres database.
FROM node:20-alpine AS build
RUN apk add --no-cache build-base python3
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
# The admin-panel bundle needs more heap than Node auto-sizes on small VPSes.
ENV NODE_OPTIONS=--max-old-space-size=3072
RUN npm run build

FROM node:20-alpine AS runtime
RUN apk add --no-cache vips-dev
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app ./
EXPOSE 1337
CMD ["npm", "run", "start"]
