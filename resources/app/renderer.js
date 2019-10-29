let ipcRenderer = require('electron').ipcRenderer
const {remote} = require('electron')
const log = require('electron-log');
log.transports.console.level = false;
log.transports.file.file = "./data/log/log.log";
log.transports.console.level = 'silly';

const app = remote.app

let zoom = document.getElementById('zoom')
let minimize = document.getElementById('minimize')
let close = document.getElementById('close')
let add = document.getElementById('add')
let input = document.getElementById('input')
let download = document.getElementById('download')
let socket = new Map()
let savePath = new Map()
let ws
let taskStatusMap = new Map([
    [1, 'Waiting'],
    [2, 'Downloading'],
    [3, 'Success'],
    [4, 'Paused'],
    [5, 'Errored']
])

ipcRenderer.send('runExec')

ipcRenderer.once('websocket', function () {
    ws = new WebSocket("ws://localhost:4800/main")
    //连接打开时触发
    ws.onopen = function (evt) {
        log.info("Connection open ...")
    }

    ws.onmessage = function (evt) {
        let data
        let err
        try {
            data = JSON.parse(evt.data)
        } catch (e) {
            log.error("json parse err:" + e)
            err = true
        }
        if (err === true) {
            return
        }
        if (data.op === 0 || data.op === undefined) {
            ipcRenderer.send('loading')
            initTask(data)
            ipcRenderer.send('load-close')
        }
        if (data.op === 1 || data.op === 4 || data.op === 5) {
            ipcRenderer.send('set-close')
            return
        }
        if (data.op === 2) {
            ipcRenderer.send('set-min')
            return
        }
        if (data.op === 6) {
            remote.dialog.showOpenDialog({
                defaultPath: data.savePath,
                properties: ['openDirectory']
            }).then(function (res) {
                if (!res.canceled) {
                    if (res.filePaths[0].lastIndexOf('/') !== -1) {
                        data.savePath = res.filePaths[0] + '/'
                    } else {
                        data.savePath = res.filePaths[0] + '\\'
                    }
                    ws.send(JSON.stringify(data))
                }
            })
            return
        }
        if (data.op === 7) {
            ipcRenderer.send('file-close')
            return
        }
        if (data.op === 8) {
            ipcRenderer.send('file-min')
            return
        }
        if (data.op === 9) {
            ipcRenderer.send('file-close')
            addTask(data.fileList)
        }
    }

//连接关闭时触发
    ws.onclose = function (evt) {
        remote.dialog.showErrorBox('错误', '服务崩溃')
        ipcRenderer.send('window-close')
    }
})

function connect(id) {
    socket[id] = new WebSocket("ws://localhost:4800/getTaskInfo?id=" + id)
    let item = document.getElementById(id)
    let bar = document.getElementById(id + 'progress')
    let state = document.getElementById(id + 'filestatus')
    let size = document.getElementById(id + 'downsize')
    let op = document.getElementById(id + 'item-op')
    //连接打开时触发
    socket[id].onopen = function (evt) {
        log.info("Connection open ...")
    }
    //接收到消息时触发
    socket[id].onmessage = function (evt) {
        log.info("Received Message: " + evt.data)
        let err = false
        let data
        try {
            data = JSON.parse(evt.data)
        } catch (e) {
            log.info("json parse err:" + e)
            err = true
        }
        if (err === true) {
            return
        }
        let downsize = 0
        if (data.downloadCount >= 1024 * 1024) {
            downsize = parseInt(data.downloadCount / (1024 * 1024)) + ' MB '
        } else {
            downsize = parseInt(data.downloadCount / 1024) + ' KB '
        }
        bar.value = parseInt(data.downloadCount * 100 / data.filelength)
        size.innerText = downsize + size.innerText.slice(size.innerText.indexOf('Of'), size.innerText.length)
        if (data.status === 3 && state.innerText !== 'Success') {
            getFileIcon(id, op)
            item.removeChild(bar)
            let count = document.getElementById(id + 'filenum')
            count.innerText = ''
            size.innerText = size.innerText.slice(size.innerText.indexOf('Of') + 3, size.innerText.length)
            state.innerText = taskStatusMap.get(data.status)
            state.style.display = 'none'
            state.hidden = true
            state.setAttribute('hidden', true)
            let filename = document.getElementById(id + 'filename')
            let myNotification = new Notification('下载完成', {
                body: '文件' + filename.innerText + '下载完成！'
            })
            myNotification.onclick = () => {
                myNotification.remove()
            }
            socket[id].close()
        }
        if (data.status === 1 || data.status === 2) {
            op.src = './icon/c_pau.png'
        }
        if (data.status === 4 || data.status === 5) {
            op.src = './icon/play1.png'
            if (data.status === 5 && state.innerText !== 'Errored') {
                state.innerText = taskStatusMap.get(data.status)
                let filename = document.getElementById(id + 'filename')
                let myNotification = new Notification('下载失败', {
                    body: '文件' + filename.innerText + '下载失败！'
                })
                myNotification.onclick = () => {
                    myNotification.remove()
                }
            }
            socket[id].close()
        }
        state.innerText = taskStatusMap.get(data.status)
    }
    //连接关闭时触发
    socket[id].onclose = function (evt) {
        saveUI()
        delete socket[id]
        // if (state.innerText === 'Downloading' || state.innerText === 'Waiting') {
        //     state.innerText = 'Errored'
        //     op.src = './icon/play1.png'
        //     saveUI()
        // }
    }
    //连接发生错误时触发
    socket[id].onerror = function (evt) {
        //如果出现连接、处理、接收、发送数据失败的时候触发onerror事件
        if (state.innerText !== 'Success') {
            state.innerText = 'Errored'
            saveUI()
        }
        log.error(error);
    }
}

