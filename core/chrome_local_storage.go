package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/syndtr/goleveldb/leveldb"
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

	return string(value), nil
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
