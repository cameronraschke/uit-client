//go:build linux && amd64

package config

import "uitclient/types"

func SetTagnumber(tag *int64) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.Tagnumber, tag)
	})
}

func GetTagnumber() *int64 {
	cd := GetClientData()
	if cd.Tagnumber == nil {
		return nil
	}
	val := *cd.Tagnumber
	return &val
}

func SetSystemSerial(serial *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.Serial, serial)
	})
}

func SetSystemUUID(systemUUID *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.UUID, systemUUID)
	})
}

func SetManufacturer(manufacturer *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.Manufacturer, manufacturer)
	})
}

func SetModel(model *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.Model, model)
	})
}

func SetProductFamily(productFamily *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.ProductFamily, productFamily)
	})
}

func SetProductName(productName *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.ProductName, productName)
	})
}

func SetSKU(sku *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.SKU, sku)
	})
}
