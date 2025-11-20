package hardware

import "strconv"

func GetChassisSerial() string {
	return string(readFileAndTrim("/sys/class/dmi/id/chassis_serial"))
}

func GetChassisType() string {
	// https://github.com/mirror/dmidecode/blob/master/dmidecode.c#L599
	data := readUint("/sys/class/dmi/id/chassis_type")
	switch data {
	case 1:
		return "Other"
	case 2:
		return "Unknown"
	case 3:
		return "Desktop"
	case 4:
		return "Low Profile Desktop"
	case 6:
		return "Mini Tower"
	case 7:
		return "Tower"
	case 8:
		return "Portable"
	case 9:
		return "Laptop"
	case 10:
		return "Notebook"
	case 13:
		return "All in One"
	case 15:
		return "Space-saving"
	case 30:
		return "Tablet"
	case 31:
		return "Convertible"
	case 32:
		return "Detachable"
	case 34:
		return "Embedded PC"
	case 35:
		return "Mini PC"
	case 36:
		return "Stick PC"
	default:
		if strconv.FormatUint(data, 10) != "" {
			return "Unknown/" + strconv.FormatUint(data, 10)
		}
		return "Not in DMI table"
	}
}
