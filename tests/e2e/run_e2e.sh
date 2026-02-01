#!/bin/bash

# E2Eテスト実行ラッパー
# コンテナ内でテストを実行

set -e

echo "========================================"
echo "E2Eテスト開始 - Docker Compose環境"
echo "========================================"
echo ""

# コンテナ起動確認
if ! docker compose ps | grep -q "str-1.*Up"; then
    echo "strコンテナが起動していません"
    echo "docker compose up -d を実行してください"
    exit 1
fi

# 色付き出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0

# テスト実行（コンテナ内でwgetを使用）
run_test() {
    local name="$1"
    local url="$2"
    local method="${3:-GET}"
    local data="$4"
    local expected="$5"

    if [ "$method" = "POST" ]; then
        RESPONSE=$(docker compose exec -T str wget -qO- --post-data="$data" "http://localhost:8080$url" 2>/dev/null || echo "ERROR")
    else
        RESPONSE=$(docker compose exec -T str wget -qO- "http://localhost:8080$url" 2>/dev/null || echo "ERROR")
    fi

    if echo "$RESPONSE" | grep -q "$expected"; then
        echo -e "${GREEN}✓ PASS${NC}: $name"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC}: $name"
        echo "  期待: $expected"
        echo "  実際: ${RESPONSE:0:200}..."
        FAILED=$((FAILED + 1))
    fi
}

# エラーテスト（HTTPエラーが返されることを確認）
run_error_test() {
    local name="$1"
    local url="$2"
    local data="$3"

    # エラー出力をキャプチャ
    RESPONSE=$(docker compose exec -T str wget -qO- --post-data="$data" "http://localhost:8080$url" 2>&1 || true)

    if echo "$RESPONSE" | grep -qE "(error|Error|400|Bad Request)"; then
        echo -e "${GREEN}✓ PASS${NC}: $name"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC}: $name"
        echo "  期待: HTTPエラー"
        echo "  実際: ${RESPONSE:0:200}..."
        FAILED=$((FAILED + 1))
    fi
}

# ================================================
echo -e "${BLUE}--- テスト1: ヘルスチェック ---${NC}"
run_test "ヘルスチェック応答" "/health" "GET" "" '"status":"ok"'
echo ""

# ================================================
echo -e "${BLUE}--- テスト2: トップページ ---${NC}"
run_test "トップページ表示" "/" "GET" "" "トラック運賃"
run_test "HTMXスクリプト読み込み" "/" "GET" "" "htmx.org"
echo ""

# ================================================
echo -e "${BLUE}--- テスト3: ルート取得API ---${NC}"
run_test "ルート取得成功" "/api/route?origin=東京&dest=大阪" "GET" "" '"success":true'
run_test "出発地未指定エラー" "/api/route?dest=大阪" "GET" "" '"success":false'
run_test "目的地未指定エラー" "/api/route?origin=東京" "GET" "" '"success":false'
run_test "同一地点エラー" "/api/route?origin=東京&dest=東京" "GET" "" '"success":false'
echo ""

# ================================================
echo -e "${BLUE}--- テスト4: 運賃計算API（JSON） ---${NC}"

# 基本計算テスト
run_test "運賃計算成功" "/api/fare/calculate/json" "POST" \
    "region_code=3&vehicle_code=3&distance_km=100&driving_minutes=120&loading_minutes=60" \
    "DistanceFare"

run_test "時間制運賃含む" "/api/fare/calculate/json" "POST" \
    "region_code=3&vehicle_code=3&distance_km=100&driving_minutes=120&loading_minutes=60" \
    "TimeFare"

run_test "赤帽運賃含む" "/api/fare/calculate/json" "POST" \
    "region_code=3&vehicle_code=3&distance_km=100&driving_minutes=120&loading_minutes=60" \
    "AkabouDistanceResult"
echo ""

