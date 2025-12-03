package hardware

func GetMotherboardSerial() *string {
	return readFileAndTrim("/sys/class/dmi/id/board_serial")
}

func GetMotherboardBiosDate() *string {
	return readFileAndTrim("/sys/class/dmi/id/bios_date")
}

func GetMotherboardBiosVersion() *string {
	return readFileAndTrim("/sys/class/dmi/id/bios_version")
}

func GetMotherboardManufacturer() *string {
	return readFileAndTrim("/sys/class/dmi/id/board_vendor")
}

func GetMotherboardProductName() *string {
	return readFileAndTrim("/sys/class/dmi/id/board_name")
}

func GetEmbeddedControllerVersion() *string {
	if readFileAndTrim("/sys/ec_firmware_release/dmi/id/ec_firmware_release") != nil {
		return readFileAndTrim("/sys/ec_firmware_release/dmi/id/ec_firmware_release")
	} else if readFileAndTrim("/sys/class/dmi/id/board_ec_version") != nil {
		return readFileAndTrim("/sys/class/dmi/id/board_ec_version")
	}
	return nil
}

func GetDellSecureBootEnabled() *bool {
	secureBootEnabled := readFileAndTrim("/sys/class/firmware-attributes/dell-wmi-sysman/attributes/SecureBoot/current_value")  // Can be "Enabled", "Disabled"
	secureBootMode := readFileAndTrim("/sys/class/firmware-attributes/dell-wmi-sysman/attributes/SecureBootMode/current_value") // Can be "AuditMode", "DeployedMode"
	tpmEnabled := readFileAndTrim("/sys/class/firmware-attributes/dell-wmi-sysman/attributes/TpmSecurity/current_value")        // Can be "Enabled", "Disabled"
	if secureBootEnabled != nil && secureBootMode != nil && tpmEnabled != nil && *secureBootEnabled == "Enabled" && *secureBootMode == "DeployedMode" && *tpmEnabled == "Enabled" {
		trueVal := true
		return &trueVal
	}

	falseVal := false
	return &falseVal
}
