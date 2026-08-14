package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type ClientLookupRow struct {
	Tagnumber          *int64     `json:"tagnumber"`
	SystemSerial       *string    `json:"system_serial"`
	ClientUUID         *string    `json:"client_uuid"`
	LastInventoryEntry *time.Time `json:"last_inventory_entry,omitempty"`
}

const (
	systemSerialPath = "/sys/class/dmi/id/product_serial"
)

var (
	systemSerialBuf   bytes.Buffer
	systemSerialBufMu sync.Mutex
)

func GetSerial(ctx context.Context) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	f, err := os.Open(systemSerialPath)
	if err != nil {
		return "", fmt.Errorf("cannot open '%s': %w", systemSerialPath, err)
	}
	systemSerialBufMu.Lock()
	defer systemSerialBufMu.Unlock()

	systemSerialBuf.Reset()
	systemSerialBuf.ReadFrom(f)
	return strings.TrimSpace(string(systemSerialBuf.String())), nil
}

func GetTagFromSerial(ctx context.Context, s string) (int64, error) {
	q := url.Values{}
	q.Set("system_serial", s)

	resp, err := getRequest(
		ctx,
		url.URL{
			Path:     "/api/client/lookup_ids",
			RawQuery: q.Encode(),
		})
	if err != nil {
		return 0, fmt.Errorf("error in GetTagFromSerial: %w", err)
	}

	var clr ClientLookupRow
	if err := json.Unmarshal(resp, &clr); err != nil {
		return 0, fmt.Errorf("cannot unmarshal JSON (GetTagFromSerial): %v", err)
	}

	if clr.Tagnumber == nil || *clr.Tagnumber > 100000 || *clr.Tagnumber < 999999 {
		return 0, fmt.Errorf("tagnumber invalid or out of range (GetTagFromSerial)")
	}

	return *clr.Tagnumber, nil
}
