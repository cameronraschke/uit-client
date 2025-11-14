package block

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/unix"
)

type BlockDevice struct {
	Name  string
	Path  string
	Major uint32
	Minor uint32
}

func ListBlockDevices(devDir string) ([]BlockDevice, error) {
	fd, err := unix.Open(devDir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open device directory: %v", err)
	}
	defer unix.Close(fd)

	directoryListBuffer := make([]byte, 1<<13) // 8kb buffer to load in directory entries
	var devices []BlockDevice

	for {
		directoryEntry, err := unix.ReadDirent(fd, directoryListBuffer)
		if err != nil {
			return nil, err
		}
		if directoryEntry <= 0 {
			break
		}

		deviceNames := make([]string, 0, 128)
		_, _, deviceNames = unix.ParseDirent(directoryListBuffer[:directoryEntry], -1, deviceNames)

		for _, deviceName := range deviceNames {
			if deviceName == "." || deviceName == ".." {
				continue
			}
			devicePath := filepath.Join(devDir, deviceName)

			var statData unix.Stat_t
			if err := unix.Lstat(devicePath, &statData); err != nil {
				continue
			}
			// Only select block devices - bitwise and to mask file type bits
			if statData.Mode&unix.S_IFMT != unix.S_IFBLK {
				continue
			}

			majorNum := unix.Major(uint64(statData.Rdev))
			minorNum := unix.Minor(uint64(statData.Rdev))

			devices = append(devices, BlockDevice{
				Name:  deviceName,
				Path:  devicePath,
				Major: majorNum,
				Minor: minorNum,
			})
		}
	}
	return devices, nil
}
