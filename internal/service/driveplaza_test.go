package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

func TestDrivePlazaClient_ParseICListXML(t *testing.T) {
	// モックXMLレスポンス
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<NexcoIC>
  <IcItem>
    <Code>1010001</Code>
    <Name>東京</Name>
    <Yomi>とうきょう</Yomi>
    <Type>1</Type>
    <RoadNo>1010</RoadNo>
    <RoadName>【E1】東名高速道路</RoadName>
  </IcItem>
  <IcItem>
    <Code>1010002</Code>
    <Name>用賀</Name>
    <Yomi>ようが</Yomi>
    <Type>1</Type>
    <RoadNo>1010</RoadNo>
    <RoadName>【E1】東名高速道路</RoadName>
  </IcItem>
</NexcoIC>`

	ics, err := ParseICListXML([]byte(xmlData))
	if err != nil {
		t.Fatalf("ParseICListXML failed: %v", err)
	}

	if len(ics) != 2 {
		t.Fatalf("件数: 期待=2, 実際=%d", len(ics))
	}

	// 1件目確認
	if ics[0].Code != "1010001" {
		t.Errorf("Code: 期待=1010001, 実際=%s", ics[0].Code)
	}
	if ics[0].Name != "東京" {
		t.Errorf("Name: 期待=東京, 実際=%s", ics[0].Name)
	}
	if ics[0].Yomi != "とうきょう" {
		t.Errorf("Yomi: 期待=とうきょう, 実際=%s", ics[0].Yomi)
	}
	if ics[0].Type != 1 {
		t.Errorf("Type: 期待=1, 実際=%d", ics[0].Type)
	}
	if ics[0].RoadName != "【E1】東名高速道路" {
		t.Errorf("RoadName: 期待=【E1】東名高速道路, 実際=%s", ics[0].RoadName)
	}
}

func TestDrivePlazaClient_FetchICList(t *testing.T) {
	// モックサーバー
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<NexcoIC>
  <IcItem>
    <Code>1010001</Code>
    <Name>東京</Name>
    <Yomi>とうきょう</Yomi>
    <Type>1</Type>
    <RoadNo>1010</RoadNo>
    <RoadName>【E1】東名高速道路</RoadName>
  </IcItem>
</NexcoIC>`
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(xml))
	}))
	defer server.Close()

	client := NewDrivePlazaClient()
	client.icSearchURL = server.URL

	ics, err := client.FetchICList()
	if err != nil {
		t.Fatalf("FetchICList failed: %v", err)
	}

	if len(ics) != 1 {
		t.Errorf("件数: 期待=1, 実際=%d", len(ics))
	}
}

func TestDrivePlazaClient_ParseTollHTML(t *testing.T) {
	// モックHTMLレスポンス（実際のドラぷらのHTML構造を模倣）
	html := `<!DOCTYPE html>
<html>
<body>
<div class="price-wrap">
	<dl class="li-price">
		<dt>通常料金</dt>
		<dd><em>8,350</em>円</dd>
	</dl>
	<dl class="li-price">
		<dt>ETC料金</dt>
		<dd><em><span id="fee_etc1">5,840</span></em>円</dd>
	</dl>
	<dl class="li-price">
		<dt>ETC2.0料金</dt>
		<dd><em><span id="fee_etc21">5,840</span></em>円</dd>
	</dl>
</div>
<div class="times-wrap">
	<dl class="li-normal">
		<dt>通常時間</dt>
		<dd>3時間30分</dd>
	</dl>
	<dl class="li-distance">
		<dt>距離</dt>
		<dd>325.5km</dd>
	</dl>
</div>
</body>
</html>`

	toll, err := ParseTollHTML([]byte(html))
	if err != nil {
		t.Fatalf("ParseTollHTML failed: %v", err)
	}

	if toll.NormalToll != 8350 {
		t.Errorf("NormalToll: 期待=8350, 実際=%d", toll.NormalToll)
	}
	if toll.EtcToll != 5840 {
		t.Errorf("EtcToll: 期待=5840, 実際=%d", toll.EtcToll)
	}
	if toll.DistanceKm != 325.5 {
		t.Errorf("DistanceKm: 期待=325.5, 実際=%f", toll.DistanceKm)
	}
	if toll.DurationMin != 210 {
		t.Errorf("DurationMin: 期待=210, 実際=%d", toll.DurationMin)
	}
}

func TestDrivePlazaClient_FetchToll(t *testing.T) {
	// モックサーバー
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html>
<body>
<div class="price-wrap">
	<dl class="li-price">
		<dt>通常料金</dt>
		<dd><em>8,350</em>円</dd>
	</dl>
	<dl class="li-price">
		<dt>ETC料金</dt>
		<dd><em><span id="fee_etc1">5,840</span></em>円</dd>
	</dl>
	<dl class="li-price">
		<dt>ETC2.0料金</dt>
		<dd><em><span id="fee_etc21">5,840</span></em>円</dd>
	</dl>
</div>
<div class="times-wrap">
	<dl class="li-normal">
		<dt>通常時間</dt>
		<dd>3時間30分</dd>
	</dl>
	<dl class="li-distance">
		<dt>距離</dt>
		<dd>325.5km</dd>
	</dl>
</div>
</body>
</html>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	client := NewDrivePlazaClient()
	client.tollSearchURL = server.URL

	toll, err := client.FetchToll("東京", "名古屋", model.CarTypeLarge)
	if err != nil {
		t.Fatalf("FetchToll failed: %v", err)
	}

	if toll.NormalToll != 8350 {
		t.Errorf("NormalToll: 期待=8350, 実際=%d", toll.NormalToll)
	}
}

func TestParseTollAmount(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"8,350円", 8350},
		{"5,840円", 5840},
		{"12,340円", 12340},
		{"100円", 100},
		{"0円", 0},
	}

	for _, tt := range tests {
		got := parseTollAmount(tt.input)
		if got != tt.expected {
			t.Errorf("parseTollAmount(%q): 期待=%d, 実際=%d", tt.input, tt.expected, got)
		}
	}
}

func TestParseDistance(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"325.5km", 325.5},
		{"100km", 100.0},
		{"50.3km", 50.3},
	}

	for _, tt := range tests {
		got := parseDistance(tt.input)
		if got != tt.expected {
			t.Errorf("parseDistance(%q): 期待=%f, 実際=%f", tt.input, tt.expected, got)
		}
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"3時間30分", 210},
		{"1時間0分", 60},
		{"2時間15分", 135},
		{"30分", 30},
		{"5時間", 300},
	}

	for _, tt := range tests {
		got := parseDuration(tt.input)
		if got != tt.expected {
			t.Errorf("parseDuration(%q): 期待=%d, 実際=%d", tt.input, tt.expected, got)
		}
	}
}
