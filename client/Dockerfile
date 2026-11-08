FROM node:22-alpine AS builder
WORKDIR /app
COPY client/package*.json ./
RUN npm install
COPY client/ .
RUN npm run build
# note: midpoint of this file

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY client/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80