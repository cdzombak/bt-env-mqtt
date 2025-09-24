package main

import (
	"os"
)

// Test helper functions

func createTempFile(name, content string) (string, error) {
	tmpfile, err := os.CreateTemp("", name)
	if err != nil {
		return "", err
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		_ = tmpfile.Close()
		_ = os.Remove(tmpfile.Name())
		return "", err
	}

	if err := tmpfile.Close(); err != nil {
		_ = os.Remove(tmpfile.Name())
		return "", err
	}

	return tmpfile.Name(), nil
}

func removeTempFile(filename string) error {
	return os.Remove(filename)
}