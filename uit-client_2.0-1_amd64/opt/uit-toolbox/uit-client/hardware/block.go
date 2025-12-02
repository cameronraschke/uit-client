//go:build linux
// +build linux

package hardware

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"uitclient/config"

	"golang.org/x/sys/unix"
)

func ListBlockDevices(devDir string) ([]*config.DiskHardwareData, error) {
	fd, err := unix.Open(devDir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open device directory: %v", err)
	}
	defer unix.Close(fd)

	directoryListBuffer := make([]byte, 1<<13) // 8kb buffer to load in directory entries
	var devices []config.DiskHardwareData

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
			if deviceName == "." || deviceName == ".." || deviceName == "" {
				continue
			}
			devicePath := filepath.Join(devDir, deviceName)
			if !fileExists(devicePath) {
				continue
			}

			var statData unix.Stat_t
			if err := unix.Lstat(devicePath, &statData); err != nil {
				continue
			}
			// Only select block devices - bitwise and to mask file type bits
			if statData.Mode&unix.S_IFMT != unix.S_IFBLK {
				continue
			}

			major := unix.Major(uint64(statData.Rdev))
			minor := unix.Minor(uint64(statData.Rdev))

			if major == 0 && minor == 0 {
				continue
			}
			majorNum := int64(major)
			minorNum := int64(minor)

			// Determine disk type based on major number and device path prefix
			diskType = "Unknown"

			if slices.Contains(scsiOrSataMajors, major) && strings.HasPrefix(devicePath, "/dev/sd") {
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
			} else if slices.Contains(scsiIdeMajors, major) && strings.HasPrefix(devicePath, "/dev/hd") {
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
			removable := readUintBoolPtr(filepath.Join(sysBlock, "removable"))

			rotating := readUintBoolPtr(filepath.Join(sysBlock, "queue", "rotational"))
			lBlocks := readUintPtr(filepath.Join(sysBlock, "queue", "logical_block_size"))
			var logicalBlockSizeBytes uint32
			if lBlocks != nil {
				logicalBlockSizeBytes = uint32(*lBlocks)
			} else {
				logicalBlockSizeBytes = 512 // Default to 512 if not found
			}
			pBlocks := readUintPtr(filepath.Join(sysBlock, "queue", "physical_block_size"))
			var physicalBlockSizeBytes uint32
			if pBlocks != nil {
				physicalBlockSizeBytes = uint32(*pBlocks)
			} else {
				physicalBlockSizeBytes = 512 // Default to 512 if not found
			}

			sectors := readUintPtr(filepath.Join(sysBlock, "size"))
			var sizeBytes uint64
			if sectors != nil {
				sizeBytes = *sectors * uint64(logicalBlockSizeBytes)
			} else {
				sizeBytes = 0
			}
			sizeMib := float64(sizeBytes) / 1024.0 / 1024.0 // Convert bytes to MiB

			diskWWID := readFileAndTrim(filepath.Join(sysBlock, "wwid"))

			deviceSymLink := filepath.Join(sysBlock, "device")
			deviceRealPath, _ := filepath.EvalSymlinks(deviceSymLink)
			if deviceRealPath == "" || deviceRealPath == "/" {
				continue
			}
			diskModel := readFileAndTrim(filepath.Join(deviceSymLink, "model"))
			if diskModel != nil {
				trimmedModel := strings.TrimSpace(*diskModel)
				diskModel = &trimmedModel
			}
			diskManufacturer := readFileAndTrim(filepath.Join(deviceSymLink, "vendor"))
			diskSerial := ""
			serial := readFileAndTrim(filepath.Join(deviceSymLink, "serial"))
			if serial != nil {
				diskSerial = *serial
			} else if eui := readFileAndTrim(filepath.Join(devicePath, "eui")); eui != nil {
				diskSerial = *eui
			} else if euiFallback := readFileAndTrim(filepath.Join(deviceSymLink, "eui")); euiFallback != nil {
				diskSerial = *euiFallback
			}

			diskFirmware := ""
			if firmware := readFileAndTrim(filepath.Join(deviceSymLink, "firmware_rev")); firmware != nil {
				diskFirmware = *firmware
			} else if firmwareFallback := readFileAndTrim(filepath.Join(deviceSymLink, "rev")); firmwareFallback != nil {
				diskFirmware = *firmwareFallback
			}

			nvmeQualifiedName := readFileAndTrim(filepath.Join(deviceSymLink, "subsysnqn"))

			// No need to check the symlink here
			nvmeControllerSubsystemPath := filepath.Join(deviceSymLink, "subsystem", "nvme0", "device")

			pcieCurrentLinkSpeed := readFileAndTrim(filepath.Join(nvmeControllerSubsystemPath, "current_link_speed"))
			pcieCurrentLinkWidth := readFileAndTrim(filepath.Join(nvmeControllerSubsystemPath, "current_link_width"))
			pcieMaxLinkSpeed := readFileAndTrim(filepath.Join(nvmeControllerSubsystemPath, "max_link_speed"))
			pcieMaxLinkWidth := readFileAndTrim(filepath.Join(nvmeControllerSubsystemPath, "max_link_width"))

			if minorNum == 0 {
				if diskType == "SCSI/SATA" && rotating != nil && *rotating {
					diskType = diskType + " (HDD)"
				} else if diskType == "SCSI/SATA" && removable != nil && *removable {
					diskType = diskType + " (Removable)"
				} else if diskType == "SCSI/SATA" && rotating != nil && !*rotating {
					diskType = diskType + " (SSD)"
				} else if diskType == "NVMe" && rotating != nil && !*rotating {
					diskType = diskType + " (NVMe SSD)"
				}

				devices = append(devices, config.DiskHardwareData{
					LinuxAlias:           &deviceName,
					LinuxDevicePath:      &devicePath,
					LinuxMajorNumber:     &majorNum,
					LinuxMinorNumber:     &minorNum, // always 0 here
					InterfaceType:        &diskType, // always populated
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
