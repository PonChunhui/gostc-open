//go:build darwin

package window

/*
#include <stdlib.h>
#include "webview_darwin.h"
*/
import "C"

import (
	"unsafe"

	"gostc-sub/gui/global"
)

// ShowSettings 弹出 macOS 原生监听设置对话框（NSAlert + 输入框）。
// 由托盘菜单回调（非主线程）触发，内部 dispatch_sync 到主线程执行。
func ShowSettings() {
	cur := C.CString(global.Config.Address)
	defer C.free(unsafe.Pointer(cur))
	res := C.gostcSettingsDialog(cur)
	if res == nil {
		return // 取消
	}
	defer C.free(unsafe.Pointer(res))
	addr := C.GoString(res)
	if addr == "" || addr == global.Config.Address {
		return
	}
	applyAddress(addr)
}

// CloseSettings macOS 设置对话框为模态，无需常驻关闭逻辑。
func CloseSettings() {}
