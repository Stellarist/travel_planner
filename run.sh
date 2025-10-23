#!/bin/bash
# Travel Planner 启动脚本
# Docker 环境：启动 Redis、后端和 Nginx
# 开发环境：启动 Redis、后端和前端

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}========================================"
echo -e "  Travel Planner 开发环境启动脚本"
echo -e "========================================${NC}"
echo ""

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# 检测是否在 Docker 环境中
IN_DOCKER=false
if [ -f /.dockerenv ] || grep -q 'docker\|lxc' /proc/1/cgroup 2>/dev/null; then
    IN_DOCKER=true
fi

# 检查 Redis 是否已安装
if ! command -v redis-server &> /dev/null; then
    echo -e "${YELLOW}[警告] Redis 未安装或未在 PATH 中${NC}"
    echo -e "${YELLOW}请确保 Redis 已安装并运行在 127.0.0.1:6379${NC}"
    echo ""
    read -p "是否继续启动后端和前端? (y/n): " continue_without_redis
    if [ "$continue_without_redis" != "y" ]; then
        exit 1
    fi
    REDIS_AVAILABLE=false
else
    echo -e "${GREEN}[✓] 检测到 Redis${NC}"
    REDIS_AVAILABLE=true
fi

# Docker 环境下跳过 Go 和 Node 检查
if [ "$IN_DOCKER" = false ]; then
    # 检查 Go 是否已安装
    if ! command -v go &> /dev/null; then
        echo -e "${RED}[✗] Go 未安装或未在 PATH 中${NC}"
        exit 1
    fi
    echo -e "${GREEN}[✓] 检测到 Go${NC}"

    # 检查 Node.js 是否已安装
    if ! command -v node &> /dev/null; then
        echo -e "${RED}[✗] Node.js 未安装或未在 PATH 中${NC}"
        exit 1
    fi
    echo -e "${GREEN}[✓] 检测到 Node.js${NC}"
fi

echo ""
echo -e "${CYAN}正在启动所有服务...${NC}"
echo ""

# 创建日志目录
mkdir -p logs

# PID 文件
REDIS_PID_FILE="logs/redis.pid"
BACKEND_PID_FILE="logs/backend.pid"
FRONTEND_PID_FILE="logs/frontend.pid"

# 清理函数
cleanup() {
    echo ""
    echo -e "${YELLOW}正在停止所有服务...${NC}"
    
    if [ -f "$REDIS_PID_FILE" ]; then
        REDIS_PID=$(cat "$REDIS_PID_FILE")
        if ps -p $REDIS_PID > /dev/null 2>&1; then
            echo -e "  ${CYAN}停止 Redis (PID: $REDIS_PID)...${NC}"
            kill $REDIS_PID 2>/dev/null
        fi
        rm -f "$REDIS_PID_FILE"
    fi
    
    if [ -f "$BACKEND_PID_FILE" ]; then
        BACKEND_PID=$(cat "$BACKEND_PID_FILE")
        if ps -p $BACKEND_PID > /dev/null 2>&1; then
            echo -e "  ${CYAN}停止后端 (PID: $BACKEND_PID)...${NC}"
            kill $BACKEND_PID 2>/dev/null
        fi
        rm -f "$BACKEND_PID_FILE"
    fi
    
    if [ -f "$FRONTEND_PID_FILE" ]; then
        FRONTEND_PID=$(cat "$FRONTEND_PID_FILE")
        if ps -p $FRONTEND_PID > /dev/null 2>&1; then
            echo -e "  ${CYAN}停止前端 (PID: $FRONTEND_PID)...${NC}"
            kill $FRONTEND_PID 2>/dev/null
        fi
        rm -f "$FRONTEND_PID_FILE"
    fi
    
    echo ""
    echo -e "${GREEN}所有服务已停止!${NC}"
    exit 0
}

# 注册信号处理
trap cleanup SIGINT SIGTERM

# 启动 Redis
if [ "$IN_DOCKER" = true ]; then
    echo -e "${YELLOW}[Redis] 启动 Redis (Docker 模式)...${NC}"
    redis-server --port 6379 --bind 127.0.0.1 --daemonize yes
    sleep 2
    echo -e "${GREEN}[Redis] ✓ Redis 已启动${NC}"
elif [ "$REDIS_AVAILABLE" = true ]; then
    echo -e "${YELLOW}[Redis] 启动 Redis 服务器...${NC}"
    redis-server --port 6379 --bind 127.0.0.1 --daemonize no > logs/redis.log 2>&1 &
    REDIS_PID=$!
    echo $REDIS_PID > "$REDIS_PID_FILE"
    sleep 2
    
    if ps -p $REDIS_PID > /dev/null; then
        echo -e "${GREEN}[Redis] ✓ Redis 已启动 (PID: $REDIS_PID)${NC}"
    else
        echo -e "${RED}[Redis] ✗ Redis 启动失败${NC}"
        cat logs/redis.log
        cleanup
    fi
