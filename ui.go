package main

import "github.com/andlabs/ui"

var (
	MainWin       *ui.Window
	UIChangeEvent *UIEvent
)

type UIEvent struct {
	DpInsert chan int
	DpRemove chan int
	DpChange chan int
	CpInsert chan int
	CpRemove chan int
	CpChange chan int
}

func SetUI() func() {
	return setUI
}

func setUI() {
	MainWin = ui.NewWindow("下载器", 1000, 480, true)
	MainWin.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		MainWin.Destroy()
		return true
	})
	tab := ui.NewTab()
	MainWin.SetChild(tab)
	tab.InsertAt("正在下载", 0, downloadingPage())
	tab.InsertAt("下载完成", 1, completePage())
	tab.InsertAt("设置", 2, settingPage())
	tab.SetMargined(0, true)
	MainWin.Show()
}
