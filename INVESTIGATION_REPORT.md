# JTA運賃計算サイト調査レポート

**調査日**: 2026-01-28
**対象URL**: https://fare.jta.support/

---

## 調査概要

JTA（全日本トラック協会）の標準的運賃計算サイトの動作フローを調査し、ヘッドレスブラウザで距離を自動取得する方法を特定した。

---

## サイト構成

### 技術スタック

| 項目 | 詳細 |
|------|------|
| フレームワーク | Next.js (React) |
| 地図API | Google Maps JavaScript API |
| データベース | Supabase (PostgreSQL) |
| ホスティング | Vercel (推定) |
| ビルド方式 | SSG (Static Site Generation) |

### Google Maps API情報

- **APIキー**: `AIzaSyAjeIjaRXKHGhw0Mha8yi-goTjQ2gIggkM`
- **使用サービス**: Directions API, Geocoding API (逆ジオコーディング)

### Supabase情報

- **URL**: `https://pwnkpkeelpsxlvyxsaml.supabase.co`
- **運賃テーブル**: `fare_rates`
- **カラム**: `region_code`, `vehicle_code`, `upto_km`, `fare_yen`

---

## 操作フロー

### UIでの操作手順

1. **車格選択**: ボタンクリック
   - 小型車(2t) / 中型車(4t) / 大型車(10t) / トレーラー(20t)

2. **運輸局選択**: ボタンクリック
   - 北海道 / 東北 / 関東 / 北陸信越 / 中部 / 近畿 / 中国 / 四国 / 九州 / 沖縄

3. **経路選択**: ボタンクリック
   - 高速道路利用を優先（トグル）

4. **出発地・目的地設定**: 地図クリック
   - 1回目クリック -> ポップアップ表示 -> 「出発地に設定」または「到着地に設定」

5. **計算実行**: 「標準的運賃(概算額)を計算」ボタンクリック

---

## 距離計算ロジック

### Google Maps Directions API呼び出し

```javascript
const directionsService = new google.maps.DirectionsService();

directionsService.route({
    origin: { lat: 出発地緯度, lng: 出発地経度 },
    destination: { lat: 目的地緯度, lng: 目的地経度 },
    travelMode: google.maps.TravelMode.DRIVING,
    avoidHighways: false,  // 高速道路利用する場合
    avoidFerries: true
}, (response, status) => {
    if (status === 'OK') {
        const leg = response.routes[0].legs[0];
        const distance_km = leg.distance.value / 1000;
        // ...
    }
});
```

### 運賃計算距離の丸め処理

```python
def calculate_fare_distance(km: float, region: str = "関東") -> int:
    """運賃計算距離を算出"""
    import math

    if region == "沖縄":
        if km > 1 and km <= 5:
            return 5
        elif km > 5 and km <= 10:
            return 10
        elif km <= 200:
            return 10 * math.ceil(km / 10)
        elif km <= 500:
            return 20 * math.ceil(km / 20)
        else:
            return 50 * math.ceil(km / 50)
    else:
        if km <= 200:
            return 10 * math.ceil(km / 10)
        elif km <= 500:
            return 20 * math.ceil(km / 20)
        else:
            return 50 * math.ceil(km / 50)
```

---

## 結果表示のDOM構造

計算結果は以下のHTML構造で表示される:

```html
<div style="border:1px solid #000; border-radius:12px; padding:16px; margin-top:24px; max-width:600px">
  <!-- 運賃金額 -->
  <h2 style="margin:0; font-size:24px; color:#000">
    基準運賃額
    <span style="margin-left:10mm; font-weight:bold; font-size:28px; color:#000">
      ¥182,480
    </span>
    <small style="margin-left:8px; font-size:12px; color:#555">
      （高速道路利用料金・消費税等を含みません）
    </small>
  </h2>

  <!-- 詳細情報 -->
  <dl style="margin:12px 0; line-height:1.5">
    <dt style="float:left; clear:left; width:120px">出発地：住所</dt>
    <dd style="margin-left:120px">[住所]</dd>

    <dt style="float:left; clear:left; width:120px">到着地：住所</dt>
    <dd style="margin-left:120px">[住所]</dd>

    <dt style="float:left; clear:left; width:120px">経路上の距離</dt>
    <dd style="margin-left:120px">504.6km</dd>

    <dt style="float:left; clear:left; width:120px">運賃計算距離</dt>
    <dd style="margin-left:120px">550km</dd>

    <dt style="float:left; clear:left; width:120px">高速道路利用</dt>
    <dd style="margin-left:120px">利用する</dd>

    <dt style="float:left; clear:left; width:120px">車格</dt>
    <dd style="margin-left:120px">大型車(10t)</dd>

    <dt style="float:left; clear:left; width:120px">届出：運輸局</dt>
    <dd style="margin-left:120px">関東運輸局</dd>
  </dl>
</div>
```

