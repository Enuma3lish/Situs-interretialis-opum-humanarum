# build 階段
FROM node:20-alpine AS builder
WORKDIR /app
COPY . .
RUN npm install && npm run build

# serve 階段
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
# 可選：自訂 nginx.conf
EXPOSE 80
