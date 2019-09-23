package main

import (
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/valyala/fasthttp"
	"strings"
	"yan.com/downloader/models"
)

var fileList []models.FileInfo = make([]models.FileInfo, 0)

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
	fileInfo := models.FileInfo{
		Renewal:  renewal,
		Length:   resp.Header.ContentLength(),
		Url:      url,
		FileName: file.Name(),
		FilePath: "./" + file.Name(),
		File:     file,
		Exit:     make(chan bool),
		Down:     true,
	}
	fileList = append(fileList, fileInfo)
	return fileInfo, nil
}

func getHead(url string) *fasthttp.Request {
	head := fasthttp.RequestHeader{}
	head.SetRequestURI(url)
	head.SetUserAgent(browser.Chrome())
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
	head.SetUserAgent(browser.Chrome())
	head.SetByteRange(start, end)
	head.SetMethod(fasthttp.MethodGet)
	request := &fasthttp.Request{
		Header: head,
	}
	return request
}

func send(req *fasthttp.Request, resp *fasthttp.Response, segment *models.SegMent, fileInfo models.FileInfo, ok *chan bool) {
LOOP:
	for {
		err := client.Do(req, resp)
		if err != nil {
			segment.Count++
			if segment.Count >= 3 {
				fileInfo.Exit <- true
				break
			}
			break LOOP
		}
		err = writeFile(fileInfo, segment, resp.Body())
		if err != nil {
			segment.Count++
			if segment.Count >= 3 {
				fileInfo.Exit <- true
				break
			}
			break LOOP
		}
		size := getFileSize(fileInfo.FilePath)
		*ok <- true
		if size == int64(fileInfo.Length) {
			fmt.Println("\n下载完成")
			fileInfo.File.Close()
			fileInfo.Exit <- true
			group.Done()
			break
		}
	}
}
