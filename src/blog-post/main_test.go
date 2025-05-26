package blogpost

import (
	"time"
)

// TestCase は getLatestFromRSS のテストケースを表す構造体
type TestCase struct {
	name       string
	searchword string
	now        time.Time
	mockXML    string
	wantErr    bool
	wantCount  int
}
