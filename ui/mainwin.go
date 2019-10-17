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
	MainWin    *widgets.QMainWindow
	MainWidget *widgets.QWidget
	List       *widgets.QWidget
	ListLayout *widgets.QVBoxLayout
	ListData   []*ListItem
	ListLock   sync.Mutex
)

type taskItem struct {
	*widgets.QWidget
	download *widgets.QPushButton
	input    *widgets.QLineEdit
}

type ListItem struct {
	widget *widgets.QWidget
	bar    *widgets.QProgressBar
	op     *widgets.QPushButton
	cancel *widgets.QPushButton
}

func getAddTaskWdiget() *taskItem {
	widget := widgets.NewQWidget(MainWidget, core.Qt__WindowStaysOnTopHint)
	layout := widgets.NewQHBoxLayout()
	layout.SetSpacing(5)
	//自定义关闭，最小化，最大化按钮
	closeButton := widgets.NewQToolButton(widget)
	minButton := widgets.NewQToolButton(widget)
	maxButton := widgets.NewQToolButton(widget)
	closeIcon := MainWin.Style().StandardIcon(widgets.QStyle__SP_TitleBarCloseButton,
		widgets.NewQStyleOptionToolButton(), nil)
	closeButton.SetIcon(closeIcon)
	minIcon := MainWin.Style().StandardIcon(widgets.QStyle__SP_TitleBarMinButton,
		nil, nil)
	minButton.SetIcon(minIcon)
	maxIcon := MainWin.Style().StandardIcon(widgets.QStyle__SP_TitleBarMaxButton,
		nil, nil)
	maxButton.SetIcon(maxIcon)
	minButton.SetToolTip("最小化")
	closeButton.SetToolTip("关闭")
	maxButton.SetToolTip("最大化")
	closeButton.SetStyleSheet("background-color:transparent;")
	minButton.SetStyleSheet("background-color:transparent;")
	maxButton.SetStyleSheet("background-color:transparent;")
	closeButton.ConnectClicked(func(bool) {
		MainWin.Close()
	})
	minButton.ConnectClicked(func(bool) {
		MainWin.ShowMinimizedDefault()
	})
	size := core.QSize{}
	isMax := false
	var winPoint *core.QPoint
	maxButton.ConnectClicked(func(bool) {
		if !isMax {
			winPoint = MainWin.FrameGeometry().TopLeft()
			size.SetHeight(MainWin.Size().Height())
			size.SetWidth(MainWin.Size().Width())
			MainWin.ShowMaximizedDefault()
			MainWin.Show()
			isMax = true
			return
		}
		MainWin.Resize2(size.Width(), size.Height())
		MainWin.Move2(winPoint.X(), winPoint.Y())
		isMax = false
	})
	// 下载框
	wd := widgets.NewQWidget(widget, 0)
	wd.SetLayout(widgets.NewQHBoxLayout())
	wd.Layout().SetSpacing(0)
	// 下载按钮
	download := widgets.NewQPushButton(widget)
	// 输入框
	input := widgets.NewQLineEdit(widget)
	input.SetEchoMode(widgets.QLineEdit__Normal)
	input.SetPlaceholderText("请输入下载链接")
	wd.Layout().AddWidget(download)
	wd.Layout().AddWidget(input)
	layout.AddWidget(closeButton, 2, core.Qt__AlignLeft)
	layout.AddWidget(minButton, 0, core.Qt__AlignLeft)
	layout.AddWidget(maxButton, 0, core.Qt__AlignLeft)
	layout.AddWidget(wd, 200, core.Qt__AlignHCenter)
	widget.SetLayout(layout)
	m_bDrag := false
	var mouseStartPoint *core.QPoint
	var windowTopLeftPoint *core.QPoint
	widget.ConnectMousePressEvent(func(event *gui.QMouseEvent) {
		if event.Button() == core.Qt__LeftButton {
			m_bDrag = true
			//获得鼠标的初始位置
			mouseStartPoint = event.GlobalPos()
			//mouseStartPoint = event.pos();
			//获得窗口的初始位置
			windowTopLeftPoint = MainWin.FrameGeometry().TopLeft()
		}
	})
	widget.ConnectMouseMoveEvent(func(event *gui.QMouseEvent) {
		if m_bDrag {
			//获得鼠标移动的距离
			distanceX := event.GlobalPos().X() - mouseStartPoint.X()
			distanceY := event.GlobalPos().Y() - mouseStartPoint.Y()
			//QPoint distance = event->pos() - mouseStartPoint;
			//改变窗口的位置
			MainWin.Move2(windowTopLeftPoint.X()+distanceX, windowTopLeftPoint.Y()+distanceY)
		}
	})
	widget.ConnectMouseReleaseEvent(func(event *gui.QMouseEvent) {
		if event.Button() == core.Qt__LeftButton {
			m_bDrag = false
		}
	})
	item := &taskItem{
		QWidget:  widget,
		download: download,
		input:    input,
	}
	return item
}

