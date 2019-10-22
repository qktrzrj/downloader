package downloader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type bt struct {
	id int
	//ctx  context.Context
	task    *Task
	request *http.Request
}

func (bt *bt) start() {
	//ctx, cancel := context.WithCancel(context.TODO())
	//timer := time.AfterFunc(10*time.Second, func() {
	//	cancel()
	//})
	//bt.request = bt.request.WithContext(ctx)
	errNum := 0
	for {
		if errNum >= 3 && bt.task.Status == Downloading {
			log.Println(fmt.Sprintf("task %d, worker %d error", bt.task.Id, bt.id))
			bt.task.Status = Errored
			close(bt.task.btCancel)
			break
		}
		select {
		case <-bt.task.btCancel:
			return
		default:
			segment := bt.task.getSeg()
			if segment == nil {
				return
			}
			err := bt.downSeg(segment, nil)
			if err != nil {
				if segment.finish != segment.start {
					segment.start = segment.finish + 1
				}
				log.Println(fmt.Sprintf("task %d, worker %d segment start %d end %d error %v",
					bt.task.Id, bt.id, segment.start, segment.end, err))
				errNum++
				bt.task.segErr(segment)
				continue
			}
		}
	}
}

func (bt *bt) downSeg(segment *SegMent, timer *time.Timer) (err error) {
	bt.request.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", segment.start, segment.end))
	response, err := bt.task.client.Do(bt.request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusPartialContent {
		return errors.New("下载错误")
	}
	reader := response.Body
	buf := bt.task.BufferPool.Get().(*bytes.Buffer)
	stream := make([]byte, 1024)
	buffLeft := 0
	for {
		//timer.Reset(2 * time.Second)
		bin := stream[:]
		buffLeft = buf.Cap() - buf.Len()
		if buffLeft < len(stream) {
			bin = stream[:buffLeft]
		}
		var l int
		select {
		case <-func() chan struct{} {
			ch := make(chan struct{})
			go func() {
				l, err = reader.Read(bin)
				if err != nil && strings.Contains(err.Error(), "EOF") {
					err = nil
				}
				close(ch)
			}()
			return ch
		}():
		case <-bt.task.btCancel:
			bufLen := int64(buf.Len())
			if bufLen > 0 {
				err = bt.task.writeToDisk(segment, buf)
				if err == nil {
					segment.finish = segment.start + bufLen - 1
				}
			}
			return
		}
		if l <= 0 {
			if buf.Len() > 0 {
				writeErr := bt.task.writeToDisk(segment, buf)
				buf.Reset() // 重置缓冲区
				if writeErr != nil {
					err = writeErr
				}
			}
			break
		}
		buf.Write(bin[:l])
		atomic.AddInt64(&bt.task.DownloadCount, int64(l))
		if buf.Len() == buf.Cap() || err == io.EOF { // 缓存满了, 或者流尾, 写入磁盘
			writeErr := bt.task.writeToDisk(segment, buf)
			buf.Reset() // 重置缓冲区
			if writeErr != nil {
				err = writeErr
				break
			}
		}
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			break
		}
	}
	return
}
