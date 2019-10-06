package ui

import "github.com/andlabs/ui"

var MainWin *ui.Window

func SetUI() func() {
	return setUI
}

func setUI() {
	MainWin = ui.NewWindow("下载器", 800, 480, false)
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
