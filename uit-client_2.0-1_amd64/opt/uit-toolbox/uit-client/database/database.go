package database

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
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

func GetDatabaseConnectionString() (host string, port string, user string, password string, dbname string, err error) {
	configFile, err := os.ReadFile("/etc/uit-client.conf")
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to read configuration file: %v", err)
	}
	lines := strings.Split(string(configFile), "\n")
	if len(lines) < 5 {
		return "", "", "", "", "", errors.New("configuration file is incomplete")
	}

	for line := range lines {
		key, value, ok := strings.Cut(lines[line], "=")
		if !ok {
			return "", "", "", "", "", fmt.Errorf("invalid configuration line: %s", lines[line])
		}
		switch key {
		case "UIT_CLIENT_DB_HOST":
			host = value
		case "UIT_CLIENT_DB_PORT":
			_, err = strconv.Atoi(value)
			if err != nil {
				return "", "", "", "", "", fmt.Errorf("invalid port value: %v", err)
			}
			port = value
		case "UIT_CLIENT_DB_USER":
			user = value
		case "UIT_CLIENT_DB_PASSWD":
			password = value
		case "UIT_CLIENT_DB_NAME":
			dbname = value
		}
	}
	return host, port, user, password, dbname, nil
}

func CreateDBConnection() (*sql.DB, error) {
	dbHost, dbPort, dbUsername, dbPassword, dbName, err := GetDatabaseConnectionString()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection parameters: %v", err)
	}

	dbConnURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(dbUsername, dbPassword),
		Host:   net.JoinHostPort(dbHost, dbPort),
		Path:   dbName,
	}
	dbConnQuery := dbConnURL.Query()
	dbConnQuery.Set("sslmode", "disable")
	dbConnURL.RawQuery = dbConnQuery.Encode()

	db, err := sql.Open("pgx", dbConnURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	return db, nil
}
