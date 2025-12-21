//go:build linux && amd64

package config

import (
	"fmt"
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
	cd := clientData.Load()
	if cd == nil {
		return &ClientData{}
	}
	// Return an immutable snapshot (deep copy if callers might mutate)
	return copyClientData(cd)
}

// UpdateClientData performs an unconditional copy-on-write update.
func UpdateClientData(mutate func(*ClientData)) {
	currentSnapshot := clientData.Load()
	if currentSnapshot == nil {
		currentSnapshot = &ClientData{}
	}
	newSnapshot := copyClientData(currentSnapshot)
	mutate(newSnapshot)
	clientData.Store(newSnapshot)
}

// UpdateUniqueClientData performs a copy-on-write update only if mutate reports change.
func UpdateUniqueClientData(mutate func(*ClientData) bool) {
	for {
		oldSnapshot := clientData.Load()
		newSnapshot := copyClientData(oldSnapshot) // returns empty struct if currentSnapshot is nil
		if !mutate(newSnapshot) {
			return
		}
		if clientData.CompareAndSwap(oldSnapshot, newSnapshot) {
			return
		}
	}
}

func SetTagnumber(tag *int64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		return updateOptional(&cd.Tagnumber, tag)
	})
}

func SetSystemSerial(serial *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		return updateOptional(&cd.Serial, serial)
	})
}

func SetSystemUUID(systemUUID *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		return updateOptional(&cd.UUID, systemUUID)
	})
}

func SetManufacturer(manufacturer *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		return updateOptional(&cd.Manufacturer, manufacturer)
	})
}

func SetModel(model *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		return updateOptional(&cd.Model, model)
	})
}

func SetProductFamily(productFamily *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		return updateOptional(&cd.ProductFamily, productFamily)
	})
}

func SetProductName(productName *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		return updateOptional(&cd.ProductName, productName)
	})
}

func SetSKU(sku *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		return updateOptional(&cd.SKU, sku)
	})
}

func SetJobUUID(uid uuid.UUID) {
	if uid == uuid.Nil {
		return
	}
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.UUID == uid {
			return false
		}
		cd.JobData.UUID = uid
		return true
	})
}

func SetBootDuration(duration time.Duration) {
	if duration.Seconds() <= 0 {
		return
	}
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
		return updateOptional(&cd.ConnectedToHost, connected)
	})
}

func SetTimeSynced(synced *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		return updateOptional(&cd.NTPSynced, synced)
	})
}
