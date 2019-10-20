package util

import (
	"errors"
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"net/http"
	"strings"
)

type FileInfo struct {
	Id        int    `json:"id"`
	Renewal   bool   `json:"renewal"`
	FileName  string `json:"filename"`
	FinalLink string `json:"finallink"`
	SavePath  string `json:"savepath"`
	MD5       string `json:"md5"`
	FileType  string `json:"filetype"`
	Length    int64  `json:"length"`
}

// 生产gout实例
func NewClient() *http.Client {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.Header.Del("Referer")
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}
	return client
}

// 重定向获取最终链接
func redirect(url string, client *http.Client) (string, error) {
	resp, err := client.Head(url)
	if err != nil {
		return "", fmt.Errorf("请求链接失败: %w", err)
	}
	return resp.Request.URL.String(), nil
}

// 请求目标文件任务信息
func GetFileInfo(url string, client *http.Client) (fileInfo FileInfo, err error) {
	defer client.CloseIdleConnections()
	if url == "" {
		return fileInfo, errors.New("空链接")
	}
	fileInfo.Renewal = false
	finalLink, err := redirect(url, client)
	if err != nil {
		return
	}
	fileInfo.FinalLink = finalLink
	resp, err := client.Head(finalLink)
	defer resp.Body.Close()
	if err != nil {
		return fileInfo, fmt.Errorf("请求目标文件信息失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fileInfo, fmt.Errorf("请求目标文件信息失败: %d", resp.StatusCode)
	}
	// 判断是否支持断点续传
	if temp := resp.Header.Get("Accept-Ranges"); len(temp) != 0 {
		fileInfo.Renewal = true
	}
	// 获取文件长度
	fileInfo.Length = resp.ContentLength
	// 获取文件MD5
	fileInfo.MD5 = resp.Header.Get("Content-MD5")
	// 获取文件名称
	u := []byte(finalLink)
	s := strings.LastIndex(finalLink, "/")
	if s == -1 {
		s = 0
		fileInfo.FileName = string(u[s:])
	} else {
		fileInfo.FileName = string(u[s+1:])
		fileInfo.FileType = fileInfo.FileName[strings.LastIndex(fileInfo.FileName, ".")+1:]
	}
	return
}

func GetRequest(url string) *http.Request {
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("User-Agent", browser.Chrome())
	request.Header.Set("Connection", "keep-alive")
	return request
}
