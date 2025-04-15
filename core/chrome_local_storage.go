package core

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/text/encoding/unicode"

	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/text/transform"
)

func ListChromeLocalStorageKeys(localStoragePath string, url string) ([]string, error) {
	if localStoragePath == "" {
		return nil, fmt.Errorf("local storage path is empty")
	}
	if url == "" {
		return nil, fmt.Errorf("url is empty")
	}

	tempDir, err := os.MkdirTemp("", "chrome_local_storage")
	if err != nil {
		return nil, fmt.Errorf("failed to create a temporary directory\n%s", err)
	}
	defer os.RemoveAll(tempDir)

	// Workaround: The script cannot open the database if the Chrome is running.
	err = copyLevelDbFile(localStoragePath, tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to copy the LevelDB file: src: '%s' dst: '%s'\n%s", localStoragePath, tempDir, err)
	}

	err = removeLockFile(tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to remove the lock file\n%s", err)
	}

	db, err := leveldb.OpenFile(tempDir, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open the LevelDB file\n%s", err)
	}
	defer db.Close()

	keys := make([]string, 0)
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		keys = append(keys, string(key))
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		return nil, fmt.Errorf("failed to iterate the LevelDB\n%s", err)
	}

	return keys, nil
}

func GetChromeLocalStorage(localStoragePath string, url string, localStorageKey string) (string, error) {
	if localStoragePath == "" {
		return "", fmt.Errorf("local storage path is empty")
	}
	if url == "" {
		return "", fmt.Errorf("url is empty")
	}
	if localStorageKey == "" {
		return "", fmt.Errorf("local storage key is empty")
	}

	tempDir, err := os.MkdirTemp("", "chrome_local_storage")
	if err != nil {
		return "", fmt.Errorf("failed to create a temporary directory\n%s", err)
	}
	defer os.RemoveAll(tempDir)

	// Workaround: The script cannot open the database if the Chrome is running.
	err = copyLevelDbFile(localStoragePath, tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to copy the LevelDB file: src: '%s' dst: '%s'\n%s", localStoragePath, tempDir, err)
	}

	err = removeLockFile(tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to remove the lock file\n%s", err)
	}

	db, err := leveldb.OpenFile(tempDir, nil)
	if err != nil {
		return "", fmt.Errorf("failed to open the LevelDB file\n%s", err)
	}
	defer db.Close()

	key := []byte(fmt.Sprintf("_%s\x00\x01%s", url, localStorageKey))
	exist, err := db.Has(key, nil)
	if err != nil {
		return "", fmt.Errorf("failed to check the key: %s\n%s", key, err)
	}
	if !exist {
		return "", fmt.Errorf("key not found: %s", key)
	}

	value, err := db.Get(key, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get the key: %s\n%s", key, err)
	}

	if value[0] == 0x01 {
		value = value[1:]
	}

	json, err := recoverBrokenJson(value)
	if err != nil {
		return "", fmt.Errorf("failed to recover broken json: %s\n%s", value, err)
	}

	return json, nil
}

func decodeUTF16(b []byte) (string, error) {
	decoder := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()
	reader := transform.NewReader(bytes.NewReader(b), decoder)
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func isLikelyUTF16(b []byte) bool {
	if len(b) < 4 || len(b)%2 != 0 {
		return false
	}

	zeroCount := 0
	for i := 0; i < len(b); i += 2 {
		if b[i] == 0x00 {
			zeroCount++
		}
	}

	return zeroCount >= len(b)/4
}

func recoverBrokenJson(input []byte) (string, error) {
	var decoded string

	utf16input := input[:len(input)-1]
	if isLikelyUTF16(utf16input) {
		var err error
		decoded, err = decodeUTF16(utf16input)
		if err != nil {
			return "", fmt.Errorf("failed to decode UTF-16: %s", err)
		}
	} else {
		decoded = string(input)
	}

	re := regexp.MustCompile(`:"([^",]+),"`)
	fixed := re.ReplaceAllString(decoded, `:"$1","`)
	return fixed, nil
}

func copyLevelDbFile(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, data, info.Mode())
	})
}

func removeLockFile(dir string) error {
	return os.Remove(filepath.Join(dir, "LOCK"))
}
