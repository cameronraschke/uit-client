package hardware

func GetProductFamily() string {
	return string(readFileAndTrim("/sys/class/dmi/id/product_family"))
}

func GetSystemManufacturer() string {
	return string(readFileAndTrim("/sys/class/dmi/id/sys_vendor"))
}

func GetSystemModel() string {
	return string(readFileAndTrim("/sys/class/dmi/id/product_name"))
}

func GetProductSKU() string {
	return string(readFileAndTrim("/sys/class/dmi/id/product_sku"))
}
