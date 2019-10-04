package main

import (
	"github.com/andlabs/ui"
	"math/rand"
	"time"
)

var mainwin *ui.Window

func completePage() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	return vbox
}

func settingPage() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	return vbox
}

func setUI() {
	mainwin = ui.NewWindow("下载器", 800, 480, false)
	mainwin.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		mainwin.Destroy()
		return true
	})
	tab := ui.NewTab()
	mainwin.SetChild(tab)
	tab.InsertAt("正在下载", 0, downloadingPage())
	tab.InsertAt("下载完成", 1, completePage())
	tab.InsertAt("设置", 2, settingPage())
	tab.SetMargined(0, true)
	mainwin.Show()
}

func main() {
	//url := "https://download.jetbrains.8686c.com/idea/ideaIC-2019.2.2.dmg"
	// 下载器调度启动
	rand.Seed(time.Now().Unix())
	fileQueue.New()
	go func() {
		for {
			select {
			case <-sch:
				if filenum < 2 && fileQueue.Size() > 0 {
					fileId := *fileQueue.Dequeue()
					schedule <- fileId.(int)
				}
			}
		}
	}()
	go scheduler()
	ui.Main(setUI)
}
