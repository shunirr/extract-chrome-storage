package core

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func GetChromeCookies(chromeCookiePath string, hostKey string, keychainAccount string) (map[string]string, error) {
	db, err := sql.Open("sqlite3", chromeCookiePath)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	versionStr, err := getDbVersion(db)
	if err != nil {
		return nil, err
	}
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return nil, err
	}

	password, err := GetKeychainPassword(keychainAccount)
	if err != nil {
		return nil, err
	}

	cookies, err := getCookies(db, hostKey, password, version)
	if err != nil {
		return nil, err
	}

	cookiesMap := make(map[string]string)
	for _, cookie := range cookies {
		cookiesMap[cookie[0]] = cookie[1]
	}

	return cookiesMap, nil
}

func getDbVersion(db *sql.DB) (string, error) {
	row := db.QueryRow("SELECT value FROM meta WHERE key = 'version' LIMIT 1")
	var version string
	err := row.Scan(&version)
	if err != nil {
		return "", err
	}

	return version, nil
}

func getCookies(db *sql.DB, hostKey string, password string, version int) ([][]string, error) {
	rows, err := db.Query(
		fmt.Sprintf(
			"SELECT name, encrypted_value FROM cookies WHERE host_key = '%s' AND encrypted_value IS NOT NULL",
			hostKey,
		),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cookies [][]string
	for rows.Next() {
		var name string
		var encryptedValue []byte
		err := rows.Scan(&name, &encryptedValue)
		if err != nil {
			continue
		}

		key := GetDecryptKey(password)
		decryptedValue, err := DecryptChromeCookieValue(encryptedValue, key, version)
		if err != nil {
			continue
		}

		cookies = append(cookies, []string{name, decryptedValue})
	}

	return cookies, nil
}
