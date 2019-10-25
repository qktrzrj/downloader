let savepath
let threadnum
let path = document.getElementById('path')
let thread = document.getElementById('thread')
// let zoom = document.getElementById('set-zoom')
let minimize = document.getElementById('set-minimize')
let close = document.getElementById('set-close')
let cancel = document.getElementById('set-cancel')
let save = document.getElementById('set-save')
let divObj = document.getElementById("promptDiv");

let set = new WebSocket("ws://localhost:4800/getSetting")
//连接打开时触发
set.onopen = function (evt) {
    console.log("Connection open ...")
}

set.onmessage = function (evt) {
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
    if (data.op === 6) {
        path.innerText = data.savePath
        return
    }
    savepath = path.innerText = data.savePath
    threadnum = thread.value = data.maxRoutineNum
}

//连接关闭时触发
set.onclose = function (evt) {
    console.log("Connection close ...")
}

//传入 event 对象
function ShowPrompt(objEvent) {
    divObj.style.visibility = "visible";
}

//隐藏提示框

function HiddenPrompt() {
    divObj = document.getElementById("promptDiv");
    divObj.innerText = thread.value
    divObj.style.visibility = "hidden";
}

if (divObj) {
    divObj.addEventListener('mousemove', (event) => {
        divObj.innerText = thread.value
        //使用这一行代码，提示层将出现在鼠标附近(如要使用，记得注释 divOjb.style.left = 60+5; 这一句)
        divObj.style.marginLeft = event.clientX + 5;   //clientX 为鼠标在窗体中的 x 坐标值
        // divObj.style.left = 60 + 5;
        divObj.style.marginTop = event.clientY + 5;     //clientY 为鼠标在窗体中的 y 坐标值
    })
}

if (thread) {
    thread.addEventListener('mouseover', (event) => {
        ShowPrompt(event)
    })

    thread.addEventListener('mouseleave', () => {
        HiddenPrompt()
    })

    thread.onchange = function () {
        divObj.innerText = thread.value
    }
}

if (path) {
    path.addEventListener('click', () => {
        let data = {
            op: 6,
            savePath: path.innerText,
            maxRoutineNum: parseInt(threadnum)
        }
        set.send(JSON.stringify(data))
    })
}

if (save) {
    save.addEventListener('click', () => {
        savepath = path.innerText
        threadnum = thread.value
        let data = {
            op: 5,
            savePath: savepath,
            maxRoutineNum: parseInt(threadnum)
        }
        set.send(JSON.stringify(data))
    })
}

if (cancel) {
    cancel.addEventListener('click', () => {
        path.innerText = savepath
        thread.value = threadnum
        let data = {
            op: 4,
            savePath: savepath,
            maxRoutineNum: parseInt(threadnum)
        }
        set.send(JSON.stringify(data))
    })
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
            op: 1,
            savePath: savepath,
            maxRoutineNum: parseInt(threadnum)
        }
        set.send(JSON.stringify(data))
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
            op: 2,
            savePath: savepath,
            maxRoutineNum: parseInt(threadnum)
        }
        set.send(JSON.stringify(data))
        minimize.setAttribute('src', './icon/minimize.png')
    })
}

// if (zoom) {
//     zoom.addEventListener('mouseover', () => mouseListener(1))
//     zoom.addEventListener('mouseleave', () => mouseListener(2))
//     zoom.addEventListener('click', () => {
//         zoom.setAttribute('src', './icon/zoom-pressed.png')
//         let data = {
//             op: 3,
//             savePath: savepath,
//             maxRoutineNum: parseInt(threadnum)
//         }
//         set.send(JSON.stringify(data))
//         zoom.setAttribute('src', './icon/zoom.png')
//     })
// }