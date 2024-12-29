package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

func cacheDir() (string, error) {
	dir, err := os.UserCacheDir()

	if err != nil {
		return "", err
	}

	wd, err := os.Getwd()

	if err != nil {
		return "", err
	}

	dir = filepath.Join(dir, filepath.Base(os.Args[0]), filepath.Base(wd))

	if err := os.MkdirAll(dir, dirMode); err != nil {
		return "", err
	}
	return dir, nil
}

func cacheKey(path string) string {
	sha256 := sha256.New()
	io.WriteString(sha256, path)
	return hex.EncodeToString(sha256.Sum(nil))
}

func CacheHash(path string) ([]byte, error) {
	b, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	sha256 := sha256.New()
	sha256.Write(b)

	return sha256.Sum(nil), nil
}

func CacheGet(path string) ([]byte, error) {
	dir, err := cacheDir()

	if err != nil {
		return nil, err
	}

	key := cacheKey(path)

	b, err := os.ReadFile(filepath.Join(dir, key))

	if err != nil {
		return nil, err
	}
	return b, nil
}

func CachePut(path string) error {
	dir, err := cacheDir()

	if err != nil {
		return err
	}

	key := cacheKey(path)

	f, err := os.Create(filepath.Join(dir, key))

	if err != nil {
		return err
	}

	defer f.Close()

	hash, err := CacheHash(path)

	if err != nil {
		return err
	}

	if _, err := f.Write(hash); err != nil {
		return err
	}
	return nil
}
