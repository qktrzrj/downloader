let ipcRenderer = require('electron').ipcRenderer
const {remote} = require('electron')


let zoom = document.getElementById('zoom')
let minimize = document.getElementById('minimize')
let close = document.getElementById('close')
let add = document.getElementById('add')
let input = document.getElementById('input')
let download = document.getElementById('download')
let socket = []
let taskStatusMap = new Map([
    [1, 'Waiting'],
    [2, 'Downloading'],
    [3, 'Success'],
    [4, 'Paused'],
    [5, 'Errored']
])

let ws = new WebSocket("ws://localhost:4800/checkActive")
//连接打开时触发
ws.onopen = function (evt) {
    console.log("Connection open ...")
}

//连接关闭时触发
ws.onclose = function (evt) {
    remote.dialog.showErrorBox('错误', '服务崩溃')
    ipcRenderer.send('window-close')
}

function connect(id) {
    socket[id] = new WebSocket("ws://localhost:4800/getTaskInfo?id=" + id)
    let bar = document.getElementById(id + 'progress')
    let state = document.getElementById(id + 'filestatus')
    let size = document.getElementById(id + 'downsize')
    //连接打开时触发
    socket[id].onopen = function (evt) {
        console.log("Connection open ...")
    }
    //接收到消息时触发
    socket[id].onmessage = function (evt) {
        console.log("Received Message: " + evt.data)
        let data = JSON.parse(evt.data)
        let downsize = 0
        if (data.downloadCount >= 1024 * 1024) {
            downsize = parseInt(data.downloadCount / (1024 * 1024)) + ' MB '
        } else {
            downsize = parseInt(data.downloadCount / 1024) + ' KB '
        }
        bar.value = parseInt(data.downloadCount * 100 / data.filelength)
        state.innerText = taskStatusMap.get(data.status)
        size.innerText = downsize + size.innerText.slice(size.innerText.indexOf('Of'), size.innerText.length)
    }
    //连接关闭时触发
    socket[id].onclose = function (evt) {
        if (state.innerText === 'Errored' || state.innerText === 'Success') {
            return
        }
        if (state.innerText === 'Downloading' || state.innerText === 'Waiting') {
            state.innerText = 'Errored'
            return
        }
        remote.dialog.showErrorBox('错误', '服务异常')
        ipcRenderer.send('window-close')
    }
    //连接发生错误时触发
    socket[id].onerror = function (evt) {
        //如果出现连接、处理、接收、发送数据失败的时候触发onerror事件
        console.log(error);
    }
}


function additem(id, name, size) {
    let itemdiv = document.createElement('div')
    itemdiv.className = 'item'
    itemdiv.id = id
    download.appendChild(itemdiv)
    let itemop = document.createElement('img')
    itemop.className = 'item-op'
    itemop.id = id + 'item-op'
    itemop.src = './icon/c_pauD.png'
    itemdiv.appendChild(itemop)
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
    del.src = ''
    download.appendChild(del)
    countuner()
}


function addListener(id) {
    let op = document.getElementById(id + 'item-op')
    let item = document.getElementById(id + 'item-ex')
    let del = document.getElementById(id + 'item-delete')
    let state = document.getElementById(id + 'filestatus')

    op.addEventListener('click', () => {
        if (state === 'Downloading' || state === 'Waiting') {
            operate(id, 1)
            return
        }
        if (state !== 'Success') {
            operate(id, 2)
        }
        operate(id, 5)
    })

    item.addEventListener('mouseover', () => {
        del.src = 'icon/delete.png'
    })

    item.addEventListener('mouseleave', () => {
        del.src = ''
    })

    del.addEventListener('mouseover', () => {
        del.src = 'icon/delete.png'
    })

    del.addEventListener('mouseleave', () => {
        del.src = ''
    })

    del.addEventListener('click', () => {
        if (del.src !== '') {
            if (state !== 'Success') {
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
    input.addEventListener('submit', () => {
        //if (e.keyCode === 13) {
        addfunc()
        //}
    })
}

function addfunc() {
    if (input.value === '') {
        remote.dialog.showErrorBox('错误', '无效的url')
        return
    }
    fetch('http://localhost:4800/getFileInfo?url=' + input.value).then(function (res) {
        input.value = ''
        if (res.status !== 200) {
            remote.dialog.showErrorBox('错误', '服务异常')
            return
        }
        res.json().then((data) => {
            if (data !== "" && data.code === 1) {
                let filename = data.data.filename
                let length = data.data.length
                let savepath = data.data.savepath
                const options = {
                    title: '保存文件',
                    defaultPath: savepath + filename,
                    filters: [
                        {name: 'All Files', extensions: ['*']}
                    ]
                }
                let path = remote.dialog.showSaveDialogSync(options)
                if (path === undefined) {
                    return
                }
                data.data.savepath = path.slice(0, path.lastIndexOf('/') + 1)
                data.data.filename = path.slice(path.lastIndexOf("/") + 1, path.length)
                fetch('http://localhost:4800/addTask', {
                    method: 'POST',
                    body: JSON.stringify(data.data),
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
                            additem(data.data, filename, length)
                            addListener(data.data)
                            connect(data.data)
                            return
                        }
                        remote.dialog.showErrorBox('错误', data.msg)
                    })
                }).catch(function (error) {
                    remote.dialog.showErrorBox('错误', '请求失败:' + error.message)
                })
                return
            }
            remote.dialog.showErrorBox('错误', data.msg)
        })
    }).catch(function (error) {
        remote.dialog.showErrorBox('错误', '请求失败:' + error.message)
    })
}

function operate(id, event) {
    let data = {
        taskid: id,
        enum: event,
    }
    fetch('http://localhost:4800/oprate', {
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

