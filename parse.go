package main

import (
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/valyala/fasthttp"
	"strings"
	"sync"
	"yan.com/downloader/models"
)

func getFileInfo(url string) (models.FileInfo, error) {
	// 请求目标文件信息
	resp := &fasthttp.Response{}
	err := client.Do(getHead(url), resp)
	if err != nil {
		return models.FileInfo{}, fmt.Errorf("请求目标文件信息失败: %w", err)
	}
	var renewal = false
	if temp := resp.Header.Peek("Accept-Ranges"); len(temp) != 0 {
		renewal = true
	}
	u := []byte(url)
	fullName := ""
	s := strings.LastIndex(url, "/")
	if s == -1 {
		s = 0
		fullName = string(u[s:])
	} else {
		fullName = string(u[s+1:])
	}
	// 创建目标文件
	file, err := creatFile("./" + fullName)
	if err != nil {
		return models.FileInfo{}, err
	}
	//file.Close()
	fileInfo := models.FileInfo{
		Renewal:   renewal,
		Length:    resp.Header.ContentLength(),
		Url:       url,
		FileName:  file.Name(),
		FilePath:  "./" + file.Name(),
		File:      file,
		Exit:      make(chan bool),
		TaskExit:  make(chan bool),
		RateExit:  make(chan bool),
		Down:      true,
		Lock:      sync.Mutex{},
		Task:      sync.RWMutex{},
		FileCache: make(map[int][]byte),
	}
	fileMap[fileInfo.FilePath] = &fileInfo
	return fileInfo, nil
}

func getHead(url string) *fasthttp.Request {
	head := fasthttp.RequestHeader{}
	head.SetRequestURI(url)
	head.SetUserAgent(browser.Random())
	//head.Set(fasthttp.HeaderIfRange, "true")
	head.SetMethod(fasthttp.MethodHead)
	request := &fasthttp.Request{
		Header: head,
	}
	return request
}

func getRequest(url string, start int, end int) *fasthttp.Request {
	head := fasthttp.RequestHeader{}
	head.SetRequestURI(url)
	head.SetUserAgent(browser.Random())
	head.SetByteRange(start, end)
	head.SetMethod(fasthttp.MethodGet)
	request := &fasthttp.Request{
		Header: head,
	}
	return request
}

func send(req *fasthttp.Request, segment *models.SegMent, fileInfo *models.FileInfo) {
	resp := &fasthttp.Response{}
LOOP:
	for {
		err := client.Do(req, resp)
		if err != nil {
			segment.Count++
			if segment.Count >= 3 {
				fileInfo.Exit <- true
				return
			}
			break LOOP
		}
		segment.Cache = resp.Body()
		resp.ConnectionClose()
		return
	}
}
