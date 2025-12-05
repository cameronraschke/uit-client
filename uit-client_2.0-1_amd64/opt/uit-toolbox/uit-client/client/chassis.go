//go:build linux && amd64

package client

import "strconv"

func GetChassisSerial() *string {
	return readFileAndTrim("/sys/class/dmi/id/chassis_serial")
}

func GetChassisType() *string {
	// https://github.com/mirror/dmidecode/blob/master/dmidecode.c#L599
	var data *int64
	var notInDmiTable = "Not in DMI table"
	data = readUintPtr("/sys/class/dmi/id/chassis_type")
	if data == nil {
		return &notInDmiTable
	}
	switch *data {
	case 1:
		other := "Other"
		return &other
	case 2:
		unknown := "Unknown"
		return &unknown
	case 3:
		desktop := "Desktop"
		return &desktop
	case 4:
		lowProfileDesktop := "Low Profile Desktop"
		return &lowProfileDesktop
	case 6:
		miniTower := "Mini Tower"
		return &miniTower
	case 7:
		tower := "Tower"
		return &tower
	case 8:
		portable := "Portable"
		return &portable
	case 9:
		laptop := "Laptop"
		return &laptop
	case 10:
		notebook := "Notebook"
		return &notebook
	case 13:
		allInOne := "All in One"
		return &allInOne
	case 15:
		spaceSaving := "Space-saving"
		return &spaceSaving
	case 30:
		tablet := "Tablet"
		return &tablet
	case 31:
		convertible := "Convertible"
		return &convertible
	case 32:
		detachable := "Detachable"
		return &detachable
	case 34:
		embeddedPC := "Embedded PC"
		return &embeddedPC
	case 35:
		miniPC := "Mini PC"
		return &miniPC
	case 36:
		stickPC := "Stick PC"
		return &stickPC
	default:
		if strconv.FormatInt(*data, 10) != "" {
			unknownType := "Unknown/" + strconv.FormatInt(*data, 10)
			return &unknownType
		}
		return &notInDmiTable
	}
}
