# Use an official Golang image to build the backend
FROM golang:alpine AS backend-build

WORKDIR /app

# Copy backend source and build it
COPY ./go.mod /app
COPY ./go.sum /app
COPY ./internal /app/internal
COPY ./cmd /app/cmd
RUN go build -o backend ./cmd/main.go

# Use a Node.js image to build the frontend
FROM node:alpine AS frontend-build

WORKDIR /app

# Copy frontend source and install dependencies
COPY flowchartmaker-frontend/package*.json /app/

RUN npm install

COPY flowchartmaker-frontend/public /app/public
COPY flowchartmaker-frontend/src /app/src
COPY flowchartmaker-frontend/index.html /app/index.html
COPY flowchartmaker-frontend/vite.config.js /app/vite.config.js
COPY flowchartmaker-frontend/jsconfig.json /app/jsconfig.json

RUN npm run build

# Final stage: Use nginx:alpine to serve everything
FROM nginx:alpine

# Copy the backend binary
COPY --from=backend-build /app/backend /app/backend
RUN mkdir -p /storage

# Copy frontend build files to Nginx's HTML directory
COPY --from=frontend-build /app/dist /usr/share/nginx/html

# Copy custom Nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Expose ports
EXPOSE 80 8080

# Start the backend and Nginx
CMD ["/bin/sh", "-c", "/app/backend & nginx -g 'daemon off;'"]
