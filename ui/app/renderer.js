let ipcRenderer = require('electron').ipcRenderer
const { remote } = require('electron')


var zoom = document.getElementById('zoom');
var minimize = document.getElementById('minimize');
var close = document.getElementById('close');
var add = document.getElementById('add')
var input = document.getElementById('input')
var download = document.getElementById('download')

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
    downsize.innerText = '0 MB of ' + size + ' MB'
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

function countuner(){
    var count=document.getElementsByClassName('filenum')
    var c=1
    for(var i=0;i<count.length;i++){
        if(count[i].parentElement.parentElement.hidden){
            continue
        }
        count[i].innerText='['+c+']'
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
        additem(1, 'add', 120)
    })
}
