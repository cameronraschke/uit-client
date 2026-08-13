package requests

const (
	hwmonDirRoot     = "/sys/class/hwmon/"
	powercapRootDir  = "/sys/class/powercap/"
	powercapRootDir2 = "/sys/class/powercap/intel-rapl/"
	netIfRootDir     = "/sys/class/net/"
)

type LiveDataRequest struct {
	Hardware   *HardwareDataRequest `json:"hardware"`
	Job        *JobQueueDataRequest `json:"job"`
	Status     *AppStatusRequest    `json:"status"`
	Screenshot []byte               `json:"live_screenshot"`
}

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
