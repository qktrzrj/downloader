package downloader

import (
	"errors"
	"fmt"
	"github.com/guonaihong/gout"
	"net/http"
	url2 "net/url"
	"strings"
	"time"
)

type FileInfo struct {
	Renewal   bool
	FileName  string
	FinalLink string
	SavePath  string
	MD5       string
	FileType  string
	Length    int64
}

// 生产gout实例
func NewGout() *gout.Gout {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.Header.Del("Referer")
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
		Timeout: time.Second * 10,
	}
	return gout.New(client)
}

// 重定向获取最终链接
func redirect(url string, g *gout.Gout) (string, error) {
	resp, err := g.Head(url)
	defer resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("请求链接失败: %w", err)
	}
	return resp.Request.URL.String(), nil
}

// 请求目标文件任务信息
func GetFileInfo(url string, g *gout.Gout) (fileInfo FileInfo, err error) {
	fileInfo.Renewal = false
	finalLink, err := redirect(url, g)
	if err != nil {
		return
	}
	resp, err := g.Head(finalLink)
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
	parse, _ := url2.Parse(url)
	u := []byte(parse.Path)
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
