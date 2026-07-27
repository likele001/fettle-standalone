#!/bin/bash
# fettle 启动脚本
# 按依赖顺序启动所有服务

set -e

echo "=== Starting fettle AI Agent Platform ==="

# 1. 检查基础服务
echo "[1/3] Starting infrastructure services..."
cd "$(dirname "$0")/../deploy/docker"
docker compose -f docker-compose.infra.yml up -d
echo "Infrastructure services started."

# 2. 等待数据库就绪
echo "[2/3] Waiting for database..."
sleep 5

# 3. 启动后端服务
echo "[3/3] Starting backend services..."
docker compose -f docker-compose.services.yml up -d
echo "All services started."

echo ""
echo "=== Services Available ==="
echo "Gateway:       http://localhost:20001"
echo "User Service:  http://localhost:20002"
echo "Agent Service: http://localhost:20003"
echo "Chat Service:  http://localhost:20004"
echo "Skill Service: http://localhost:20005"
echo "Billing:       http://localhost:20006"
echo "AI Engine:     http://localhost:20007"
echo "AI Engine gRPC: localhost:20008"
echo "Web Admin:     http://localhost:20009"
echo ""
echo "To view logs: docker compose -f deploy/docker/docker-compose.yml logs -f"
