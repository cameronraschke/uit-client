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

