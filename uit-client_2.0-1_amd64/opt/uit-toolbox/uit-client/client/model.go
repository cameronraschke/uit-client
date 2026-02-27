//go:build linux && amd64

package client

func GetProductFamily() *string {
	return ReadFileAndTrim("/sys/class/dmi/id/product_family")
}

func GetSystemManufacturer() *string {
	return ReadFileAndTrim("/sys/class/dmi/id/sys_vendor")
}

func GetSystemModel() *string {
	return ReadFileAndTrim("/sys/class/dmi/id/product_name")
}

func GetProductSKU() *string {
	return ReadFileAndTrim("/sys/class/dmi/id/product_sku")
}