# ================================================
echo -e "${BLUE}--- テスト5: 各運輸局テスト ---${NC}"
for region in 1 2 3 4 5 6 7 8 9 10; do
    run_test "運輸局$region" "/api/fare/calculate/json" "POST" \
        "region_code=$region&vehicle_code=3&distance_km=100&driving_minutes=120" \
        "DistanceFare"
done
echo ""

# ================================================
echo -e "${BLUE}--- テスト6: 各車格テスト ---${NC}"
VEHICLE_NAMES=("" "小型車(2t)" "中型車(4t)" "大型車(10t)" "トレーラー(20t)")
for vehicle in 1 2 3 4; do
    run_test "車格$vehicle(${VEHICLE_NAMES[$vehicle]})" "/api/fare/calculate/json" "POST" \
        "region_code=3&vehicle_code=$vehicle&distance_km=100&driving_minutes=120" \
        "DistanceFare"
done
echo ""

# ================================================
echo -e "${BLUE}--- テスト7: バリデーションテスト ---${NC}"
run_error_test "距離0エラー" "/api/fare/calculate/json" \
    "region_code=3&vehicle_code=3&distance_km=0&driving_minutes=120"

run_error_test "不正運輸局コードエラー" "/api/fare/calculate/json" \
    "region_code=99&vehicle_code=3&distance_km=100&driving_minutes=120"

run_error_test "不正車格コードエラー" "/api/fare/calculate/json" \
    "region_code=3&vehicle_code=99&distance_km=100&driving_minutes=120"

run_error_test "走行時間0エラー" "/api/fare/calculate/json" \
    "region_code=3&vehicle_code=3&distance_km=100&driving_minutes=0"
echo ""

# ================================================
echo -e "${BLUE}--- テスト8: 運賃計算API（HTML/HTMX用） ---${NC}"
run_test "HTML形式レスポンス" "/api/fare/calculate" "POST" \
    "region_code=3&vehicle_code=3&distance_km=100&driving_minutes=120" \
    "<"
echo ""

# ================================================
echo -e "${BLUE}--- テスト9: 高速道路ページ ---${NC}"
run_test "高速道路ページ表示" "/highway" "GET" "" "highway"
echo ""

# ================================================
echo -e "${BLUE}--- テスト10: 様々な距離でのテスト ---${NC}"
for km in 10 50 100 200 500; do
    run_test "距離${km}km" "/api/fare/calculate/json" "POST" \
        "region_code=3&vehicle_code=3&distance_km=$km&driving_minutes=120" \
        "DistanceFare"
done
echo ""

# ================================================
# nginx経由のテスト（Basic認証確認）
echo -e "${BLUE}--- テスト11: nginx経由アクセス ---${NC}"

# ヘルスチェックは認証不要
RESPONSE=$(curl -s http://localhost:80/health 2>/dev/null || echo "ERROR")
if echo "$RESPONSE" | grep -q '"status":"ok"'; then
    echo -e "${GREEN}✓ PASS${NC}: nginx経由ヘルスチェック（認証不要）"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: nginx経由ヘルスチェック"
    FAILED=$((FAILED + 1))
fi

# 認証なしでルートアクセス → 401
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:80/ 2>/dev/null || echo "000")
if [ "$HTTP_CODE" = "401" ]; then
    echo -e "${GREEN}✓ PASS${NC}: nginx認証要求確認 (HTTP 401)"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: nginx認証要求確認 (HTTP $HTTP_CODE)"
    FAILED=$((FAILED + 1))
fi
echo ""

# ================================================
# 結果サマリー
# ================================================
echo "========================================"
echo "テスト結果サマリー"
echo "========================================"
echo -e "${GREEN}PASSED${NC}: $PASSED"
echo -e "${RED}FAILED${NC}: $FAILED"
TOTAL=$((PASSED + FAILED))
echo "TOTAL : $TOTAL"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ 全テスト成功！${NC}"
    exit 0
else
    echo -e "${RED}✗ ${FAILED}件のテストが失敗${NC}"
    exit 1
fi