function initTask(html) {
    download.innerHTML = html
    let items = document.getElementsByClassName('item')
    for (let i = 0; i < items.length; i++) {
        savePath[items[i].id] = items[i].value
        let status = document.getElementById(items[i].id + 'filestatus')
        let op = document.getElementById(items[i].id + 'item-op')
        if (status.innerText !== 'Success' && !status.hidden) {
            connect(items[i].id)
            op.src = './icon/play1.png'
            if (status.innerText !== 'Errored') {
                status.innerText = 'Paused'
            }
        }
        addListener(items[i].id)
    }
}


function additem(id, name, size) {
    let itemdiv = document.createElement('div')
    itemdiv.className = 'item'
    itemdiv.id = id
    itemdiv.value = savePath[id]
    download.insertBefore(itemdiv, download.firstChild)
    let itemopdiv = document.createElement('div')
    itemopdiv.style = 'height: 50px;width: 50px;float: left'
    itemdiv.appendChild(itemopdiv)
    let itemop = document.createElement('img')
    itemop.className = 'item-op'
    itemop.id = id + 'item-op'
    itemop.src = './icon/c_pauD.png'
    itemopdiv.appendChild(itemop)
    let itemsp = document.createElement('div')
    itemsp.style = 'float:left;margin-left: auto;width: 1px;height: 100%;background: darkgray;'
    itemdiv.appendChild(itemsp)
    let iteminfo = document.createElement('div')
    iteminfo.className = 'item-info'
    iteminfo.id = id + 'item-info'
    itemdiv.appendChild(iteminfo)
    let filename = document.createElement('span')
    filename.className = 'filename'
    filename.id = id + 'filename'
    filename.innerText = name
    iteminfo.appendChild(filename)
    let br = document.createElement('br')
    filename.appendChild(br)
    let filenum = document.createElement('span')
    filenum.className = 'filenum'
    filenum.id = id + 'filenum'
    iteminfo.appendChild(filenum)
    let downsize = document.createElement('span')
    downsize.className = 'downsize'
    downsize.id = id + 'downsize'
    if (size >= 1024 * 1024) {
        downsize.innerText = '0 KB Of ' + parseInt(size / (1024 * 1024)) + ' MB'
    } else {
        downsize.innerText = '0 KB Of ' + parseInt(size / 1024) + ' KB'
    }
    iteminfo.appendChild(downsize)
    let filestatus = document.createElement('span')
    filestatus.className = 'filestatus'
    filestatus.id = id + 'filestatus'
    filestatus.innerText = 'waiting'
    itemdiv.appendChild(filestatus)
    let progress = document.createElement('progress')
    progress.className = 'progress'
    progress.id = id + 'progress'
    progress.value = 0
    progress.max = 100
    itemdiv.appendChild(progress)
    let itemex = document.createElement('div')
    itemex.className = 'item-ex'
    itemex.id = id + 'item-ex'
    itemdiv.appendChild(itemex)
    let del = document.createElement('img')
    del.className = 'item-delete'
    del.id = id + 'item-delete'
    del.src = 'icon/delete.png'
    itemdiv.appendChild(del)
    countuner()
    saveUI()
}


