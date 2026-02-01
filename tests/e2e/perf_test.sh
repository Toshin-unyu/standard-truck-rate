#!/bin/bash

# パフォーマンステストスクリプト
# Standard Truck Rate - Phase 6

echo "========================================"
echo "パフォーマンステスト"
echo "========================================"
echo ""

# 色付き出力
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# ================================================
echo -e "${BLUE}--- 1. ヘルスチェック応答時間 ---${NC}"
for i in 1 2 3 4 5; do
    TIME=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:80/health)
    echo "  試行$i: ${TIME}s"
done
echo ""

# ================================================
echo -e "${BLUE}--- 2. 運賃計算API応答時間（コンテナ内直接） ---${NC}"
for i in 1 2 3 4 5; do
    START=$(date +%s.%N)
    docker compose exec -T str wget -qO- --post-data="region_code=3&vehicle_code=3&distance_km=100&driving_minutes=120" "http://localhost:8080/api/fare/calculate/json" >/dev/null 2>&1
    END=$(date +%s.%N)
    TIME=$(echo "$END - $START" | bc)
    echo "  試行$i: ${TIME}s"
done
echo ""

# ================================================
echo -e "${BLUE}--- 3. ルート取得API応答時間（モック） ---${NC}"
for i in 1 2 3 4 5; do
    START=$(date +%s.%N)
    docker compose exec -T str wget -qO- "http://localhost:8080/api/route?origin=東京&dest=大阪" >/dev/null 2>&1
    END=$(date +%s.%N)
    TIME=$(echo "$END - $START" | bc)
    echo "  試行$i: ${TIME}s"
done
echo ""

# ================================================
echo -e "${BLUE}--- 4. コンテナリソース使用状況 ---${NC}"
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"
echo ""

# ================================================
echo -e "${BLUE}--- 5. コンテナログ（最新10行） ---${NC}"
echo -e "${YELLOW}[strコンテナ]${NC}"
docker compose logs --tail=10 str 2>/dev/null | grep -v "level=warning"
echo ""
echo -e "${YELLOW}[nginxコンテナ]${NC}"
docker compose logs --tail=10 nginx 2>/dev/null
echo ""

# ================================================
echo -e "${BLUE}--- 6. 連続リクエストテスト（10回） ---${NC}"
TOTAL_TIME=0
for i in $(seq 1 10); do
    START=$(date +%s.%N)
    docker compose exec -T str wget -qO- --post-data="region_code=3&vehicle_code=3&distance_km=$((i * 50))&driving_minutes=120" "http://localhost:8080/api/fare/calculate/json" >/dev/null 2>&1
    END=$(date +%s.%N)
    TIME=$(echo "$END - $START" | bc)
    TOTAL_TIME=$(echo "$TOTAL_TIME + $TIME" | bc)
done
AVG_TIME=$(echo "scale=4; $TOTAL_TIME / 10" | bc)
echo "  合計: ${TOTAL_TIME}s"
echo "  平均: ${AVG_TIME}s"
echo ""

echo "========================================"
echo -e "${GREEN}パフォーマンステスト完了${NC}"
echo "========================================"
