package requests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"strconv"
	"sync"
)

type JobQueueDataRequest struct {
	CloneMode     string `json:"clone_mode"`
	EraseMode     string `json:"erase_mode"`
	IsQueued      *bool  `json:"is_queued"`
	IsRunning     *bool  `json:"is_running"`
	Name          string `json:"name"`
	NameFormatted string `json:"name_formatted"`
	DiskImageName string `json:"disk_image_name"`
	QueuePosition string `json:"queue_position"`
}

var (
	jobQueueDataBuf   bytes.Buffer
	jobQueueDataBufMu sync.Mutex
)

func getJobQueueData(tag int64) (JobQueueDataRequest, error) {
	jq := JobQueueDataRequest{}

	q := url.Values{}
	q.Set("tagnumber", strconv.FormatInt(tag, 10))
	resp, err := getRequest(url.URL{
		Path:     "/api/v2/app/live/job",
		RawQuery: q.Encode(),
	})
	if err != nil {
		return JobQueueDataRequest{}, err
	}

	jobQueueDataBufMu.Lock()
	defer jobQueueDataBufMu.Unlock()

	jobQueueDataBuf.Reset()

	jobQueueDataBuf, err := io.ReadAll(resp)
	json.Unmarshal(jobQueueDataBuf, &jq)

	return jq, err
}
