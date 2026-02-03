package service

import "testing"

func TestGeocodingClient_GetPrefecture(t *testing.T) {
	// モッククライアントを使用
	client := NewMockGeocodingClient()

	tests := []struct {
		name           string
		address        string
		wantPrefecture string
		wantErr        bool
	}{
		{
			name:           "東京の住所",
			address:        "東京都千代田区丸の内1-1-1",
			wantPrefecture: "東京都",
			wantErr:        false,
		},
		{
			name:           "大阪の住所",
			address:        "大阪府大阪市中央区心斎橋",
			wantPrefecture: "大阪府",
			wantErr:        false,
		},
		{
			name:           "北海道の住所",
			address:        "北海道札幌市中央区",
			wantPrefecture: "北海道",
			wantErr:        false,
		},
		{
			name:           "沖縄の住所",
			address:        "沖縄県那覇市",
			wantPrefecture: "沖縄県",
			wantErr:        false,
		},
		{
			name:           "空文字列",
			address:        "",
			wantPrefecture: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrefecture, err := client.GetPrefecture(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPrefecture(%q) error = %v, wantErr %v", tt.address, err, tt.wantErr)
				return
			}
			if gotPrefecture != tt.wantPrefecture {
				t.Errorf("GetPrefecture(%q) = %v, want %v", tt.address, gotPrefecture, tt.wantPrefecture)
			}
		})
	}
}

func TestGeocodingClient_GetAddressComponents(t *testing.T) {
	client := NewMockGeocodingClient()

	tests := []struct {
		name           string
		address        string
		wantPrefecture string
		wantCity       string
		wantErr        bool
	}{
		{
			name:           "東京都千代田区",
			address:        "東京都千代田区丸の内1-1-1",
			wantPrefecture: "東京都",
			wantCity:       "千代田区",
			wantErr:        false,
		},
		{
			name:           "大阪市",
			address:        "大阪府大阪市中央区心斎橋",
			wantPrefecture: "大阪府",
			wantCity:       "大阪市",
			wantErr:        false,
		},
		{
			name:           "札幌市",
			address:        "北海道札幌市中央区",
			wantPrefecture: "北海道",
			wantCity:       "札幌市",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components, err := client.GetAddressComponents(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressComponents(%q) error = %v, wantErr %v", tt.address, err, tt.wantErr)
				return
			}
			if components.Prefecture != tt.wantPrefecture {
				t.Errorf("GetAddressComponents(%q).Prefecture = %v, want %v", tt.address, components.Prefecture, tt.wantPrefecture)
			}
			if components.City != tt.wantCity {
				t.Errorf("GetAddressComponents(%q).City = %v, want %v", tt.address, components.City, tt.wantCity)
			}
		})
	}
}

func TestExtractPrefectureFromAddress(t *testing.T) {
	tests := []struct {
		name           string
		address        string
		wantPrefecture string
		wantOk         bool
	}{
		{"東京都", "東京都千代田区丸の内", "東京都", true},
		{"大阪府", "大阪府大阪市中央区", "大阪府", true},
		{"北海道", "北海道札幌市中央区", "北海道", true},
		{"京都府", "京都府京都市", "京都府", true},
		{"神奈川県", "神奈川県横浜市", "神奈川県", true},
		{"沖縄県", "沖縄県那覇市", "沖縄県", true},
		{"都道府県なし", "千代田区丸の内", "", false},
		{"空文字列", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrefecture, gotOk := ExtractPrefectureFromAddress(tt.address)
			if gotOk != tt.wantOk {
				t.Errorf("ExtractPrefectureFromAddress(%q) ok = %v, want %v", tt.address, gotOk, tt.wantOk)
				return
			}
			if gotPrefecture != tt.wantPrefecture {
				t.Errorf("ExtractPrefectureFromAddress(%q) = %v, want %v", tt.address, gotPrefecture, tt.wantPrefecture)
			}
		})
	}
}
