package main

import (
	_ "embed"
	"fmt"
	"gioui.org/io/system"
	service2 "github.com/SianHH/frp-package/package"
	"github.com/energye/systray"
	"gostc-sub/gui/global"
	"gostc-sub/gui/registry"
	"gostc-sub/gui/window"
	"gostc-sub/internal/common"
	"gostc-sub/webui/backend/bootstrap"
	_ "gostc-sub/webui/backend/router"
	_ "gostc-sub/webui/backend/todo"
	"log"
	"os"
)

//go:embed static/logo.ico
var logo []byte

var win *window.Window

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	moveWebuiCfgDir(global.BasePath, global.BasePath+"/data") // 移动旧配置文件内容
	global.BasePath = global.BasePath + "/data"
	if err := global.LoadConfig(global.BasePath + "/config.yaml"); err != nil {
		log.Fatalln(err)
	}

	bootstrap.InitLogger()
	bootstrap.InitFS(global.BasePath)
	bootstrap.InitTodo()
	bootstrap.InitTask()
	bootstrap.InitRouter()
	_ = bootstrap.StartServer(global.Config.Address)

	systray.SetIcon(logo)
	systray.SetTitle("GOSTC")
	systray.SetTooltip("GOSTC GUI " + common.VERSION)
	systray.SetOnClick(func(menu systray.IMenu) {
		window.ShowMain()
	})
	systray.SetOnRClick(func(menu systray.IMenu) {
		_ = menu.ShowMenu()
	})

	mRegistry := systray.AddMenuItem("开机自启", "修改注册表实现开机自启动")
	if registry.Registered() {
		mRegistry.Check()
	}
	mRegistry.Click(func() {
		if mRegistry.Checked() {
			registry.UnRegister()
			mRegistry.Uncheck()
		} else {
			if err := registry.Register(); err == nil {
				mRegistry.Check()
			} else {
				fmt.Println(err)
			}
		}
	})

	mMain := systray.AddMenuItem("打开管理界面", "打开内嵌管理窗口")
	mMain.Click(func() {
		window.ShowMain()
	})

	mSettings := systray.AddMenuItem("监听设置", "修改本地Web服务监听地址")
	mSettings.Click(func() {
		if WindowShow {
			win.Perform(system.ActionRaise)
			win.Perform(system.ActionUnmaximize)
			return
		} else {
			win = window.NewWindow("GOSTC GUI 监听设置 "+common.VERSION, window.OptionCloseCallback(func() {
				WindowShow = false
			}))
			_ = win.Create()
			WindowShow = true
		}
	})

	mQuit := systray.AddMenuItem("退出", "彻底退出程序")
	mQuit.Click(func() {
		systray.Quit()
	})

	window.ShowMain()
}

func onExit() {
	window.CloseMain()
	if WindowShow {
		win.Close()
	}
	bootstrap.Release()
	service2.Range(func(key string, value service2.Service) {
		value.Stop()
	})
	fmt.Println("onExit")
}

func moveWebuiCfgDir(basePath string, targetPath string) {
	_ = os.MkdirAll(targetPath, 0755)
	if cfgData, err := os.ReadFile(basePath + "/config.yaml"); err == nil {
		if err = os.WriteFile(targetPath+"/config.yaml", cfgData, 0644); err == nil {
			_ = os.Remove(basePath + "/config.yaml")
		}
	}
	if cfgData, err := os.ReadFile(basePath + "/client.json"); err == nil {
		if err = os.WriteFile(targetPath+"/client.json", cfgData, 0644); err == nil {
			_ = os.Remove(basePath + "/client.json")
		}
	}
	if cfgData, err := os.ReadFile(basePath + "/tunnel.json"); err == nil {
		if err = os.WriteFile(targetPath+"/tunnel.json", cfgData, 0644); err == nil {
			_ = os.Remove(basePath + "/tunnel.json")
		}
	}
	if cfgData, err := os.ReadFile(basePath + "/p2p.json"); err == nil {
		if err = os.WriteFile(targetPath+"/p2p.json", cfgData, 0644); err == nil {
			_ = os.Remove(basePath + "/p2p.json")
		}
	}
}
