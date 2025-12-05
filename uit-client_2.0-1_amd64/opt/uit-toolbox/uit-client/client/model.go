//go:build linux && amd64

package client

func GetProductFamily() *string {
	return readFileAndTrim("/sys/class/dmi/id/product_family")
}

func GetSystemManufacturer() *string {
	return readFileAndTrim("/sys/class/dmi/id/sys_vendor")
}

func GetSystemModel() *string {
	return readFileAndTrim("/sys/class/dmi/id/product_name")
}

func GetProductSKU() *string {
	return readFileAndTrim("/sys/class/dmi/id/product_sku")
}