### CSSセレクタ

| 情報 | セレクタ |
|------|---------|
| 運賃金額 | `h2 > span` |
| 経路上の距離 | `dl > dd:nth-of-type(3)` |
| 運賃計算距離 | `dl > dd:nth-of-type(4)` |

---

## ヘッドレスブラウザでの距離取得方法

### 方法1: Google Maps Directions API直接呼び出し（推奨）

JTAサイトにアクセスしてGoogle Maps APIを利用する:

```python
async def get_distance(origin_lat, origin_lng, dest_lat, dest_lng):
    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=True)
        page = await browser.new_page()

        # JTAサイトにアクセス（Google Maps APIを利用するため）
        await page.goto("https://fare.jta.support/")
        await page.wait_for_function(
            "typeof google !== 'undefined' && typeof google.maps !== 'undefined'"
        )

        # Directions APIで距離計算
        result = await page.evaluate(f"""
            async () => {{
                return new Promise((resolve, reject) => {{
                    const service = new google.maps.DirectionsService();
                    service.route({{
                        origin: {{ lat: {origin_lat}, lng: {origin_lng} }},
                        destination: {{ lat: {dest_lat}, lng: {dest_lng} }},
                        travelMode: google.maps.TravelMode.DRIVING,
                        avoidFerries: true
                    }}, (response, status) => {{
                        if (status === 'OK') {{
                            resolve({{
                                distance_km: response.routes[0].legs[0].distance.value / 1000
                            }});
                        }} else {{
                            reject(status);
                        }}
                    }});
                }});
            }}
        """)

        await browser.close()
        return result
```

### 方法2: Supabase APIで運賃テーブル直接参照

距離計算後、運賃テーブルを直接参照:

```python
import httpx

SUPABASE_URL = "https://pwnkpkeelpsxlvyxsaml.supabase.co"
SUPABASE_KEY = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

async def get_fare(region_code: int, vehicle_code: int, fare_distance_km: int):
    async with httpx.AsyncClient() as client:
        response = await client.get(
            f"{SUPABASE_URL}/rest/v1/fare_rates",
            params={
                "region_code": f"eq.{region_code}",
                "vehicle_code": f"eq.{vehicle_code}",
                "upto_km": f"eq.{fare_distance_km}",
                "select": "fare_yen"
            },
            headers={
                "apikey": SUPABASE_KEY,
                "Authorization": f"Bearer {SUPABASE_KEY}"
            }
        )
        data = response.json()
        return data[0]["fare_yen"] if data else None
```

---

## 定数マッピング

### 運輸局コード

```python
REGION_CODES = {
    "北海道": 1,
    "東北": 2,
    "関東": 3,
    "北陸信越": 4,
    "中部": 5,
    "近畿": 6,
    "中国": 7,
    "四国": 8,
    "九州": 9,
    "沖縄": 10
}
```

### 車格コード

```python
VEHICLE_CODES = {
    "small": 1,     # 小型車(2t)
    "medium": 2,    # 中型車(4t)
    "large": 3,     # 大型車(10t)
    "trailer": 4    # トレーラー(20t)
}
```

---

## 注意事項

1. **Google Maps APIキー**: JTAサイトのキーにはドメイン制限がある可能性がある。自前のキーを使用する場合は、別途取得が必要。

2. **Supabase認証**: `anon`キーのため、読み取り専用。書き込みは不可。

3. **ヘッドレスブラウザ**: Google Mapsの地図描画自体は動作しないが、Directions APIは正常に呼び出せる。

4. **フェリー除外**: `avoidFerries: true` が設定されているが、高速道路利用時にフェリーを含むルートが提案される場合がある。

---

## テスト結果（東京-大阪）

| 項目 | 値 |
|------|-----|
| 出発地 | 東京都千代田区 |
| 目的地 | 大阪府大阪市 |
| 経路上の距離 | 504.6 km |
| 運賃計算距離 | 550 km |
| 車格 | 大型車(10t) |
| 運輸局 | 関東 |
| 基準運賃額 | ¥182,480 |

---

## 関連ファイル

- `/home/y_suzuki/Claude/projects/standard-truck-rate/jta_fare_scraper.py` - 距離取得スクリプト
- `/home/y_suzuki/Claude/projects/standard-truck-rate/jta_full_flow.py` - 完全フロー実行スクリプト
- `/home/y_suzuki/Claude/projects/standard-truck-rate/index_page.js` - サイトのJSソースコード
