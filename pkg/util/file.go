package util

import "os"

func IsDirectoryExists(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	if stat.IsDir() {
		return true
	}
	return false
}

func EnsureDirectoryExists(path string) error {
	if !IsDirectoryExists(path) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

func IsFileExists(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	if stat.IsDir() {
		return false
	}
	return true
}

func RemoveFile(filePath string) {
	_ = os.RemoveAll(filePath)
}

func WriteIntoFile(content string, fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}

func ReadFromFile(fileName string) ([]byte, error) {
	return os.ReadFile(fileName)
}
