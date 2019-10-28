package common

import (
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type FileInfo struct {
	Id        string `json:"id"`
	Renewal   bool   `json:"renewal"`
	FileName  string `json:"filename"`
	FinalLink string `json:"finallink"`
	SavePath  string `json:"savepath"`
	IsExist   bool   `json:"isexist"`
	MD5       string `json:"md5"`
	FileType  string `json:"filetype"`
	Length    int64  `json:"length"`
}

// 生产gout实例
func NewClient() *http.Client {
	client := &http.Client{
		//CheckRedirect: func(req *http.Request, via []*http.Request) error {
		//	req.Header.Del("Referer")
		//	if len(via) >= 10 {
		//		return errors.New("stopped after 10 redirects")
		//	}
		//	return nil
		//},
	}
	return client
}

// 重定向获取最终链接
func redirect(url string, client *http.Client) string {
	finalLink := ""
	req := GetRequest(url)
	for {
		resp, err := client.Do(req)
		if err != nil {
			return url
		}
		defer resp.Body.Close()
		//if resp.StatusCode != 302 {
		//	return url
		//}
		finalLink = resp.Header.Get("Referrer")
		if finalLink == "" {
			finalLink = resp.Header.Get("referrer")
		}
		if finalLink != "" {
			url = finalLink
			_ = req.Body.Close()
			req = GetRequest(url)
			continue
		}
		return url
	}
}

// 请求目标文件任务信息
func GetFileInfo(finalLink string, client *http.Client) (fileInfo FileInfo, err error) {
	req := GetRequest(finalLink)
	resp, err := client.Do(req)
	if err != nil {
		return fileInfo, fmt.Errorf("请求目标文件信息失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden {
		resp, err = client.Get(finalLink)
		if err != nil {
			return fileInfo, fmt.Errorf("请求目标文件信息失败: %w", err)
		}
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fileInfo, fmt.Errorf("请求目标文件信息失败: %d", resp.StatusCode)
	}
	fileInfo.Renewal = false
	fileInfo.FinalLink = finalLink
	// 判断是否支持断点续传
	if temp := resp.Header.Get("Accept-Ranges"); len(temp) != 0 {
		fileInfo.Renewal = true
	}
	if temp := resp.Header.Get("accept-ranges"); len(temp) != 0 {
		fileInfo.Renewal = true
	}
	// 获取文件长度
	fileInfo.Length = resp.ContentLength
	// 获取文件MD5
	fileInfo.MD5 = resp.Header.Get("Content-MD5")
	// 获取文件名称
	fileInfo.FileName, fileInfo.FileType = getFileName_TypeInUrl(finalLink, resp)
	fileInfo.IsExist = FileExist(AllPath + fileInfo.FileName)
	fileInfo.SavePath = AllPath
	return
}

func GetRequest(url string) *http.Request {
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("User-Agent", browser.Chrome())
	request.Header.Set("Connection", "keep-alive")
	return request
}

func getFileName_TypeInUrl(finalLink string, response *http.Response) (fileName string, fileType string) {
	disposition := response.Header.Get("Content-Disposition")
	if disposition == "" {
		disposition = response.Header.Get("content-disposition")
	}
	if disposition != "" {
		disps := strings.Split(disposition, ";")
		for _, value := range disps {
			if strings.Contains(value, "filename=") {
				fileName = value[strings.LastIndex(value, "filename=")+1:]
				if fileName[0:0] == "\"" && fileName[len(fileName)-1:len(fileName)-1] == "\"" {
					fileName = fileName[1 : len(fileName)-2]
				}
			}
		}
	}
	if fileName == "" {
		uri, _ := url.ParseRequestURI(finalLink)
		fileName = path.Base(uri.Path)
		//u := []byte(finalLink)
		//s := strings.LastIndex(finalLink, "/")
		//if s == -1 {
		//	s = 0
		//	fileName = string(u[s:])
		//} else {
		//	fileName = string(u[s+1:])
		//	l := strings.LastIndex(fileName, "?")
		//	if l != -1 {
		//		fileName = fileName[:l]
		//	}
		//}
	}

	fileType = Http_Cotent_Type[response.Header.Get("Content-Type")]
	if fileType == "" {
		fileType = Http_Cotent_Type[response.Header.Get("content-type")]
	}
	if fileType == "" && strings.LastIndex(fileName, ".") != -1 {
		fileType = fileName[strings.LastIndex(fileName, ".")+1:]
	}
	return
}
