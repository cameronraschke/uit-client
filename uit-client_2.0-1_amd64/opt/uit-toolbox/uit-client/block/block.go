package block

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

type BlockDevice struct {
	Name                 string
	Path                 string
	Major                uint32
	Minor                uint32
	DiskType             string
	Removable            bool
	Rotating             bool
	LogicalBlockSize     uint32
	PhysicalBlockSize    uint32
	SectorCount          uint64
	CapacityMiB          float64
	Model                string
	Manufacturer         string
	Serial               string
	WWID                 string
	Firmware             string
	NvmeQualifiedName    string
	PCIeCurrentLinkSpeed string
	PCIeCurrentLinkWidth string
	PCIeMaxLinkSpeed     string
	PCIeMaxLinkWidth     string
}

func ListBlockDevices(devDir string) ([]*BlockDevice, error) {
	fd, err := unix.Open(devDir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open device directory: %v", err)
	}
	defer unix.Close(fd)

	directoryListBuffer := make([]byte, 1<<13) // 8kb buffer to load in directory entries
	var devices []*BlockDevice

	diskType := "Unknown"
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

			if minorNum != 0 {
				continue
			}

			sysBlock := filepath.Join("/sys/class/block", deviceName)
			removable := readUintBool(filepath.Join(sysBlock, "removable"))
			rotating := readUintBool(filepath.Join(sysBlock, "queue", "rotational"))
			lBlocks := readUint(filepath.Join(sysBlock, "queue", "logical_block_size"))
			logicalBlockSizeBytes := uint32(lBlocks)
			pBlocks := readUint(filepath.Join(sysBlock, "queue", "physical_block_size"))
			physicalBlockSizeBytes := uint32(pBlocks)

			sectors := readUint(filepath.Join(sysBlock, "size"))
			sizeBytes := sectors * uint64(logicalBlockSizeBytes)
			sizeMib := float64(sizeBytes) / 1024.0 / 1024.0 // Convert bytes to MiB

			diskWWID := readFileAndTrim(filepath.Join(sysBlock, "wwid"))

			deviceSymLink := filepath.Join(sysBlock, "device")
			deviceRealPath, _ := filepath.EvalSymlinks(deviceSymLink)
			if deviceRealPath == "" || deviceRealPath == "/" {
				continue
			}
			diskModel := readFileAndTrim(filepath.Join(deviceSymLink, "model"))
			diskManufacturer := readFileAndTrim(filepath.Join(deviceSymLink, "vendor"))
			diskSerial := ""
			if serial := readFileAndTrim(filepath.Join(deviceSymLink, "serial")); serial != "" {
				diskSerial = serial
			} else if eui := readFileAndTrim(filepath.Join(devicePath, "eui")); eui != "" {
				diskSerial = eui
			} else if euiFallback := readFileAndTrim(filepath.Join(deviceSymLink, "eui")); euiFallback != "" {
				diskSerial = euiFallback
			}

			diskFirmware := ""
			if firmware := readFileAndTrim(filepath.Join(deviceSymLink, "firmware_rev")); firmware != "" {
				diskFirmware = firmware
			} else if firmwareFallback := readFileAndTrim(filepath.Join(deviceSymLink, "rev")); firmwareFallback != "" {
				diskFirmware = firmwareFallback
			}

			nvmeQualifiedName := readFileAndTrim(filepath.Join(deviceSymLink, "subsysnqn"))

			// No need to check the symlink here
			nvmeControllerSubsystemPath := filepath.Join(deviceSymLink, "subsystem", "nvme0", "device")

			pcieCurrentLinkSpeed := readFileAndTrim(filepath.Join(nvmeControllerSubsystemPath, "current_link_speed"))
			pcieCurrentLinkWidth := readFileAndTrim(filepath.Join(nvmeControllerSubsystemPath, "current_link_width"))
			pcieMaxLinkSpeed := readFileAndTrim(filepath.Join(nvmeControllerSubsystemPath, "max_link_speed"))
			pcieMaxLinkWidth := readFileAndTrim(filepath.Join(nvmeControllerSubsystemPath, "max_link_width"))

			if minorNum == 0 {
				if diskType == "SCSI/SATA" && rotating {
					diskType = diskType + " (HDD)"
				}
				diskType = diskType + " (Whole Disk)"

				devices = append(devices, &BlockDevice{
					Name:                 deviceName,
					Path:                 devicePath,
					Major:                majorNum,
					Minor:                minorNum,
					DiskType:             diskType,
					Removable:            removable,
					Rotating:             rotating,
					LogicalBlockSize:     logicalBlockSizeBytes,
					PhysicalBlockSize:    physicalBlockSizeBytes,
					SectorCount:          sectors,
					CapacityMiB:          sizeMib,
					Model:                diskModel,
					Manufacturer:         diskManufacturer,
					Serial:               diskSerial,
					WWID:                 diskWWID,
					Firmware:             diskFirmware,
					NvmeQualifiedName:    nvmeQualifiedName,
					PCIeCurrentLinkSpeed: pcieCurrentLinkSpeed,
					PCIeCurrentLinkWidth: pcieCurrentLinkWidth,
					PCIeMaxLinkSpeed:     pcieMaxLinkSpeed,
					PCIeMaxLinkWidth:     pcieMaxLinkWidth,
				})
			}
		}
	}
	return devices, nil
}

func readFileAndTrim(filePath string) string {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(fileBytes))
}

func readUint(filePath string) uint64 {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return 0
	}
	value, err := strconv.ParseUint(strings.TrimSpace(string(fileBytes)), 10, 64)
	if err != nil {
		return 0
	}
	return value
}

func readUintBool(filePath string) bool {
	return readUint(filePath) == 1
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
