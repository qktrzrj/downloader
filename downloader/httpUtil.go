package downloader

import (
	"bytes"
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/valyala/fasthttp"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

var (
	strHTTP  = []byte("http")
	strHTTPS = []byte("https")
	count    = 1
)

func addMissingPort(addr string, isTLS bool) string {
	n := strings.Index(addr, ":")
	if n >= 0 {
		return addr
	}
	port := 80
	if isTLS {
		port = 443
	}
	return net.JoinHostPort(addr, strconv.Itoa(port))
}

// 构建客户端
func NewHostClient(url string) *fasthttp.HostClient {
	head := getHead(url)
	uri := head.URI()
	host := uri.Host()
	isTLS := false
	scheme := uri.Scheme()
	if bytes.Equal(scheme, strHTTPS) {
		isTLS = true
	} else if !bytes.Equal(scheme, strHTTP) {
		return nil
	}
	client := &fasthttp.HostClient{Addr: addMissingPort(string(host), isTLS), IsTLS: isTLS}
	return client
}

// 重定向
func Redirect(url string, client *fasthttp.HostClient) (finalLink string) {
	head := getHead(url)
	resp := fasthttp.AcquireResponse()
	redirectsCount := 0
	for {
		if err := client.Do(head, resp); err != nil {
			break
		}
		statusCode := resp.Header.StatusCode()
		if statusCode != fasthttp.StatusMovedPermanently &&
			statusCode != fasthttp.StatusFound &&
			statusCode != fasthttp.StatusSeeOther &&
			statusCode != fasthttp.StatusTemporaryRedirect &&
			statusCode != fasthttp.StatusPermanentRedirect {
			break
		}

		redirectsCount++
		if redirectsCount > 16 {
			break
		}
		referer := resp.Header.Peek("Referer")
		if len(referer) == 0 {
			break
		}
		uri := fasthttp.AcquireURI()
		uri.Update(url)
		uri.UpdateBytes(referer)
		url = uri.String()
		fasthttp.ReleaseURI(uri)
	}
	return url
}

// 请求目标文件任务信息
func GetFileTask(url string, client *fasthttp.HostClient) (task *Task, err error) {
	finalLink := Redirect(url, client)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req := getHead(finalLink)
	defer fasthttp.ReleaseRequest(req)
	err = client.Do(req, resp)
	if err != nil {
		return task, fmt.Errorf("请求目标文件信息失败: %w", err)
	}
	// 判断是否支持断点续传
	var renewal = false
	if temp := resp.Header.Peek("Accept-Ranges"); len(temp) != 0 {
		renewal = true
	}
	if temp := resp.Header.Peek("accept-ranges"); len(temp) != 0 {
		renewal = true
	}
	// 获取文件名称
	u := []byte(finalLink)
	fullName := ""
	s := strings.LastIndex(finalLink, "/")
	if s == -1 {
		s = 0
		fullName = string(u[s:])
	} else {
		fullName = string(u[s+1:])
	}

	task = &Task{
		id:         rand.Int(),
		renewal:    renewal,
		Status:     waiting,
		fileLength: resp.Header.ContentLength(),
		Url:        url,
		finalLink:  finalLink,
		file:       nil,
		FileName:   fullName,
		SavePath:   Downloader.SavePath,
		client:     client,
	}
	return
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

func send(req *fasthttp.Request, segment SegMent, fileId int) {
	fileInfo := readFileMap(fileId)
	if segment.Complete {
		return
	}
	if segment.Cache != nil {
		fileInfo.FileChan <- segment.Index
		return
	}
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	defer fasthttp.ReleaseRequest(req)
	err := fileInfo.Client.Do(req, resp)
	if err != nil || (resp.StatusCode() != 200 && resp.StatusCode() != 206) {
		segment.Count++
		main.updateSegIndex(fileId, segment.Index, segment)
		if segment.Count >= 3 {
			close(main.fileMap[fileId].Exit)
			return
		}
		// 退出当前活动队列，进行任务队列重排
		<-activeTaskList[fileId]
		taskList[fileId] <- segment
		return
	}
	segment.Cache = resp.Body()
	main.updateSegIndex(fileId, segment.Index, segment)
	fileInfo.FileChan <- segment.Index
	return
}
