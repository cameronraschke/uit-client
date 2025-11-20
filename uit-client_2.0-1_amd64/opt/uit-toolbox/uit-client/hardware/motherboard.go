package hardware

func GetMotherboardSerial() string {
	return string(readFileAndTrim("/sys/class/dmi/id/board_serial"))
}

func GetMotherboardBiosDate() string {
	return string(readFileAndTrim("/sys/class/dmi/id/bios_date"))
}

func GetMotherboardBiosVersion() string {
	return string(readFileAndTrim("/sys/class/dmi/id/bios_version"))
}

func GetMotherboardManufacturer() string {
	return string(readFileAndTrim("/sys/class/dmi/id/board_vendor"))
}

func GetMotherboardProductName() string {
	return string(readFileAndTrim("/sys/class/dmi/id/board_name"))
}

func GetEmbeddedControllerVersion() string {
	if readFileAndTrim("/sys/ec_firmware_release/dmi/id/ec_firmware_release") != "" {
		return string(readFileAndTrim("/sys/ec_firmware_release/dmi/id/ec_firmware_release"))
	} else if readFileAndTrim("/sys/class/dmi/id/board_ec_version") != "" {
		return string(readFileAndTrim("/sys/class/dmi/id/board_ec_version"))
	}
	return ""
}

func GetDellSecureBootEnabled() bool {
	secureBootEnabled := readFileAndTrim("/sys/class/firmware-attributes/dell-wmi-sysman/attributes/SecureBoot/current_value")  // Can be "Enabled", "Disabled"
	secureBootMode := readFileAndTrim("/sys/class/firmware-attributes/dell-wmi-sysman/attributes/SecureBootMode/current_value") // Can be "AuditMode", "DeployedMode"
	tpmEnabled := readFileAndTrim("/sys/class/firmware-attributes/dell-wmi-sysman/attributes/TpmSecurity/current_value")        // Can be "Enabled", "Disabled"
	if secureBootEnabled == "Enabled" && secureBootMode == "DeployedMode" && tpmEnabled == "Enabled" {
		return true
	}

	return false
}
