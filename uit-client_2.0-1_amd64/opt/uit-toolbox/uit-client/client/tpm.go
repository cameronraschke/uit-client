//go:build linux && amd64

package client

func GetTPMVersion() *string {
	if fileExists("/dev/tpmrm0") {
		version := "2.0"
		return &version
	}
	data := ReadUintPtr("/sys/class/tpm/tpm0/tpm_version_major")
	if data == nil {
		notPresent := "Not Present"
		return &notPresent
	}
	switch *data {
	case 2:
		version := "2.0"
		return &version
	case 1:
		version := "1.2"
		return &version
	default:
		notPresent := "Not Present"
		return &notPresent
	}
}

func GetTPMDescription() *string {
	description1 := ReadFileAndTrim("/sys/class/tpm/tpm0/device/firmware_node/description")
	description2 := ReadFileAndTrim("/sys/class/tpm/tpm0/device/description")
	if description1 != nil && *description1 != "" {
		return description1
	} else if description2 != nil && *description2 != "" {
		return description2
	}
	return nil
}
