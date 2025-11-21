package config

import (
	"fmt"
	"maps"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var (
	clientConfig atomic.Pointer[ClientConfig]
	clientData   atomic.Pointer[ClientData]
)

type ClientConfig struct {
	UIT_CLIENT_DB_USER   string `json:"UIT_CLIENT_DB_USER"`
	UIT_CLIENT_DB_PASSWD string `json:"UIT_CLIENT_DB_PASSWD"`
	UIT_CLIENT_DB_NAME   string `json:"UIT_CLIENT_DB_NAME"`
	UIT_CLIENT_DB_HOST   string `json:"UIT_CLIENT_DB_HOST"`
	UIT_CLIENT_DB_PORT   string `json:"UIT_CLIENT_DB_PORT"`
	UIT_CLIENT_NTP_HOST  string `json:"UIT_CLIENT_NTP_HOST"`
	UIT_CLIENT_PING_HOST string `json:"UIT_CLIENT_PING_HOST"`
	UIT_SERVER_HOSTNAME  string `json:"UIT_SERVER_HOSTNAME"`
	UIT_WEB_HTTP_HOST    string `json:"UIT_WEB_HTTP_HOST"`
	UIT_WEB_HTTP_PORT    string `json:"UIT_WEB_HTTP_PORT"`
	UIT_WEB_HTTPS_HOST   string `json:"UIT_WEB_HTTPS_HOST"`
	UIT_WEB_HTTPS_PORT   string `json:"UIT_WEB_HTTPS_PORT"`
	UIT_WEBMASTER_NAME   string `json:"UIT_WEBMASTER_NAME"`
}

func InitializeClientConfig(config *ClientConfig) error {
	if config == nil {
		return fmt.Errorf("cannot initialize app, client config is nil")
	}
	clientConfig.Store(config)
	return nil
}

func GetClientConfig() *ClientConfig {
	return clientConfig.Load()
}

func InitializeClientData(data *ClientData) error {
	if data == nil {
		return fmt.Errorf("cannot initialize app, client data is nil")
	}
	clientData.Store(data)
	return nil
}

func GetClientData() *ClientData {
	return clientData.Load()
}

// UpdateClientData performs an unconditional copy-on-write update.
func UpdateClientData(mutate func(*ClientData)) {
	currentSnapshot := clientData.Load()
	if currentSnapshot == nil {
		currentSnapshot = &ClientData{}
	}
	newSnapshot := cloneClientData(currentSnapshot)
	mutate(newSnapshot)
	clientData.Store(newSnapshot)
}

// UpdateUniqueClientData performs a copy-on-write update only if mutate reports change.
func UpdateUniqueClientData(mutate func(*ClientData) bool) {
	currentSnapshot := clientData.Load()
	if currentSnapshot == nil {
		currentSnapshot = &ClientData{}
	}
	newSnapshot := cloneClientData(currentSnapshot)
	if !mutate(newSnapshot) {
		return
	}
	clientData.Store(newSnapshot)
}

func SetTagnumber(tag int64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Tagnumber == tag {
			return false
		}
		cd.Tagnumber = tag
		return true
	})
}

func SetSystemSerial(serial string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Serial == serial {
			return false
		}
		cd.Serial = serial
		return true
	})
}

func SetManufacturer(manufacturer string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Manufacturer == manufacturer {
			return false
		}
		cd.Manufacturer = manufacturer
		return true
	})
}

func SetModel(model string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Model == model {
			return false
		}
		cd.Model = model
		return true
	})
}

func SetProductFamily(productFamily string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.ProductFamily == productFamily {
			return false
		}
		cd.ProductFamily = productFamily
		return true
	})
}

func SetProductName(productName string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.ProductName == productName {
			return false
		}
		cd.ProductName = productName
		return true
	})
}

func SetSKU(sku string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.SKU == sku {
			return false
		}
		cd.SKU = sku
		return true
	})
}

func SetUUID(uid uuid.UUID) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.UUID == uid {
			return false
		}
		cd.UUID = uid
		return true
	})
}

func SetOEMStrings(oemStrings map[string]string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if maps.Equal(cd.OEMStrings, oemStrings) {
			return false
		}
		newMap := make(map[string]string, len(oemStrings))
		maps.Copy(newMap, oemStrings)
		cd.OEMStrings = newMap
		return true
	})
}

func SetBootDuration(duration time.Duration) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.BootDuration == duration {
			return false
		}
		cd.BootDuration = duration
		return true
	})
}

func SetConnectedToHost(connected *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if connected == nil {
			if cd.ConnectedToHost == nil {
				return false
			}
			cd.ConnectedToHost = nil
			return true
		}
		pointerVal := *connected
		if cd.ConnectedToHost != nil && *cd.ConnectedToHost == pointerVal {
			return false
		}
		// Store a new pointer to a copied value
		newCopy := pointerVal
		cd.ConnectedToHost = &newCopy
		return true
	})
}

func SetTimeSynced(synced *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if synced == nil {
			if cd.TimeSynced == nil {
				return false
			}
			cd.TimeSynced = nil
			return true
		}
		pointerVal := *synced
		if cd.TimeSynced != nil && *cd.TimeSynced == pointerVal {
			return false
		}
		newCopy := pointerVal
		cd.TimeSynced = &newCopy
		return true
	})
}
