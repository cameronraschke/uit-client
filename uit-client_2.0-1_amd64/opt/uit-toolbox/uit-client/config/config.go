//go:build linux && amd64

package config

import (
	"fmt"
	"sync"
	"sync/atomic"
	"uitclient/types"
)

var (
	clientConfig atomic.Pointer[ClientConfig]
	clientDataMu sync.RWMutex
	clientData   *types.ClientData
)

type ClientLookup struct {
	Tagnumber    *int64  `json:"tagnumber"`
	SystemSerial *string `json:"system_serial"`
}

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

func InitializeClientData(data *types.ClientData) error {
	clientDataMu.Lock()
	defer clientDataMu.Unlock()
	if data == nil {
		return fmt.Errorf("cannot initialize app, client data is nil")
	}
	clientData = data
	return nil
}

func GetClientData() types.ClientData {
	clientDataMu.RLock()
	defer clientDataMu.RUnlock()

	if clientData == nil {
		return types.ClientData{}
	}
	return *clientData // shallow snapshot
}

// UpdateClientData performs an unconditional copy-on-write update.
func UpdateClientData(mutate func(*types.ClientData)) {
	clientDataMu.Lock()
	defer clientDataMu.Unlock()

	if clientData == nil {
		clientData = &types.ClientData{}
	}
	mutate(clientData)
}

// UpdateUniqueClientData performs a copy-on-write update only if mutate reports change.
func UpdateUniqueClientData(mutate func(*types.ClientData) bool) {
	clientDataMu.Lock()
	defer clientDataMu.Unlock()

	if clientData == nil {
		clientData = &types.ClientData{}
	}
	changed := mutate(clientData)
	if !changed {
		return
	}
}

func SetConnectedToHost(connected *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.ConnectedToHost, connected)
	})
}

func SetTimeSynced(synced *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.NTPSynced, synced)
	})
}