function addListener(id) {
    let op = document.getElementById(id + 'item-op')
    let itemex = document.getElementById(id + 'item-ex')
    let del = document.getElementById(id + 'item-delete')
    let state = document.getElementById(id + 'filestatus')

    op.addEventListener('click', () => {
        if (state.innerText === 'Downloading' || state.innerText === 'Waiting') {
            operate(id, 1)
            return
        }
        if (state.innerText !== 'Success') {
            operate(id, 2)
            connect(id)
            return
        }
        remote.dialog.showOpenDialog({defaultPath: savePath[id], properties: ['openFile']})
    })

    itemex.addEventListener('mouseover', () => {
        del.style.display = 'block'
    })

    itemex.addEventListener('mouseleave', () => {
        del.style.display = 'none'
    })

    del.addEventListener('mouseover', () => {
        del.style.display = 'block'
    })

    del.addEventListener('mouseleave', () => {
        del.style.display = 'none'
    })

    del.addEventListener('click', () => {
        if (del.src !== '') {
            let statu = state.innerText
            let item = document.getElementById(id)
            download.removeChild(item)
            if (statu !== 'Success') {
                operate(id, 3)
                return
            }
            operate(id, 4)
        }
    })
}

function countuner() {
    let count = document.getElementsByClassName('filenum')
    let c = 1
    for (let i = 0; i < count.length; i++) {
        if (count[i].parentElement.parentElement.hidden) {
            continue
        }
        let state = document.getElementById(count[i].parentElement.parentElement.id + 'filestatus')
        if (state.innerText === 'Success' || state.hidden) {
            continue
        }
        count[i].innerText = '[' + c + ']'
        c++
    }
}

function removeItem(id) {
    let item = document.getElementById(id)
    download.removeChild(item)
    countuner()
}

function hiddenItem(id) {
    let item = document.getElementById(id)
    item.hidden = true
}

function displayitem(id) {
    let item = document.getElementById(id)
    item.hidden = false
}

function mouseListener(ac) {
    if (ac === 1) {
        close.setAttribute('src', './icon/close-rollover.png');
        minimize.setAttribute('src', './icon/minimize-rollover.png')
        zoom.setAttribute('src', './icon/zoom-rollover.png')
        return
    }
    close.setAttribute('src', './icon/close.png');
    minimize.setAttribute('src', './icon/minimize.png')
    zoom.setAttribute('src', './icon/zoom.png')
}

if (close) {
    close.addEventListener('mouseover', () => mouseListener(1))
    close.addEventListener('mouseleave', () => mouseListener(2))
    close.addEventListener('click', () => {
        close.setAttribute('src', './icon/close-pressed.png')
        saveUI()
        ipcRenderer.send('window-close')
    })
}

if (minimize) {
    minimize.addEventListener('mouseover', () => mouseListener(1))
    minimize.addEventListener('mouseleave', () => mouseListener(2))
    minimize.addEventListener('click', () => {
        // ipcRenderer.send('window-minimize')
        minimize.setAttribute('src', './icon/minimize-pressed.png')
        remote.getCurrentWindow().minimize()
        minimize.setAttribute('src', './icon/minimize.png')
    })
}

if (zoom) {
    zoom.addEventListener('mouseover', () => mouseListener(1))
    zoom.addEventListener('mouseleave', () => mouseListener(2))
    zoom.addEventListener('click', () => {
        zoom.setAttribute('src', './icon/zoom-pressed.png')
        // ipcRenderer.send('window-zoom')
        if (remote.getCurrentWindow().isMaximized()) {
            remote.getCurrentWindow().unmaximize()
        } else {
            remote.getCurrentWindow().maximize()
        }
        zoom.setAttribute('src', './icon/zoom.png')
    })
}

