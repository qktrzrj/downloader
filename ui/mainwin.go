package ui

import (
	"downloader/downloader"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"math/rand"
	"sync"
)

var (
	MainWin  *widgets.QMainWindow
	Widget   *widgets.QWidget
	List     *widgets.QWidget
	ListData []*ListItem
	ListLock sync.Mutex
)

type ListItem struct {
	widget *widgets.QWidget
	bar    *widgets.QProgressBar
	op     *widgets.QPushButton
	cancel *widgets.QPushButton
}

func SetUI() {
	// 设置主窗口
	MainWin = widgets.NewQMainWindow(nil, 0)
	MainWin.SetMinimumSize2(400, 600)
	MainWin.SetWindowTitle("Downloader")

	// 设置主窗口界面
	Widget = widgets.NewQWidget(nil, 0)
	Widget.SetLayout(widgets.NewQVBoxLayout())
	MainWin.SetCentralWidget(Widget)

	// 添加新增任务按钮
	bhw := widgets.NewQWidget(Widget, core.Qt__WindowStaysOnTopHint)
	bhw.SetLayout(widgets.NewQHBoxLayout())
	input := widgets.NewQLineEdit(bhw)
	input.SetEchoMode(widgets.QLineEdit__Normal)
	input.SetPlaceholderText("请输入下载链接")
	bhw.Layout().AddWidget(input)
	download := widgets.NewQPushButton2("下载", nil)
	bhw.Layout().AddWidget(download)
	Widget.Layout().AddWidget(bhw)

	// 添加列表界面
	vw := widgets.NewQWidget(Widget, 0)
	vw.SetMinimumSize2(400, 500)
	vw.Move2(200, 350)
	scrollArea := widgets.NewQScrollArea(vw)
	scrollArea.SetVerticalScrollBarPolicy(core.Qt__ScrollBarAlwaysOff)
	scrollArea.SetHorizontalScrollBarPolicy(core.Qt__ScrollBarAlwaysOff)
	List = widgets.NewQWidget(scrollArea, 0)
	scrollArea.SetWidget(List)
	//scrollArea.SetWidgetResizable(true)
	List.SetLayout(widgets.NewQVBoxLayout())
	//vw.Layout().AddWidget(List)
	Widget.Layout().AddWidget(vw)

	// 添加下载事件
	download.ConnectClicked(func(bool) {
		// 检查是否有链接
		if input.Text() == "" {
			widgets.QMessageBox_Warning(Widget, "提示", "空链接！",
				widgets.QMessageBox__Ok, widgets.QMessageBox__Cancel)
			return
		}
		// 获取文件信息
		//g := downloader.NewGout()
		//fileInfo, err := downloader.GetFileInfo(input.Text(), g)
		//if err != nil {
		//	widgets.QMessageBox_Warning(Widget, "提示", fmt.Sprintf("获取文件信息失败:"+
		//		"%s", err), widgets.QMessageBox__Ok, widgets.QMessageBox__Cancel)
		//	return
		//}
		// 选择保存路径
		savePath := widgets.QFileDialog_GetSaveFileName(Widget, "保存", downloader.Download.SavePath+
			"example.txt", "", "txt", widgets.QFileDialog__ReadOnly)
		//fileInfo.SavePath = widgets.QFileDialog_GetSaveFileName(Widget, "保存",
		//	downloader.Download.SavePath+fileInfo.FileName, "", fileInfo.FileType, widgets.QFileDialog__ReadOnly)
		//fmt.Println("获取保存路径 " + fileInfo.SavePath)
		//if fileInfo.SavePath == "" {
		//	return
		//}
		// 添加任务
		//id, err := downloader.Download.AddTask(fileInfo, g)
		//if err != nil {
		//	widgets.QMessageBox_Warning(Widget, "提示", fmt.Sprintf("添加任务失败:"+
		//		"%s", err), widgets.QMessageBox__Ok, widgets.QMessageBox__Cancel)
		//	return
		//}
		// 添加任务项到界面
		id := downloader.TaskId(rand.Int())
		fileInfo := downloader.FileInfo{
			Renewal:   false,
			FileName:  "example.txt",
			FinalLink: input.Text(),
			SavePath:  savePath,
			MD5:       "",
			FileType:  "txt",
			Length:    1024 * 1024,
		}
		ListLock.Lock()
		item := addItem(fileInfo, id)
		ListData = append(ListData, item)
		List.Layout().AddWidget(item.widget)
		ListLock.Unlock()
	})

	MainWin.Show()
}

func addItem(fileInfo downloader.FileInfo, id downloader.TaskId) *ListItem {
	item := widgets.NewQWidget(List, 0)
	item.SetLayout(widgets.NewQVBoxLayout())
	palette := gui.NewQPalette()
	palette.SetColor(gui.QPalette__Current, gui.QPalette__Background, gui.NewQColor2(core.Qt__white))
	fileName := widgets.NewQLabel2(fileInfo.FileName, nil, 0)
	item.Layout().AddWidget(fileName)
	url := widgets.NewQLabel2(fileInfo.FinalLink, nil, 0)
	item.Layout().AddWidget(url)
	bar := widgets.NewQProgressBar(item)
	item.Layout().AddWidget(bar)
	hw := widgets.NewQWidget(item, 0)
	hw.SetLayout(widgets.NewQHBoxLayout())
	item.Layout().AddWidget(hw)
	operation := widgets.NewQPushButton2("继续", nil)
	cancel := widgets.NewQPushButton2("取消", nil)
	hw.Layout().AddWidget(operation)
	hw.Layout().AddWidget(cancel)
	listItem := &ListItem{
		widget: item,
		bar:    bar,
		op:     operation,
		cancel: cancel,
	}
	return listItem
}
