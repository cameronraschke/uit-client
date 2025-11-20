package hardware

func GetProductFamily() string {
	return string(readFileAndTrim("/sys/class/dmi/id/product_family"))
}

func GetProductName() string {
	return string(readFileAndTrim("/sys/class/dmi/id/product_name"))
}

func GetProductSKU() string {
	return string(readFileAndTrim("/sys/class/dmi/id/product_sku"))
}
