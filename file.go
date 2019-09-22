package main

import (
	"os"
	"strconv"
	"strings"
)

func isNotExist(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return true
	}
	return false
}

func creatFile(filePath string) (file *os.File, err error) {
	count := 2
	for {
		if !isNotExist(filePath) {
			// 重命名
			index := strings.LastIndex(filePath, ".")
			if index == -1 {
				filePath = filePath + strconv.Itoa(count)
				count++
				continue
			} else {
				temp := []byte(filePath)
				filePath = string(temp[0:index]) + strconv.Itoa(count) + string(temp[index:])
				count++
				continue
			}
		}
		break
	}
	file, err = os.Create(filePath)
	return
}
