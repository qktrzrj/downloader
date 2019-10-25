let minimize = document.getElementById('file-minimize')
let close = document.getElementById('file-close')
let list = document.getElementById('list')
let cancel = document.getElementById('file-cancel')
let save = document.getElementById('file-save')
let fileList = []

let file = new WebSocket("ws://localhost:4800/fileList")
//连接打开时触发
file.onopen = function (evt) {
    console.log("Connection open ...")
}

file.onmessage = function (evt) {
    console.log("Received Message: " + evt.data)
    let data
    let err
    try {
        data = JSON.parse(evt.data)
    } catch (e) {
        console.log("json parse err:" + e)
        err = true
    }
    if (err === true) {
        return
    }
    fileList = data
    if (fileList !== null && fileList !== [] && fileList !== undefined) {
        addList()
    }
}

//连接关闭时触发
file.onclose = function (evt) {
    console.log("Connection close ...")
}


if (cancel) {
    cancel.addEventListener('click', () => {
        let boxes = document.getElementsByClassName('select')
        if (cancel.innerText === '取消全选') {
            for (let i = 0; i < boxes.length; i++) {
                boxes[i].checked = false
            }
            cancel.innerText = '全选'
            return
        }
        if (cancel.innerText === '全选') {
            for (let i = 0; i < boxes.length; i++) {
                boxes[i].checked = true
            }
            cancel.innerText = '取消全选'
        }
    })
}

if (save) {
    save.addEventListener('click', () => {
        let list = []
        let boxes = document.getElementsByClassName('select')
        for (let i = 0; i < boxes.length; i++) {
            if (boxes[i].checked === true) {
                list.push(fileList[parseInt(boxes[i].id)])
            }
        }
        let data = {
            op: 9,
            fileList: list
        }
        file.send(JSON.stringify(data))
    })
}

function addList() {
    for (let i = 0; i < fileList.length; i++) {
        let div = document.createElement('div')
        div.style.marginTop = '2px'
        list.appendChild(div)
        let box = document.createElement('input')
        box.type = 'checkbox'
        box.className = 'select'
        box.checked = true
        box.id = i
        let span = document.createElement('span')
        span.style.fontSize = '15px'
        span.innerText = fileList[i].filename
        div.appendChild(box)
        div.appendChild(span)
    }
}

function mouseListener(ac) {
    if (ac === 1) {
        close.setAttribute('src', './icon/close-rollover.png');
        minimize.setAttribute('src', './icon/minimize-rollover.png')
        // zoom.setAttribute('src', './icon/zoom-rollover.png')
        return
    }
    close.setAttribute('src', './icon/close.png');
    minimize.setAttribute('src', './icon/minimize.png')
    // zoom.setAttribute('src', './icon/zoom.png')
}

if (close) {
    close.addEventListener('mouseover', () => mouseListener(1))
    close.addEventListener('mouseleave', () => mouseListener(2))
    close.addEventListener('click', () => {
        close.setAttribute('src', './icon/close-pressed.png')
        let data = {
            op: 7,
        }
        file.send(JSON.stringify(data))
        close.setAttribute('src', './icon/close.png')
    })
}

if (minimize) {
    minimize.addEventListener('mouseover', () => mouseListener(1))
    minimize.addEventListener('mouseleave', () => mouseListener(2))
    minimize.addEventListener('click', () => {
        // ipcRenderer.send('window-minimize')
        minimize.setAttribute('src', './icon/minimize-pressed.png')
        let data = {
            op: 8,
        }
        file.send(JSON.stringify(data))
        minimize.setAttribute('src', './icon/minimize.png')
    })
}