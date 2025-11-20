package hardware

func GetTPMVersion() string {
	if fileExists("/dev/tpmrm0") {
		return "2.0"
	}
	data := readUint("/sys/class/tpm/tpm0/tpm_version_major")
	switch data {
	case 2:
		return "2.0"
	case 1:
		return "1.2"
	default:
		return "Not Present"
	}
}

func GetTPMDescription() string {
	description1 := readFileAndTrim("/sys/class/tpm/tpm0/device/firmware_node/description")
	description2 := readFileAndTrim("/sys/class/tpm/tpm0/device/description")
	if description1 != "" {
		return description1
	} else if description2 != "" {
		return description2
	}
	return ""
}
