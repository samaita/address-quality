# syntax=docker/dockerfile:1
FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
ARG AQ_API_BASE_URL=/address-quality
ARG AQ_API_KEY=
ENV AQ_API_BASE_URL=$AQ_API_BASE_URL
ENV AQ_API_KEY=$AQ_API_KEY
RUN npm run build

FROM nginx:1.27-alpine
COPY nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html
# Render the template and validate the config at build time (bad config fails the build).
# The entrypoint re-renders with the real environment at container start.
RUN set -eux; \
    export SERVER_NAME="__ci__" API_UPSTREAM="127.0.0.1" TZ="UTC"; \
    /docker-entrypoint.d/20-envsubst-on-templates.sh; \
    nginx -t
EXPOSE 80
