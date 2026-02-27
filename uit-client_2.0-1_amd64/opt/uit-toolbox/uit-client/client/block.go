//go:build linux && amd64

package client

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"uitclient/types"

	"golang.org/x/sys/unix"
)

func ListBlockDevices(devDir string) ([]*types.DiskHardwareData, error) {
	if !fileExists(devDir) {
		devDir = "/dev"
	}
	devDir = filepath.Clean(devDir)
	if !filepath.IsAbs(devDir) || !strings.HasPrefix(devDir, "/dev") {
		return nil, fmt.Errorf("invalid device directory path: %s", devDir)
	}

	fd, err := unix.Open(devDir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open device directory: %v", err)
	}
	defer unix.Close(fd)

	directoryListBuffer := make([]byte, 1<<13) // 8kb buffer to load in directory entries
	var devices []*types.DiskHardwareData

	// Default to Unknown disk type
	diskType := "Unknown"
	for {
		directoryEntry, err := unix.ReadDirent(fd, directoryListBuffer)
		if err != nil {
			return nil, err
		}
		if directoryEntry <= 0 {
			break
		}

		// deviceNames will be files (or dirs) in /dev
		deviceNames := make([]string, 0, 128)
		_, _, deviceNames = unix.ParseDirent(directoryListBuffer[:directoryEntry], -1, deviceNames)

		// Common major numbers for SCSI & SATA
		scsiOrSataMajors := []uint32{8, 65, 66, 67, 68, 69, 70, 71, 128, 129, 130, 131, 132, 133, 134, 135}
		scsiIdeMajors := []uint32{3, 22, 33, 34, 56, 57, 88, 89, 90, 91}

		for _, deviceName := range deviceNames {
			deviceName = strings.TrimSpace(deviceName)
			if deviceName == "." || deviceName == ".." || deviceName == "" {
				continue
			}
			devicePath := filepath.Join(devDir, deviceName) // Join cleans, returns empty for invalid input
			if devicePath == "" {
				continue
			}

			if !fileExists(devicePath) {
				continue
			}

			// Lstat for sym links without following them
			var statData unix.Stat_t
			if err := unix.Lstat(devicePath, &statData); err != nil {
				continue
			}
			// Only select block devices - bitwise AND to mask file type bits
			// https://www.gnu.org/software/libc/manual/html_node/Testing-File-Type.html
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
			// diskType defaults to "Unknown" at start of loop
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
			removable := ReadUintBoolPtr(filepath.Join(sysBlock, "removable"))

			rotating := ReadUintBoolPtr(filepath.Join(sysBlock, "queue", "rotational"))
			lBlocks := ReadUintPtr(filepath.Join(sysBlock, "queue", "logical_block_size"))
			var logicalBlockSizeBytes *int64
			if lBlocks != nil {
				logicalBlockSizeBytes = lBlocks
			} else {
				defaultLogicalBlockSize := int64(512)
				logicalBlockSizeBytes = &defaultLogicalBlockSize
			}
			pBlocks := ReadUintPtr(filepath.Join(sysBlock, "queue", "physical_block_size"))
			var physicalBlockSizeBytes *int64
			if pBlocks != nil {
				physicalBlockSizeBytes = pBlocks
			} else {
				defaultPhysicalBlockSize := int64(512)
				physicalBlockSizeBytes = &defaultPhysicalBlockSize
			}

			sectors := ReadUintPtr(filepath.Join(sysBlock, "size"))
			var sizeBytes *int64
			if sectors != nil && logicalBlockSizeBytes != nil {
				tempSize := *sectors * *logicalBlockSizeBytes
				sizeBytes = &tempSize
			} else {
				sizeBytes = nil
			}

			var sizeMib *float64
			if sizeBytes != nil {
				tempSizeMib := float64(*sizeBytes) / 1024.0 / 1024.0 // Convert bytes to MiB
				sizeMib = &tempSizeMib
			}

			diskWWID := ReadFileAndTrim(filepath.Join(sysBlock, "wwid"))

			deviceSymLink := filepath.Join(sysBlock, "device")
			deviceRealPath, _ := filepath.EvalSymlinks(deviceSymLink)
			if deviceRealPath == "" || deviceRealPath == "/" || deviceRealPath == "." {
				continue
			}
			diskModel := ReadFileAndTrim(filepath.Join(deviceSymLink, "model"))
			if diskModel != nil {
				trimmedModel := strings.TrimSpace(*diskModel)
				diskModel = &trimmedModel
			}
			diskManufacturer := ReadFileAndTrim(filepath.Join(deviceSymLink, "vendor"))
			if diskManufacturer != nil {
				trimmedManufacturer := strings.TrimSpace(*diskManufacturer)
				diskManufacturer = &trimmedManufacturer
			}
			var diskSerial *string
			serial := ReadFileAndTrim(filepath.Join(deviceSymLink, "serial"))
			if serial != nil {
				diskSerial = serial
			} else if eui := ReadFileAndTrim(filepath.Join(devicePath, "eui")); eui != nil {
				diskSerial = eui
			} else if euiFallback := ReadFileAndTrim(filepath.Join(deviceSymLink, "eui")); euiFallback != nil {
				diskSerial = euiFallback
			}

			var diskFirmware *string
			if firmware := ReadFileAndTrim(filepath.Join(deviceSymLink, "firmware_rev")); firmware != nil {
				diskFirmware = firmware
			} else if firmwareFallback := ReadFileAndTrim(filepath.Join(deviceSymLink, "rev")); firmwareFallback != nil {
				diskFirmware = firmwareFallback
			}

			nvmeQualifiedName := ReadFileAndTrim(filepath.Join(deviceSymLink, "subsysnqn"))

			// No need to check the symlink here
			nvmeControllerSubsystemPath := filepath.Join(deviceSymLink, "subsystem", "nvme0", "device")

			pcieCurrentLinkSpeed := ReadFileAndTrim(filepath.Join(nvmeControllerSubsystemPath, "current_link_speed"))
			pcieCurrentLinkWidth := ReadFileAndTrim(filepath.Join(nvmeControllerSubsystemPath, "current_link_width"))
			pcieMaxLinkSpeed := ReadFileAndTrim(filepath.Join(nvmeControllerSubsystemPath, "max_link_speed"))
			pcieMaxLinkWidth := ReadFileAndTrim(filepath.Join(nvmeControllerSubsystemPath, "max_link_width"))

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

				devices = append(devices, &types.DiskHardwareData{
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
