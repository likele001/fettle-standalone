# 构建阶段
FROM node:20-alpine AS builder
WORKDIR /app
COPY frontend/web-admin/package.json ./
RUN npm install -g pnpm && pnpm install
COPY frontend/web-admin/ ./
RUN pnpm build

# 运行阶段
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY frontend/web-admin/deploy/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
