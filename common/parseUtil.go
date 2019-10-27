package common

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"sync"
)

func GetFileList(url string) (fileList []FileInfo, err error) {
	client := NewClient()
	defer client.CloseIdleConnections()
	finalLink := redirect(url, client)
	req := GetRequest(finalLink)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求目标文件信息失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 206 {
		return nil, fmt.Errorf("请求目标文件信息失败: %d", resp.StatusCode)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = resp.Header.Get("content-type")
	}
	if contentType != "" {
		if _, ok := Http_Cotent_Type[contentType]; !ok {
			fileInfo, err := GetFileInfo(finalLink, client)
			if err != nil {
				return nil, err
			}
			fileList = append(fileList, fileInfo)
			return fileList, nil
		}
	}
	//resp, _ = client.Get(finalLink)
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("请求目标文件信息失败: %w", err)
	}
	var lock sync.Mutex
	var group sync.WaitGroup
	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		group.Add(1)
		go func() {
			// 解析a标签
			link, exists := selection.Attr("href")
			if exists {
				if strings.Index(link, "http") == -1 {
					link = resp.Request.URL.Scheme + "://" + resp.Request.Host + link
				}
				fileInfo, err := GetFileInfo(link, client)
				if err == nil {
					lock.Lock()
					fileList = append(fileList, fileInfo)
					lock.Unlock()
				}
				err = nil
			}
			group.Done()
		}()
	})
	group.Wait()
	return
}
