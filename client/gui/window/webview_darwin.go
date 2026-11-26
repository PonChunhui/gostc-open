//go:build darwin

package window

/*
#cgo darwin CFLAGS: -x objective-c -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa -framework WebKit
#include <stdlib.h>
#include "webview_darwin.h"
*/
import "C"

import (
	"unsafe"

	"gostc-sub/internal/common"
)

// macOS 实现说明：
// 托盘库 energye/systray 在 darwin 上通过 systray.Run 持有主线程的 [NSApp run] 事件循环，
// 这里不引入额外 webview 依赖，直接用 cgo 创建 NSWindow + WKWebView，
// 所有 AppKit 调用都派发到主队列，与托盘共享同一个事件循环。
// 关闭窗口仅隐藏到托盘（windowShouldClose 返回 NO），与 Windows 版行为一致。

// ShowMain 打开内嵌管理窗口；已打开则前置。可在任意 goroutine 调用。
func ShowMain() {
	curl := C.CString(mainURL())
	ctitle := C.CString("GOSTC GUI " + common.VERSION)
	defer C.free(unsafe.Pointer(curl))
	defer C.free(unsafe.Pointer(ctitle))
	C.gostcShowMain(curl, ctitle, C.int(1100), C.int(760))
}

// RefreshMain 让已打开的内嵌窗口跳转到最新监听地址。
func RefreshMain() {
	curl := C.CString(mainURL())
	defer C.free(unsafe.Pointer(curl))
	C.gostcRefreshMain(curl)
}

// CloseMain 隐藏内嵌管理窗口（可在任意 goroutine 调用）。
func CloseMain() {
	C.gostcCloseMain()
}