if (add) {
    add.addEventListener('click', () => {
        addfunc()
    })
    input.addEventListener('keypress', (e) => {
        if (e.keyCode === 13) {
            addfunc()
        }
    })
}

function addfunc() {
    if (input.value === '') {
        remote.dialog.showErrorBox('错误', '无效的url')
        return
    }
    ipcRenderer.send('loading')
    fetch('http://localhost:4800/getFileInfo?url=', {
        method: 'POST',
        body: JSON.stringify(input.value)
    }).then(function (res) {
        input.value = ''
        ipcRenderer.send('load-close')
        if (res.status !== 200) {
            remote.dialog.showErrorBox('错误', '服务异常')
            return
        }
        res.json().then((data) => {
            if (data !== "" && data.code === 1) {
                ipcRenderer.send('download')
                return
            }
            remote.dialog.showErrorBox('错误', data.msg)
        })
    }).catch(function (error) {
        ipcRenderer.send('load-close')
        remote.dialog.showErrorBox('错误', '请求失败:' + error.message)
    })
}

function operate(id, event) {
    saveUI()
    let data = {
        taskid: id,
        enum: event,
    }
    fetch('http://localhost:4800/operate', {
        method: 'POST',
        body: JSON.stringify(data),
        headers: new Headers({
            'Content-Type': 'application/json'
        })
    }).then(function (res) {
        if (res.status !== 200) {
            remote.dialog.showErrorBox('错误', '服务异常')
            return
        }
        res.json().then((data) => {
            if (data === "" || data.code !== 1) {
                remote.dialog.showErrorBox('错误', data.msg)
            }
        })
    }).catch(function (error) {
        remote.dialog.showErrorBox('错误', '请求失败:' + error.message)
    })
}

function getFileIcon(id, op) {
    app.getFileIcon(savePath[id], {size: "normal"}).then(function (icon) {
        let buffer = icon.toPNG();
        let fs = require('fs');
        let tmpFile = './resources/app/tmp/' + id + '.png';
        let writerStream = fs.createWriteStream(tmpFile);
        writerStream.write(buffer);
        writerStream.end();  //标记文件末尾  结束写入流，释放资源
        writerStream.on('finish', function () {
            op.src = './tmp/' + id + '.png'
            log.info("写入完成。");
        });
        writerStream.on('error', function (error) {
            op.src = './icon/unkown.png'
        });
    })
}


function addTask(fileList) {
    for (let i = 0; i < fileList.length; i++) {
        let filename = fileList[i].filename
        let length = fileList[i].length
        let savepath = fileList[i].savepath
        let isexist = fileList[i].isexist
        if (isexist) {
            const options = {
                title: '保存文件',
                defaultPath: savepath + filename,
                filters: [
                    {name: 'All Files', extensions: ['*']}
                ]
            }
            let filePath = remote.dialog.showSaveDialogSync(options)
            if (filePath === undefined || filePath === "") {
                return
            }
            let index = filePath.lastIndexOf('/')
            if (index === -1) {
                index = filePath.lastIndexOf('\\')
            }
            fileList[i].savepath = filePath.slice(0, index + 1)
            fileList[i].filename = filePath.slice(index + 1, filePath.length)
        }
        newTask(fileList[i])
    }
}

function newTask(info) {
    fetch('http://localhost:4800/addTask', {
        method: 'POST',
        body: JSON.stringify(info),
        headers: new Headers({
            'Content-Type': 'application/json'
        })
    }).then(function (res) {
        if (res.status !== 200) {
            remote.dialog.showErrorBox('错误', '服务异常')
            return
        }
        res.json().then((data) => {
            if (data !== "" && data.code === 1) {
                savePath[data.data] = info.savepath + info.filename
                additem(data.data, info.filename, info.length)
                addListener(data.data)
                connect(data.data)
                return
            }
            remote.dialog.showErrorBox('错误', data.msg)
        })
    }).catch(function (error) {
        remote.dialog.showErrorBox('错误', '请求失败:' + error.message)
    })
}

function saveUI() {
    fetch('http://localhost:4800/saveUI', {
        method: 'POST',
        body: JSON.stringify(download.innerHTML),
        headers: new Headers({
            'Content-Type': 'application/json'
        })
    }).catch(function (error) {
        log.info(error)
    })
}