func SetUI() {
	// 设置主窗口
	MainWin = widgets.NewQMainWindow(nil, 0)
	MainWin.SetMinimumSize2(800, 600)
	MainWin.SetWindowFlags(core.Qt__Window | core.Qt__FramelessWindowHint | core.Qt__WindowMinMaxButtonsHint)

	// 设置主窗口界面
	MainWidget = widgets.NewQWidget(nil, 0)
	MainWidget.SetLayout(widgets.NewQVBoxLayout())
	MainWin.SetCentralWidget(MainWidget)

	// 添加新增任务栏
	bhw := getAddTaskWdiget()
	MainWidget.Layout().AddWidget(bhw)

	// 添加列表界面
	vw := widgets.NewQWidget(MainWidget, 0)
	vw.SetMinimumSize2(780, 500)
	policy := vw.SizePolicy()
	policy.SetVerticalStretch(1)
	vw.SetSizePolicy(policy)
	vw.Layout().SetContentsMargins(0, 0, 0, 0)
	scrollArea := widgets.NewQScrollArea(vw)
	scrollArea.SetVerticalScrollBarPolicy(core.Qt__ScrollBarAlwaysOff)
	scrollArea.SetHorizontalScrollBarPolicy(core.Qt__ScrollBarAlwaysOff)
	List = widgets.NewQWidget(scrollArea, 0)
	scrollArea.SetWidget(List)
	scrollArea.SetMinimumSize2(900, 700)
	scrollArea.SetWidgetResizable(true)
	ListLayout = widgets.NewQVBoxLayout()
	ListLayout.SetSpacing(10)
	ListLayout.SetContentsMargins(0, 0, 0, 0)
	List.SetLayout(ListLayout)
	MainWidget.Layout().AddWidget(vw)

	// 添加下载事件
	bhw.download.ConnectClicked(func(bool) {
		// 检查是否有链接
		if bhw.input.Text() == "" {
			widgets.QMessageBox_Warning(MainWidget, "提示", "空链接！",
				widgets.QMessageBox__Ok, widgets.QMessageBox__Cancel)
			return
		}
		// 获取文件信息
		//g := downloader.NewGout()
		//fileInfo, err := downloader.GetFileInfo(input.Text(), g)
		//if err != nil {
		//	widgets.QMessageBox_Warning(MainWidget, "提示", fmt.Sprintf("获取文件信息失败:"+
		//		"%s", err), widgets.QMessageBox__Ok, widgets.QMessageBox__Cancel)
		//	return
		//}
		// 选择保存路径
		savePath := widgets.QFileDialog_GetSaveFileName(MainWidget, "保存", downloader.Download.SavePath+
			"example.txt", "", "txt", widgets.QFileDialog__ReadOnly)
		//fileInfo.SavePath = widgets.QFileDialog_GetSaveFileName(MainWidget, "保存",
		//	downloader.Download.SavePath+fileInfo.FileName, "", fileInfo.FileType, widgets.QFileDialog__ReadOnly)
		//fmt.Println("获取保存路径 " + fileInfo.SavePath)
		//if fileInfo.SavePath == "" {
		//	return
		//}
		// 添加任务
		//id, err := downloader.Download.AddTask(fileInfo, g)
		//if err != nil {
		//	widgets.QMessageBox_Warning(MainWidget, "提示", fmt.Sprintf("添加任务失败:"+
		//		"%s", err), widgets.QMessageBox__Ok, widgets.QMessageBox__Cancel)
		//	return
		//}
		// 添加任务项到界面
		id := downloader.TaskId(rand.Int())
		fileInfo := downloader.FileInfo{
			Renewal:   false,
			FileName:  "example.txt",
			FinalLink: bhw.input.Text(),
			SavePath:  savePath,
			MD5:       "",
			FileType:  "txt",
			Length:    1024 * 1024,
		}
		ListLock.Lock()
		item := addItem(fileInfo, id)
		ListData = append(ListData, item)
		ListLayout.AddWidget(item.widget, 0, core.Qt__AlignTop)
		List.SetLayout(ListLayout)
		ListLock.Unlock()
	})

	MainWin.Show()
}

func addItem(fileInfo downloader.FileInfo, id downloader.TaskId) *ListItem {
	item := widgets.NewQWidget(List, core.Qt__WindowStaysOnTopHint)
	item.SetLayout(widgets.NewQVBoxLayout())
	item.Layout().SetContentsMargins(0, 0, 0, 0)
	palette := item.Palette()
	palette.SetColor(gui.QPalette__All, gui.QPalette__Background, gui.NewQColor2(core.Qt__white))
	item.SetAutoFillBackground(true)
	item.SetPalette(palette)
	item.SetMinimumSize2(780, 100)
	//item.SetMaximumSize2(400, 100)
	fileName := widgets.NewQLabel2(fileInfo.FileName, item, 0)
	fileName.SetMaximumHeight(10)
	item.Layout().AddWidget(fileName)
	url := widgets.NewQLabel2(fileInfo.FinalLink, item, 0)
	url.SetMaximumHeight(10)
	item.Layout().AddWidget(url)
	bar := widgets.NewQProgressBar(item)
	bar.SetMaximumHeight(2)
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
