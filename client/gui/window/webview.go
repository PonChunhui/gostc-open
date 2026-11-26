package window

import (
	"log"
	"net"
	"runtime"
	"sync"

	webview2 "github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
	"gostc-sub/gui/global"
	"gostc-sub/internal/common"
)

var (
	mainMu    sync.Mutex
	mainWv    webview2.WebView
	mainState int // 0=未创建 1=创建中 2=运行中

	user32dll               = windows.NewLazySystemDLL("user32.dll")
	procShowWindow          = user32dll.NewProc("ShowWindow")
	procSetForegroundWindow = user32dll.NewProc("SetForegroundWindow")
)

func mainURL() string {
	_, port, err := net.SplitHostPort(global.Config.Address)
	if err != nil || port == "" {
		port = "18080"
	}
	return "http://127.0.0.1:" + port
}

// ShowMain 打开内嵌管理窗口；已打开则恢复并前置。
func ShowMain() {
	mainMu.Lock()
	if mainState == 2 && mainWv != nil {
		hwnd := uintptr(mainWv.Window())
		mainMu.Unlock()
		procShowWindow.Call(hwnd, 9) // SW_RESTORE
		procSetForegroundWindow.Call(hwnd)
		return
	}
	if mainState == 1 {
		mainMu.Unlock()
		return
	}
	mainState = 1
	mainMu.Unlock()

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		// WebView2 要求创建线程初始化 STA COM
		if err := windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED); err != nil {
			log.Println("CoInitializeEx:", err)
		}
		wv := webview2.NewWithOptions(webview2.WebViewOptions{
			Debug:     false,
			AutoFocus: true,
			DataPath:  global.BasePath + "/webview2",
			WindowOptions: webview2.WindowOptions{
				Title:  "GOSTC GUI " + common.VERSION,
				Width:  1100,
				Height: 760,
				Center: true,
			},
		})
		if wv == nil {
			mainMu.Lock()
			mainState = 0
			mainMu.Unlock()
			log.Println("创建内嵌浏览器窗口失败：可能未安装 Microsoft Edge WebView2 运行时")
			text, _ := windows.UTF16PtrFromString("无法打开内嵌管理界面：未检测到 Microsoft Edge WebView2 运行时。\n请安装微软 WebView2 Runtime 后重试（Win10/Win11 一般已自带）。")
			caption, _ := windows.UTF16PtrFromString("GOSTC GUI")
			_, _ = windows.MessageBox(0, text, caption, windows.MB_OK|windows.MB_ICONERROR)
			return
		}
		mainMu.Lock()
		mainWv = wv
		mainState = 2
		mainMu.Unlock()

		wv.Navigate(mainURL())
		wv.Run()

		mainMu.Lock()
		mainWv = nil
		mainState = 0
		mainMu.Unlock()
	}()
}

// RefreshMain 让已打开的内嵌窗口跳转到最新监听地址。
func RefreshMain() {
	mainMu.Lock()
	wv := mainWv
	mainMu.Unlock()
	if wv != nil {
		url := mainURL()
		wv.Dispatch(func() {
			wv.Navigate(url)
		})
	}
}

// CloseMain 关闭内嵌管理窗口（可在任意 goroutine 调用）。
func CloseMain() {
	mainMu.Lock()
	wv := mainWv
	mainMu.Unlock()
	if wv != nil {
		wv.Destroy()
	}
}