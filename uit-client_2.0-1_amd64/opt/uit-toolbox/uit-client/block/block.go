package block

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/sys/unix"
)

type BlockDevice struct {
	Name            string
	Path            string
	Major           uint32
	Minor           uint32
	BlockDeviceType string
}

func ListBlockDevices(devDir string) ([]*BlockDevice, error) {
	fd, err := unix.Open(devDir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open device directory: %v", err)
	}
	defer unix.Close(fd)

	directoryListBuffer := make([]byte, 1<<13) // 8kb buffer to load in directory entries
	var devices []*BlockDevice

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

		scsiOrSataMajors := []uint32{8, 65, 66, 67, 68, 69, 70, 71, 128, 129, 130, 131, 132, 133, 134, 135}
		scsiIdeMajors := []uint32{3, 22, 33, 34, 56, 57, 88, 89, 90, 91}

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

			if majorNum == 0 && minorNum == 0 {
				continue
			}

			diskType := "unknown"

			if slices.Contains(scsiOrSataMajors, majorNum) && strings.HasPrefix(devicePath, "/dev/sd") {
				diskType = "SCSI/SATA"
			} else if majorNum == 9 {
				if strings.HasPrefix(devicePath, "/dev/md") {
					diskType = "MD RAID"
				}
				if strings.HasPrefix(devicePath, "/dev/st") ||
					strings.HasPrefix(devicePath, "/dev/nst") {
					diskType = "SCSI Tape"
				}
			} else if majorNum == 11 {
				diskType = "SCSI CD-ROM"
			} else if majorNum == 21 && strings.HasPrefix(devicePath, "/dev/sg") {
				diskType = "Generic SCSI"
			} else if slices.Contains(scsiIdeMajors, majorNum) && strings.HasPrefix(devicePath, "/dev/hd") {
				diskType = "SCSI IDE/CD-ROM"
			} else if majorNum == 259 && strings.HasPrefix(devicePath, "/dev/nvme") {
				diskType = "NVMe"
			} else {
				continue
			}

			devices = append(devices, &BlockDevice{
				Name:            deviceName,
				Path:            devicePath,
				Major:           majorNum,
				Minor:           minorNum,
				BlockDeviceType: diskType,
			})
		}
	}
	fmt.Printf("Total block devices found: %d\n", len(devices))
	return devices, nil
}
