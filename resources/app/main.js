// Modules to control application life and create native browser window
const {app, BrowserWindow, Tray, Menu} = require('electron')

// Keep a global reference of the window object, if you don't, the window will
// be closed automatically when the JavaScript object is garbage collected.
let mainWindow
let setwin
let download
let loading
let tray
let ipcMain = require('electron').ipcMain

const exec = require('child_process').exec

// 任何你期望执行的cmd命令，ls都可以
let cmdStr = './downloader'
// 执行cmd命令的目录，如果使用cd xx && 上面的命令，这种将会无法正常退出子进程
let cmdPath = '/Users/yan/GolandProjects/downloader/'
// 子进程名称
let workerProcess

function runExec() {
    // 执行命令行，如果命令不需要路径，或就是项目根目录，则不需要cwd参数：
    workerProcess = exec(cmdStr, {cwd: cmdPath})
    // 不受child_process默认的缓冲区大小的使用方法，没参数也要写上{}：workerProcess = exec(cmdStr, {})

    // 打印正常的后台可执行程序输出
    workerProcess.stdout.on('data', function (data) {
        console.log('stdout: ' + data);
    });

    // 打印错误的后台可执行程序输出
    workerProcess.stderr.on('data', function (data) {
        console.log('stderr: ' + data);
    });

    // 退出之后的输出
    workerProcess.on('close', function (code) {
        console.log('out code：' + code);
        app.quit()
    })
}

// 限制只可以打开一个应用, 4.x的文档
const gotTheLock = app.requestSingleInstanceLock()
if (!gotTheLock) {
    app.quit()
} else {
    app.on('second-instance', (event, commandLine, workingDirectory) => {
        // 当运行第二个实例时,将会聚焦到mainWindow这个窗口
        if (mainWindow) {
            if (mainWindow.isMinimized()) mainWindow.restore()
            mainWindow.focus()
            mainWindow.show()
        }
    })
}

function createWindow() {
    //runExec()
    // Create the browser window.
    mainWindow = new BrowserWindow({
        width: 800,
        height: 600,
        frame: false,
        icon: './icon/app.png',
        webPreferences: {
            nodeIntegration: true
        }
    })

    // and load the index.html of the app.
    mainWindow.loadFile('index.html')

    // Open the DevTools.
    mainWindow.webContents.openDevTools()

    // Emitted when the window is closed.
    mainWindow.on('closed', function () {
        // Dereference the window object, usually you would store windows
        // in an array if your app supports multi windows, this is the time
        // when you should delete the corresponding element.
        mainWindow = null
    })
    //let appIcon = require('electron').nativeImage.createFromPath('./icon/app.png')
    tray = new Tray('./icon/app.png')
    let contextMenu = Menu.buildFromTemplate([
        {
            label: '设置', click: () => {
                // 打开设置界面
                setting()
            }
        },
        {
            label: '退出', click: () => {
                app.quit()
            }
        },//我们需要在这里有一个真正的退出（这里直接强制退出）
    ])

    tray.setToolTip('下载器')
    tray.setContextMenu(contextMenu)
    tray.on('click', () => { //我们这里模拟桌面程序点击通知区图标实现打开关闭应用的功能
        mainWindow.isVisible() ? mainWindow.hide() : mainWindow.show()
        mainWindow.isVisible() ? mainWindow.setSkipTaskbar(false) : mainWindow.setSkipTaskbar(true);
    })
}

// This method will be called when Electron has finished
// initialization and is ready to create browser windows.
// Some APIs can only be used after this event occurs.
app.on('ready', createWindow)

// Quit when all windows are closed.
app.on('window-all-closed', function () {
    // On OS X it is common for applications and their menu bar
    // to stay active until the user quits explicitly with Cmd + Q
    if (process.platform !== 'darwin') {
        app.quit()
    }
})

app.on('activate', function () {
    // On OS X it's common to re-create a window in the app when the
    // dock icon is clicked and there are no other windows open.
    if (mainWindow === null) {
        createWindow()
    }
})

// In this file you can include the rest of your app's specific main process
// code. You can also put them in separate files and require them here.


ipcMain.on('window-close', function () {
    app.quit()
})

ipcMain.on('set-close', function () {
    setwin.destroy()
})

ipcMain.on('set-min', function () {
    setwin.minimize()
})

ipcMain.on('file-close', function () {
    download.destroy()
})

ipcMain.on('file-min', function () {
    download.minimize()
})

// 打开下载列表界面
ipcMain.on('download', function () {
    download = new BrowserWindow({
        parent: mainWindow, modal: true, show: false, frame: false, width: 300, height: 600,
        resizable: false,
        icon: './icon/app.png'
    })
    download.loadFile('./filelist.html')
    download.once('ready-to-show', () => {
        download.show()
    })
    download.webContents.openDevTools()
})

// 加载界面
ipcMain.on('loading', function () {
    loading = new BrowserWindow({
        parent: mainWindow, modal: true, show: false, frame: false, width: 400, height: 400,
        resizable: false, transparent: true
    })
    loading.loadFile('./loading.html')
    loading.once('ready-to-show', () => {
        loading.show()
    })
})

ipcMain.on('load-close', function () {
    loading.destroy()
})

function setting() {
    setwin = new BrowserWindow({
        parent: mainWindow, modal: true, show: false, frame: false, width: 550, height: 280,
        resizable: false,
        icon: './icon/app.png'
    })
    setwin.loadFile('./setting.html')
    setwin.once('ready-to-show', () => {
        setwin.show()
    })
    setwin.webContents.openDevTools()
}