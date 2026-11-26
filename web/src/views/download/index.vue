<script setup>
import {onBeforeMount, ref} from "vue";
import {useThemeVars} from "naive-ui";
import {apiPublicDownloadList} from "../../api/public/index.js";
import PublicHeader from "../../components/PublicHeader.vue";

const themeVars = useThemeVars()
const state = ref({
  loading: false,
  list: [],
})

const loadFunc = async () => {
  state.value.loading = true
  try {
    const res = await apiPublicDownloadList()
    state.value.list = res.data || []
  } catch (e) {
    state.value.list = []
  } finally {
    state.value.loading = false
  }
}

const sizeText = (size) => {
  if (size >= 1024 * 1024) {
    return (size / 1024 / 1024).toFixed(1) + ' MB'
  }
  return (size / 1024).toFixed(0) + ' KB'
}

const linuxArch = (lower) => {
  if (lower.includes('arm64')) return 'ARM64'
  if (lower.includes('armv7')) return 'ARMv7'
  if (lower.includes('amd64')) return 'x86_64'
  if (lower.includes('386')) return 'x86 32位'
  if (lower.includes('s390x')) return 's390x'
  return ''
}

const fileDesc = (name) => {
  const lower = name.toLowerCase()
  if (lower.includes('gui')) {
    return 'Windows 图形版 · 双击运行（需管理员权限），推荐新手使用'
  }
  if (lower.includes('linux')) {
    const arch = linuxArch(lower)
    return 'Linux 命令行版' + (arch ? '（' + arch + '）' : '') + ' · 支持注册 systemd 服务'
  }
  if (lower.includes('cli')) {
    return 'Windows 命令行版 · 轻量灵活，适合脚本与服务器'
  }
  return '客户端安装包'
}

const fileUrl = (name) => {
  return '/api/v1/public/download/file?name=' + encodeURIComponent(name)
}

onBeforeMount(() => {
  loadFunc()
})
</script>

<template>
  <div class="download">
    <PublicHeader/>
    <div class="container">
      <h1 class="page-title">客户端下载</h1>
      <p class="page-desc">在内网机器上运行客户端，填入服务端地址与连接密钥即可开始穿透。连接密钥在控制台「客户端」页面获取。</p>

      <n-spin :show="state.loading">
        <n-empty v-if="!state.loading && state.list.length === 0" description="管理员尚未上传客户端安装包" style="margin: 60px 0"/>
        <n-grid v-else cols="1 s:2 m:3" responsive="screen" :x-gap="16" :y-gap="16">
          <n-gi v-for="item in state.list" :key="item.name">
            <n-card class="file-card" size="small">
              <div class="file-name">{{ item.name }}</div>
              <div class="file-desc">{{ fileDesc(item.name) }}</div>
              <div class="file-meta">{{ sizeText(item.size) }} · 更新于 {{ item.updatedAt }}</div>
              <n-button tag="a" :href="fileUrl(item.name)" :download="item.name" type="primary" size="small" style="margin-top: 12px">下载</n-button>
            </n-card>
          </n-gi>
        </n-grid>
      </n-spin>

      <n-card class="tips" size="small" title="使用说明" style="margin-top: 24px">
        <div class="tip-line">Windows 图形版：双击运行，在界面中添加客户端（服务端地址 + 连接密钥），启动后可设置开机自启。</div>
        <div class="tip-line">命令行版：</div>
        <n-code code="gostc -addr 服务端地址 -key 连接密钥" language="bash"/>
        <div class="tip-line">服务端未启用 SSL 时追加 <n-text code>--tls=false</n-text>；Linux 注册服务：<n-text code>gostc install -addr 服务端地址 -key 连接密钥</n-text>。</div>
        <div class="tip-line">更多安装方式与排障请查看 <a href="/docs/" target="_blank" rel="noopener">帮助文档</a>。</div>
      </n-card>

      <n-card class="tips" size="small" title="如何查看本机架构" style="margin-top: 16px">
        <div class="tip-line">Linux：运行下方命令，按输出选择对应的 Linux 包。</div>
        <n-code code="uname -m" language="bash"/>
        <div class="tip-line arch-map">
          <span><n-text code>x86_64</n-text> → linux-amd64</span>
          <span><n-text code>aarch64</n-text> → linux-arm64</span>
          <span><n-text code>armv7l</n-text> → linux-armv7</span>
          <span><n-text code>i686 / i386</n-text> → linux-386</span>
          <span><n-text code>s390x</n-text> → linux-s390x</span>
        </div>
        <div class="tip-line">Windows：在「设置 → 系统 → 关于」查看“系统类型”，64 位选 amd64 包、32 位选 386 包；也可在命令行运行：</div>
        <n-code code="echo %PROCESSOR_ARCHITECTURE%" language="bash"/>
        <div class="tip-line">输出 <n-text code>AMD64</n-text> 为 64 位，<n-text code>x86</n-text> 为 32 位。</div>
      </n-card>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.download {
  min-height: 100vh;
  color: v-bind('themeVars.textColor2');
}

.container {
  max-width: 1100px;
  margin: 0 auto;
  padding: 40px 20px 60px;
}

.page-title {
  font-size: 1.8em;
  margin: 0 0 8px;
  color: v-bind('themeVars.textColor1');
}

.page-desc {
  margin: 0 0 24px;
  color: v-bind('themeVars.textColor3');
}

.file-card {
  height: 100%;

  .file-name {
    font-weight: 600;
    font-size: 1.05em;
    word-break: break-all;
    color: v-bind('themeVars.textColor1');
  }

  .file-desc {
    margin-top: 6px;
    font-size: .9em;
    color: v-bind('themeVars.textColor3');
  }

  .file-meta {
    margin-top: 8px;
    font-size: .85em;
    color: v-bind('themeVars.textColor3');
  }
}

.tips {
  .tip-line {
    margin: 8px 0;
    font-size: .92em;
  }

  .arch-map {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 18px;
  }

  a {
    color: v-bind('themeVars.primaryColor');
    text-decoration: none;
  }
}
</style>