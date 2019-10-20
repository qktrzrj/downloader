let ipcRenderer = require('electron').ipcRenderer
const { remote } = require('electron')


var zoom = document.getElementById('zoom');
var minimize = document.getElementById('minimize');
var close = document.getElementById('close');
var add = document.getElementById('add')
var input = document.getElementById('input')
var download = document.getElementById('download')
var ws = new WebSocket("ws://localhost:4800/getTaskInfo")

//连接打开时触发 
ws.onopen = function (evt) {
    console.log("Connection open ...");
};
//接收到消息时触发  
ws.onmessage = function (evt) {
    console.log("Received Message: " + evt.data);
};
//连接关闭时触发  
ws.onclose = function (evt) {
    console.log("Connection closed.");
};


function additem(id, name, size) {
    var itemdiv = document.createElement('div')
    itemdiv.className = 'item'
    itemdiv.id = id
    download.appendChild(itemdiv)
    var itemop = document.createElement('img')
    itemop.className = 'item-op'
    itemop.id = id + 'item-op'
    itemop.src = './icon/c_pauD.png'
    itemdiv.appendChild(itemop)
    var itemsp = document.createElement('div')
    itemsp.style = 'float:left;margin-left: auto;width: 1px;height: 100%;background: darkgray;'
    itemdiv.appendChild(itemsp)
    var iteminfo = document.createElement('div')
    iteminfo.className = 'item-info'
    iteminfo.id = id + 'item-info'
    itemdiv.appendChild(iteminfo)
    var filename = document.createElement('span')
    filename.className = 'filename'
    filename.id = id + 'filename'
    filename.innerText = name
    iteminfo.appendChild(filename)
    var br = document.createElement('br')
    filename.appendChild(br)
    var filenum = document.createElement('span')
    filenum.className = 'filenum'
    filenum.id = id + 'filenum'
    iteminfo.appendChild(filenum)
    var downsize = document.createElement('span')
    downsize.className = 'downsize'
    downsize.id = id + 'downsize'
    if (size >= 1024 * 1024) {
        downsize.innerText = '0 KB of ' + parseInt(size / (1024 * 1024)) + ' MB'
    } else {
        downsize.innerText = '0 KB of ' + parseInt(size / 1024) + ' KB'
    }
    iteminfo.appendChild(downsize)
    var filestatus = document.createElement('span')
    filestatus.className = 'filestatus'
    filestatus.id = id + 'filestatus'
    filestatus.innerText = 'waiting'
    itemdiv.appendChild(filestatus)
    var progress = document.createElement('progress')
    progress.className = 'progress'
    progress.id = id + 'progress'
    progress.value = 0
    progress.max = 100
    itemdiv.appendChild(progress)
    countuner()
}

function countuner() {
    var count = document.getElementsByClassName('filenum')
    var c = 1
    for (var i = 0; i < count.length; i++) {
        if (count[i].parentElement.parentElement.hidden) {
            continue
        }
        count[i].innerText = '[' + c + ']'
        c++
    }
}

function removeItem(id) {
    var item = document.getElementById(id)
    download.removeChild(item)
    countuner()
}

function hiddenItem(id) {
    var item = document.getElementById(id)
    item.hidden = true
}

function displayitem(id) {
    var item = document.getElementById(id)
    item.hidden = false
}

function mouseListener(ac) {
    if (ac == 1) {
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
        close.setAttribute('src', './icon/close-pressed.png');
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
        if (input.value == '') {
            remote.dialog.showErrorBox('错误', '无效的url')
            return
        }
        fetch('http://localhost:4800/getFileInfo?url=' + input.value).then(function (res) {
            input.value = ''
            if (res.status != 200) {
                remote.dialog.showErrorBox('错误', '服务异常')
                return
            }
            res.json().then((data) => {
                if (data != "" && data.code == 1) {
                    var filename = data.data.filename
                    var length = data.data.length
                    var savepath = data.data.savepath
                    const options = {
                        title: '保存文件',
                        defaultPath: savepath + filename,
                        filters: [
                            { name: 'All Files', extensions: ['*'] }
                        ]
                    }
                    var path = remote.dialog.showSaveDialogSync(options)
                    if (path == undefined) {
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
                        if (res.status != 200) {
                            remote.dialog.showErrorBox('错误', '服务异常')
                            return
                        }
                        res.json().then((data) => {
                            if (data != "" && data.code == 1) {
                                additem(data.data.id, filename, length)
                                ws.send(data.data.id)
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
    })
}
