#ifndef GOSTC_WEBVIEW_DARWIN_H
#define GOSTC_WEBVIEW_DARWIN_H

// 打开/前置内嵌管理窗口（内部派发到主线程执行）
void gostcShowMain(const char *url, const char *title, int width, int height);
// 重新加载管理界面地址
void gostcRefreshMain(const char *url);
// 隐藏管理窗口
void gostcCloseMain(void);
// 弹出监听设置对话框；确认返回新地址（strdup，调用方 free），取消返回 NULL
char *gostcSettingsDialog(const char *current);

#endif
