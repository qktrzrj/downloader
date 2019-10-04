package main

import (
	"downloader/models"
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/valyala/fasthttp"
	"math/rand"
	"strings"
)

func getFileInfo(url string) (int, error) {
	// 请求目标文件信息
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req := getHead(url)
	defer fasthttp.ReleaseRequest(req)
	// 构建客户端
	fileId := rand.Int()
	clientMap[fileId] = &fasthttp.Client{}
	err := clientMap[fileId].Do(req, resp)
	if err != nil {
		return 0, fmt.Errorf("请求目标文件信息失败: %w", err)
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
		return 0, err
	}
	fileInfo := models.FileInfo{
		Id:       fileId,
		Row:      rowCount,
		Renewal:  renewal,
		Length:   resp.Header.ContentLength(),
		Url:      url,
		File:     file,
		FileName: fullName,
		FilePath: "./" + fullName,
		Exit:     make(chan bool),
	}
	fileMap[fileInfo.Id] = fileInfo
	taskList[fileInfo.Id] = make(chan models.SegMent)
	activeTaskList[fileInfo.Id] = make(chan struct{}, maxTaskSize)
	return fileInfo.Id, nil
}

func getHead(url string) *fasthttp.Request {
	head := fasthttp.RequestHeader{}
	head.SetRequestURI(url)
	head.SetUserAgent(browser.Random())
	head.SetMethod(fasthttp.MethodHead)
	request := fasthttp.AcquireRequest()
	request.Header = head
	return request
}

func getRequest(url string, start int, end int) *fasthttp.Request {
	head := fasthttp.RequestHeader{}
	head.SetRequestURI(url)
	head.SetUserAgent(browser.Random())
	head.SetByteRange(start, end)
	head.SetMethod(fasthttp.MethodGet)
	request := fasthttp.AcquireRequest()
	request.Header = head
	return request
}

func send(req *fasthttp.Request, segment models.SegMent, fileId int) {
	if segment.Complete || segment.Cache != nil {
		return
	}
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	defer fasthttp.ReleaseRequest(req)
	err := clientMap[fileId].Do(req, resp)
	if err != nil || (resp.StatusCode() != 200 && resp.StatusCode() != 206) {
		segment.Count++
		seg[fileId][segment.Index] = segment
		if segment.Count >= 3 {
			close(fileMap[fileId].Exit)
			return
		}
		// 退出当前活动队列，进行任务队列重排
		<-activeTaskList[fileId]
		taskList[fileId] <- segment
		return
	}
	segment.Cache = resp.Body()
	seg[fileId][segment.Index] = segment
	resp.ConnectionClose()
	fileMap[fileId].FileChan <- segment.Index
	return
}
