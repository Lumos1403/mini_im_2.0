# 09 Docker 与部署文档

## 1. 部署目标

项目部署分三步：

```txt
第一步：本地直接运行后端、前端、MySQL、Redis
第二步：Docker Compose 本地运行全部服务
第三步：云服务器使用 Docker Compose + Nginx 上线
```

## 2. 环境变量

本地直接运行后端时建议 `.env.example`：

```env
APP_ENV=dev
SERVER_PORT=8081

MYSQL_ROOT_PASSWORD=rootpassword
MYSQL_DATABASE=go_im
MYSQL_USER=goim
MYSQL_PASSWORD=goimpassword
MYSQL_DSN=goim:goimpassword@tcp(127.0.0.1:3307)/go_im?charset=utf8mb4&parseTime=True&loc=Local

REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_ACCESS_SECRET=change-this-access-secret
JWT_REFRESH_SECRET=change-this-refresh-secret
JWT_ACCESS_EXPIRE_MINUTES=15
JWT_REFRESH_EXPIRE_DAYS=7

FILE_STORAGE_TYPE=local
FILE_LOCAL_PATH=/app/uploads
FILE_MAX_SIZE_MB=50

IM_GROUP_MAX_MEMBERS=50
IM_RECALL_MINUTES=5
```

生产环境必须修改所有密码和 Secret。

## 3. docker-compose.yml 示例

```yaml
version: "3.9"

services:
  mysql:
    image: mysql:8.0
    container_name: go-im-mysql
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: ${MYSQL_DATABASE}
      MYSQL_USER: ${MYSQL_USER}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    ports:
      - "3307:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./backend/migrations:/docker-entrypoint-initdb.d
    command: --default-authentication-plugin=mysql_native_password --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci

  redis:
    image: redis:7
    container_name: go-im-redis
    restart: always
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: go-im-backend
    restart: always
    depends_on:
      - mysql
      - redis
    env_file:
      - .env
    ports:
      - "8081:8081"
    volumes:
      - uploads_data:/app/uploads

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: go-im-frontend
    restart: always
    depends_on:
      - backend
    ports:
      - "5173:80"

volumes:
  mysql_data:
  redis_data:
  uploads_data:
```

## 4. 后端 Dockerfile 示例

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY --from=builder /app/configs /app/configs
RUN mkdir -p /app/uploads
EXPOSE 8081
CMD ["/app/server"]
```

## 5. 前端 Dockerfile 示例

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## 6. Nginx 配置示例

前端容器内 nginx：

```nginx
server {
  listen 80;
  server_name _;

  root /usr/share/nginx/html;
  index index.html;

  location / {
    try_files $uri $uri/ /index.html;
  }

  location /api/ {
    proxy_pass http://backend:8081/api/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
  }

  location /ws {
    proxy_pass http://backend:8081/ws;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_read_timeout 3600s;
  }
}
```

## 7. 本地启动流程

```bash
cp .env.example .env
docker compose up -d --build
```

访问：

```txt
前端：http://localhost:5173
后端：http://localhost:8081
MySQL：127.0.0.1:3307 -> 容器 3306
Redis：127.0.0.1:6379
```

如果后端也运行在 Docker Compose 网络内，后端容器访问 MySQL 和 Redis 时可使用服务名 `mysql:3306`、`redis:6379`。

## 8. 云服务器上线流程

```txt
购买云服务器
安装 Docker 和 Docker Compose
开放 80、443、22 端口
上传项目代码
配置 .env
配置域名解析
启动 docker compose
配置 HTTPS，可使用 Nginx Proxy Manager 或 Certbot
```

## 9. 数据持久化

必须持久化：

```txt
mysql_data
redis_data
uploads_data
```

否则重启容器可能导致数据丢失。

## 10. 健康检查

后端需要提供：

```txt
GET /api/health
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok"
  }
}
```

## 11. 部署注意事项

```txt
生产环境禁止使用默认 JWT Secret
生产环境禁止使用弱数据库密码
uploads 必须挂载 volume
数据库 migration 要可重复执行或可控执行
Nginx 必须支持 WebSocket Upgrade
前端构建时后端 API 地址必须可配置
```
