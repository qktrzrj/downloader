package main

import (
	"bytes"
	"errors"
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/valyala/fasthttp"
	"io"
	"strings"
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
	u := []byte(url)
	fullName := ""
	s := strings.LastIndex(url, "/")
	if s == -1 {
		s = 0
		fullName = string(u[s:])
	} else {
		fullName = string(u[s+1:])
	}
	file, err := creatFile("./" + fullName)
	if err != nil {
		return fmt.Errorf("创建文件失败:%w", err)
	}
	_, err = io.Copy(file, bytes.NewReader(resp.Body()))
	return err
}
