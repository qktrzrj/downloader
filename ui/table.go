package ui

import (
	"downloader/downloader"
	"github.com/andlabs/ui"
	"strconv"
)

var statusCando map[int]string = map[int]string{0: "暂停", 1: "继续"}

type modelHandler struct {
	lab    int
	colNum int
}

func (modelhandler *modelHandler) ColumnTypes(m *ui.TableModel) []ui.TableValue {
	tValue := make([]ui.TableValue, modelhandler.colNum)
	maxCol := modelhandler.colNum
	if modelhandler.lab == 0 {
		maxCol -= 3
		tValue[4] = ui.TableString("")
		tValue[5] = ui.TableString("")
		tValue[6] = ui.TableString("")
	}
	for idx := 0; idx < maxCol; idx++ {
		tValue[idx] = ui.TableString("") // Init strings columns
	}
	return tValue
}

func (modelhandler *modelHandler) NumRows(m *ui.TableModel) int {
	if modelhandler.lab == 1 {
		return len(downloader.Downloader.ActiveTaskMap)
	}
	return len(downloader.Downloader.CompleteTaskMap)
}

func (modelhandler *modelHandler) CellValue(m *ui.TableModel, row, column int) ui.TableValue {
	maxCol := modelhandler.colNum
	task := downloader.Downloader
	if modelhandler.lab == 0 {
		maxCol -= 3
		switch column {
		case 4:
			return ui.TableInt(main.barList[modelhandler.RowFileId[row]])
		case 5:
			return ui.TableString(statusCando[fileInfo.Status])
		case 6:
			return ui.TableString("取消")
		}
	}
	if column < modelhandler.colNum && row < maxCol {
		switch column {
		case 0:
			return ui.TableString(strconv.Itoa(row + 1))
		case 1:
			return ui.TableString(fileInfo.FileName)
		case 2:
			return ui.TableString(fileInfo.FilePath)
		case 3:
			return ui.TableString(fileInfo.Url)
		}
	}
	return nil
}

func (modelhandler *modelHandler) SetCellValue(m *ui.TableModel, row, column int, value ui.TableValue) {
	maxCol := modelhandler.colNum
	fileId := modelhandler.RowFileId[row]
	fileInfo := main.readFileMap(fileId)
	if modelhandler.lab == 0 {
		maxCol -= 3
	}
	if column < modelhandler.colNum && row < maxCol {
		switch column {
		case 5:
			if fileInfo.Status == 0 {
				// 暂停
				fileInfo.Status = 1
				fileInfo.Continue = make(chan struct{})
				close(fileInfo.Pause)
			} else {
				// 继续下载
				fileInfo.Status = 0
				fileInfo.Pause = make(chan struct{})
				close(fileInfo.Continue)
			}
			main.updateFileMap(fileId, fileInfo)
		}
	}
}

func newModelHandler(lab, col int) *modelHandler {
	modelh := new(modelHandler)
	modelh.lab = lab
	modelh.colNum = col
	modelh.RowFileId = make([]int, 0)
	return modelh
}

func newDownRow(url string) {

}
