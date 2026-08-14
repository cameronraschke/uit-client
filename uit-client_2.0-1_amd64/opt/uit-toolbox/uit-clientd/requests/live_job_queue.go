package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"sync"
)

type ClientJobQueueDataResponse struct {
	CloneMode     string `json:"clone_mode"`
	DiskImageName string `json:"disk_image_name"`
	EraseMode     string `json:"erase_mode"`
	IsQueued      *bool  `json:"is_queued"`
	IsRunning     *bool  `json:"is_running"`
	Name          string `json:"name"`
	NameFormatted string `json:"name_formatted"`
	QueuePosition *int64 `json:"queue_position"`
	Status        string `json:"status"`
}

var (
	jobQueueDataBuf   bytes.Buffer
	jobQueueDataBufMu sync.Mutex
)

func GetJobQueueData(ctx context.Context, tag int64) (ClientJobQueueDataResponse, error) {
	jq := ClientJobQueueDataResponse{}

	q := url.Values{}
	q.Set("tagnumber", strconv.FormatInt(tag, 10))

	jobQueueDataBufMu.Lock()
	defer jobQueueDataBufMu.Unlock()

	jobQueueDataBuf.Reset()
	jobQueueDataBuf, err := getRequest(
		ctx,
		url.URL{
			Path:     "/api/v2/app/live/job",
			RawQuery: q.Encode(),
		})
	if err != nil {
		return ClientJobQueueDataResponse{}, err
	}

	if err := json.Unmarshal(jobQueueDataBuf, &jq); err != nil {
		return ClientJobQueueDataResponse{}, fmt.Errorf("cannot unmarshal ClientJobQueueDataResponse JSON (GetJobQueueData): %v", err)
	}

	return jq, err
}
