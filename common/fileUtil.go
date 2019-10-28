package common

import "os"

// if not exist return false else return true
func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func GetFileSize(path string) int64 {
	info, _ := os.Stat(path)
	return info.Size()
}
