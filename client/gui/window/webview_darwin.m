#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>
#include <stdlib.h>
#include <string.h>
#include "webview_darwin.h"

@interface GostcWindowDelegate : NSObject <NSWindowDelegate>
@end

@implementation GostcWindowDelegate
// 关闭窗口时仅隐藏到托盘，与 Windows 版行为一致；
// 同时退回 Accessory 模式，隐藏 Dock 图标，保持纯托盘应用形态
- (BOOL)windowShouldClose:(NSWindow *)sender {
    [sender orderOut:nil];
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    return NO;
}
@end

static NSWindow *gWindow = nil;
static WKWebView *gWebView = nil;
static GostcWindowDelegate *gDelegate = nil;

static void runOnMain(dispatch_block_t block) {
    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

static void loadURL(NSString *url) {
    NSURL *u = [NSURL URLWithString:url];
    if (u != nil && gWebView != nil) {
        [gWebView loadRequest:[NSURLRequest requestWithURL:u]];
    }
}

void gostcShowMain(const char *url, const char *title, int width, int height) {
    NSString *nsURL = [NSString stringWithUTF8String:url];
    NSString *nsTitle = [NSString stringWithUTF8String:title];
    runOnMain(^{
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
        if (gWindow == nil) {
            NSRect frame = NSMakeRect(0, 0, width, height);
            NSUInteger style = NSWindowStyleMaskTitled | NSWindowStyleMaskClosable |
                               NSWindowStyleMaskMiniaturizable | NSWindowStyleMaskResizable;
            gWindow = [[NSWindow alloc] initWithContentRect:frame
                                                  styleMask:style
                                                    backing:NSBackingStoreBuffered
                                                      defer:NO];
            [gWindow setReleasedWhenClosed:NO];
            [gWindow setMinSize:NSMakeSize(640, 480)];
            gDelegate = [[GostcWindowDelegate alloc] init];
            [gWindow setDelegate:gDelegate];
            // 默认 WKWebsiteDataStore 为持久化存储，localStorage 等跨启动保留
            WKWebViewConfiguration *cfg = [[WKWebViewConfiguration alloc] init];
            gWebView = [[WKWebView alloc] initWithFrame:frame configuration:cfg];
            [gWindow setContentView:gWebView];
            [gWindow center];
        }
        [gWindow setTitle:nsTitle];
        loadURL(nsURL);
        [gWindow makeKeyAndOrderFront:nil];
        [NSApp activateIgnoringOtherApps:YES];
    });
}

void gostcRefreshMain(const char *url) {
    NSString *nsURL = [NSString stringWithUTF8String:url];
    runOnMain(^{
        loadURL(nsURL);
    });
}

void gostcCloseMain(void) {
    runOnMain(^{
        if (gWindow != nil) {
            [gWindow orderOut:nil];
        }
    });
}

char *gostcSettingsDialog(const char *current) {
    NSString *nsCurrent = [NSString stringWithUTF8String:current];
    __block NSString *result = nil;
    dispatch_block_t block = ^{
        NSAlert *alert = [[NSAlert alloc] init];
        [alert setMessageText:@"监听设置"];
        [alert setInformativeText:@"修改本地 Web 服务监听地址，例如 0.0.0.0:18080"];
        [alert addButtonWithTitle:@"保存配置并重启"];
        [alert addButtonWithTitle:@"取消"];
        NSTextField *input = [[NSTextField alloc] initWithFrame:NSMakeRect(0, 0, 320, 24)];
        [input setStringValue:nsCurrent];
        [alert setAccessoryView:input];
        [alert.window setInitialFirstResponder:input];
        NSModalResponse resp = [alert runModal];
        if (resp == NSAlertFirstButtonReturn) {
            result = [input stringValue];
        }
    };
    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_sync(dispatch_get_main_queue(), block);
    }
    if (result == nil) {
        return NULL;
    }
    return strdup([result UTF8String]);
}