fi

# 启动后端
if [ "$IN_DOCKER" = true ]; then
    echo -e "${YELLOW}[后端] 启动后端服务 (Docker 模式)...${NC}"
    /app/travel_planner_server > /app/logs/backend.log 2>&1 &
    BACKEND_PID=$!
    sleep 2
    echo -e "${GREEN}[后端] ✓ 后端已启动${NC}"
else
    echo -e "${YELLOW}[后端] 启动 Go 后端服务...${NC}"
    cd backend
    go run main.go > ../logs/backend.log 2>&1 &
    BACKEND_PID=$!
    cd ..
    echo $BACKEND_PID > "$BACKEND_PID_FILE"
    sleep 2
    
    if ps -p $BACKEND_PID > /dev/null; then
        echo -e "${GREEN}[后端] ✓ 后端已启动 (PID: $BACKEND_PID)${NC}"
        echo -e "${CYAN}[后端] API 地址: http://127.0.0.1:3000${NC}"
    else
        echo -e "${RED}[后端] ✗ 后端启动失败${NC}"
        cat logs/backend.log
        cleanup
    fi
fi

# 启动前端或 Nginx
if [ "$IN_DOCKER" = true ]; then
    echo -e "${YELLOW}[Nginx] 启动 Nginx (Docker 模式)...${NC}"
    nginx -g 'daemon off;' &
    echo -e "${GREEN}[Nginx] ✓ Nginx 已启动${NC}"
    echo -e "${CYAN}访问地址: http://localhost:8080${NC}"
else
    echo -e "${YELLOW}[前端] 启动前端开发服务器...${NC}"
    cd frontend
    npm run dev > ../logs/frontend.log 2>&1 &
    FRONTEND_PID=$!
    cd ..
    echo $FRONTEND_PID > "$FRONTEND_PID_FILE"
    sleep 3
    
    if ps -p $FRONTEND_PID > /dev/null; then
        echo -e "${GREEN}[前端] ✓ 前端已启动 (PID: $FRONTEND_PID)${NC}"
        echo -e "${CYAN}[前端] 访问地址: http://localhost:5173${NC}"
    else
        echo -e "${RED}[前端] ✗ 前端启动失败${NC}"
        cat logs/frontend.log
        cleanup
    fi
fi

echo ""
echo -e "${GREEN}========================================"
echo -e "  所有服务已启动!"
echo -e "========================================${NC}"

if [ "$IN_DOCKER" = true ]; then
    echo -e "${CYAN}Docker 模式${NC}"
    echo -e "  • Redis:  127.0.0.1:6379"
    echo -e "  • 后端:   127.0.0.1:3000"
    echo -e "  • Nginx:  0.0.0.0:80"
    echo ""
    # Docker 模式下一直运行
    tail -f /app/logs/backend.log
else
    echo ""
    echo -e "${CYAN}服务信息:${NC}"
    if [ "$REDIS_AVAILABLE" = true ]; then
        echo -e "  ${NC}• Redis:  127.0.0.1:6379 (PID: $REDIS_PID)${NC}"
    fi
    echo -e "  ${NC}• 后端:   http://127.0.0.1:3000 (PID: $BACKEND_PID)${NC}"
    echo -e "  ${NC}• 前端:   http://localhost:5173 (PID: $FRONTEND_PID)${NC}"
    echo ""
    echo -e "${CYAN}查看日志:${NC}"
    if [ "$REDIS_AVAILABLE" = true ]; then
        echo -e "  ${NC}Redis:  tail -f logs/redis.log${NC}"
    fi
    echo -e "  ${NC}后端:   tail -f logs/backend.log${NC}"
    echo -e "  ${NC}前端:   tail -f logs/frontend.log${NC}"
    echo ""
    echo -e "${YELLOW}按 Ctrl+C 停止所有服务${NC}"
    echo ""
    
    # 持续运行并监控服务
    while true; do
        sleep 2
        
        # 检查服务是否还在运行
        if [ "$REDIS_AVAILABLE" = true ] && [ -f "$REDIS_PID_FILE" ]; then
            if ! ps -p $(cat "$REDIS_PID_FILE") > /dev/null 2>&1; then
                echo -e "${RED}[错误] Redis 服务已停止${NC}"
                cat logs/redis.log
                cleanup
            fi
        fi
        
        if [ -f "$BACKEND_PID_FILE" ]; then
            if ! ps -p $(cat "$BACKEND_PID_FILE") > /dev/null 2>&1; then
                echo -e "${RED}[错误] 后端服务已停止${NC}"
                cat logs/backend.log
                cleanup
            fi
        fi
        
        if [ -f "$FRONTEND_PID_FILE" ]; then
            if ! ps -p $(cat "$FRONTEND_PID_FILE") > /dev/null 2>&1; then
                echo -e "${RED}[错误] 前端服务已停止${NC}"
                cat logs/frontend.log
                cleanup
            fi
        fi
    done
fi
