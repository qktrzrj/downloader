package main

import (
	"fmt"
	"time"
)

func main() {
	url := "https://www.typora.io/windows/typora-setup-x64.exe"
	err := download(url)
	if err != nil {
		panic(err)
	}
	fmt.Println("\n" + time.Now().String() + "退出下载")
	select {}
}
