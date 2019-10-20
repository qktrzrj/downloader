package downloader

import (
	"bytes"
	"downloader/util"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
)

type bt struct {
	id int
	//ctx  context.Context
	task *Task
}

func (bt *bt) start() {
	errNum := 0
	for {
		if errNum >= 3 && bt.task.Status == Downloading {
			log.Println(fmt.Sprintf("task %d, worker %d error", bt.task.id, bt.id))
			bt.task.Status = Errored
			go bt.task.Exit()
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
			err := bt.downSeg(segment)
			if err != nil {
				if segment.finish != segment.start {
					segment.start = segment.finish + 1
				}
				log.Println(fmt.Sprintf("task %d, worker %d segment start %d end %d error",
					bt.task.id, bt.id, segment.start, segment.end))
				errNum++
				bt.task.segErr(segment)
				continue
			}
		}
	}
}

func (bt *bt) downSeg(segment *SegMent) (err error) {
	request := util.GetRequest(bt.task.finalLink, segment.start, segment.end)
	response, err := bt.task.client.Do(request)
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
ReadStream:
	for {
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
				close(ch)
			}()
			return ch
		}():
		case <-bt.task.btCancel:
			bufLen := int64(buf.Len())
			if bufLen > 0 {
				err = bt.task.writeToDisk(segment, buf)
				if err == nil {
					segment.finish = segment.start + bufLen
				}
			} else {
				err = errors.New("canceled")
			}
			break ReadStream
		}
		if l > 0 {
			buf.Write(bin[:l])
			atomic.AddInt64(&bt.task.DownloadCount, int64(l))
			if buf.Len() == buf.Cap() || err == io.EOF { // 缓存满了, 或者流尾, 写入磁盘
				bufLen := int64(buf.Len())
				writeErr := bt.task.writeToDisk(segment, buf)
				if writeErr != nil {
					err = writeErr
					break
				}
				buf.Reset()                             // 重置缓冲区
				segment.finish = segment.start + bufLen // 片段写入磁盘偏移量
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
