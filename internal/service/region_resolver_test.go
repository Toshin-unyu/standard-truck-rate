package service

import "testing"

func TestResolveRegionCode(t *testing.T) {
	tests := []struct {
		name       string
		prefecture string
		wantCode   int
		wantErr    bool
	}{
		// 1: 北海道
		{"北海道", "北海道", 1, false},

		// 2: 東北
		{"青森県", "青森県", 2, false},
		{"岩手県", "岩手県", 2, false},
		{"宮城県", "宮城県", 2, false},
		{"秋田県", "秋田県", 2, false},
		{"山形県", "山形県", 2, false},
		{"福島県", "福島県", 2, false},

		// 3: 関東
		{"茨城県", "茨城県", 3, false},
		{"栃木県", "栃木県", 3, false},
		{"群馬県", "群馬県", 3, false},
		{"埼玉県", "埼玉県", 3, false},
		{"千葉県", "千葉県", 3, false},
		{"東京都", "東京都", 3, false},
		{"神奈川県", "神奈川県", 3, false},
		{"山梨県", "山梨県", 3, false},

		// 4: 北陸信越
		{"新潟県", "新潟県", 4, false},
		{"富山県", "富山県", 4, false},
		{"石川県", "石川県", 4, false},
		{"長野県", "長野県", 4, false},

		// 5: 中部
		{"福井県", "福井県", 5, false},
		{"岐阜県", "岐阜県", 5, false},
		{"静岡県", "静岡県", 5, false},
		{"愛知県", "愛知県", 5, false},
		{"三重県", "三重県", 5, false},

		// 6: 近畿
		{"滋賀県", "滋賀県", 6, false},
		{"京都府", "京都府", 6, false},
		{"大阪府", "大阪府", 6, false},
		{"兵庫県", "兵庫県", 6, false},
		{"奈良県", "奈良県", 6, false},
		{"和歌山県", "和歌山県", 6, false},

		// 7: 中国
		{"鳥取県", "鳥取県", 7, false},
		{"島根県", "島根県", 7, false},
		{"岡山県", "岡山県", 7, false},
		{"広島県", "広島県", 7, false},
		{"山口県", "山口県", 7, false},

		// 8: 四国
		{"徳島県", "徳島県", 8, false},
		{"香川県", "香川県", 8, false},
		{"愛媛県", "愛媛県", 8, false},
		{"高知県", "高知県", 8, false},

		// 9: 九州
		{"福岡県", "福岡県", 9, false},
		{"佐賀県", "佐賀県", 9, false},
		{"長崎県", "長崎県", 9, false},
		{"熊本県", "熊本県", 9, false},
		{"大分県", "大分県", 9, false},
		{"宮崎県", "宮崎県", 9, false},
		{"鹿児島県", "鹿児島県", 9, false},

		// 10: 沖縄
		{"沖縄県", "沖縄県", 10, false},

		// エラーケース
		{"空文字列", "", 0, true},
		{"不明な都道府県", "ハワイ", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, err := ResolveRegionCode(tt.prefecture)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveRegionCode(%q) error = %v, wantErr %v", tt.prefecture, err, tt.wantErr)
				return
			}
			if gotCode != tt.wantCode {
				t.Errorf("ResolveRegionCode(%q) = %v, want %v", tt.prefecture, gotCode, tt.wantCode)
			}
		})
	}
}

func TestResolveRegionName(t *testing.T) {
	tests := []struct {
		name       string
		prefecture string
		wantName   string
		wantErr    bool
	}{
		{"北海道", "北海道", "北海道", false},
		{"東京都", "東京都", "関東", false},
		{"大阪府", "大阪府", "近畿", false},
		{"沖縄県", "沖縄県", "沖縄", false},
		{"空文字列", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, err := ResolveRegionName(tt.prefecture)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveRegionName(%q) error = %v, wantErr %v", tt.prefecture, err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("ResolveRegionName(%q) = %v, want %v", tt.prefecture, gotName, tt.wantName)
			}
		})
	}
}

func TestResolveAkabouArea(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		wantArea string
	}{
		{"東京23区_千代田区", "東京都千代田区丸の内", "東京23区"},
		{"東京23区_新宿区", "東京都新宿区西新宿", "東京23区"},
		{"東京23区_渋谷区", "東京都渋谷区渋谷", "東京23区"},
		{"大阪市内_中央区", "大阪府大阪市中央区", "大阪市内"},
		{"大阪市内_北区", "大阪府大阪市北区梅田", "大阪市内"},
		{"その他地域_横浜", "神奈川県横浜市", ""},
		{"その他地域_名古屋", "愛知県名古屋市", ""},
		{"その他地域_八王子", "東京都八王子市", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArea := ResolveAkabouArea(tt.address)
			if gotArea != tt.wantArea {
				t.Errorf("ResolveAkabouArea(%q) = %v, want %v", tt.address, gotArea, tt.wantArea)
			}
		})
	}
}
