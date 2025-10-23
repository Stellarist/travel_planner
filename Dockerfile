# 多阶段构建 Dockerfile - 包含前端、后端和 Redis

# ==================== 前端构建阶段 ====================
FROM node:20-alpine AS frontend-builder

WORKDIR /app

# 复制配置文件到项目根目录（前端构建时需要）
COPY config.json ./

WORKDIR /app/frontend

# 复制前端依赖文件
COPY frontend/package*.json ./

# 安装依赖
RUN npm install

# 复制前端源代码
COPY frontend/ ./

# 构建前端
RUN npm run build

# ==================== 后端构建阶段 ====================
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git gcc g++ musl-dev

# 复制 Go 模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制后端源代码
COPY backend/ ./backend/

# 构建后端应用
RUN CGO_ENABLED=1 GOOS=linux go build -o travel_planner_server ./backend/main.go

# ==================== 最终运行阶段 ====================
FROM redis:7-alpine

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates nginx bash

# 创建应用目录
WORKDIR /app

# 从后端构建阶段复制编译好的二进制文件
COPY --from=backend-builder /app/travel_planner_server /app/

# 从前端构建阶段复制构建好的静态文件
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# 复制配置文件和启动脚本
COPY config.json /app/config.json
COPY run.sh /app/run.sh
RUN chmod +x /app/run.sh

# 创建日志目录
RUN mkdir -p /app/logs /app/backend/logs

# 配置 Nginx 用于服务前端和反向代理后端
RUN mkdir -p /run/nginx
COPY <<EOF /etc/nginx/http.d/default.conf
server {
    listen 80;
    server_name localhost;

    # 前端静态文件
    location / {
        root /app/frontend/dist;
        try_files \$uri \$uri/ /index.html;
    }

    # 后端 API 代理 - 代理所有到 127.0.0.1:3000 的请求
    location ~ ^/(api/.*|health)$ {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

# 暴露端口
EXPOSE 80

# 使用 run.sh 启动
CMD ["/app/run.sh"]
