# ---- Frontend build ----
FROM node:20-alpine AS frontend
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# ---- Backend build ----
FROM golang:1.25-bookworm AS backend
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o server ./cmd/server

# ---- Final image ----
FROM nginx:alpine

# Frontend static files
COPY --from=frontend /app/dist /usr/share/nginx/html

# Backend binary
COPY --from=backend /app/server /server

# Nginx config: serve frontend + proxy /api to backend
COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf

# Startup script: run nginx + backend
COPY deploy/start.sh /start.sh
RUN chmod +x /start.sh

EXPOSE 80
CMD ["/start.sh"]
