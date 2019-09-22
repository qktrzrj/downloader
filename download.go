package main

import (
	"errors"
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/valyala/fasthttp"
)

func Download(url string) error {
	// 构建请求头
	head := fasthttp.RequestHeader{}
	head.SetRequestURI(url)
	head.SetUserAgent(browser.Chrome())
	head.SetMethod("GET")
	request := &fasthttp.Request{
		Header: head,
	}
	client := fasthttp.Client{}
	resp := fasthttp.Response{}
	err := client.Do(request, &resp)
	if err != nil {
		return fmt.Errorf("获取下载连接失败:%w", err)
	}
	if resp.StatusCode() != 200 {
		return errors.New("连接失败！")
	}
	// 创建文件

	return nil
}
