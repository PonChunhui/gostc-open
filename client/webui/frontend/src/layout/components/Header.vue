<script setup>
import {NIcon, NBadge, NDrawer, NDrawerContent, NForm, NFormItem, NSelect, NInputNumber, NInput, NSwitch, NButton, NAlert, NSpace} from "naive-ui"
import {onBeforeMount, ref} from "vue";
import {localStore} from "../../store/local.js";
import {apiConfigQuery, apiConfigUpdateQuery, apiConfigUpdateSave, apiConfigUpdateCheck, apiConfigUpdateApply} from "../../api/config.js";

const state = ref({
  data: {
    version: 'v1.0.0'
  }
})

const configQueryFunc = async () => {
  try {
    let res = await apiConfigQuery()
    state.value.data = res.data
  } finally {
  }
}

// ===== 自更新 =====
const showUpdate = ref(false)
const checking = ref(false)
const applying = ref(false)
const saving = ref(false)
const pending = ref(null)
const updateSettings = ref({
  mode: 'notify',
  channel: 'stable',
  intervalHours: 24,
  panelUrl: '',
  autoRestart: true,
  publicKey: ''
})
const modeOptions = [
  {label: '关闭', value: 'off'},
  {label: '提示确认', value: 'notify'},
  {label: '静默自动', value: 'auto'}
]
const channelOptions = [
  {label: '稳定版 stable', value: 'stable'},
  {label: '测试版 beta', value: 'beta'}
]

const loadUpdate = async () => {
  try {
    let res = await apiConfigUpdateQuery()
    if (res.data) {
      updateSettings.value = Object.assign({}, updateSettings.value, res.data.settings || {})
      pending.value = res.data.pending || null
    }
  } catch (e) {
  }
}

const openUpdate = () => {
  showUpdate.value = true
  loadUpdate()
}

const saveSettings = async () => {
  saving.value = true
  try {
    await apiConfigUpdateSave(updateSettings.value)
    window.$message.create('设置已保存', {type: 'success', duration: 2000})
    await loadUpdate()
  } catch (e) {
  } finally {
    saving.value = false
  }
}

const checkUpdate = async () => {
  checking.value = true
  try {
    let res = await apiConfigUpdateCheck({panelUrl: updateSettings.value.panelUrl})
    pending.value = res.data || null
    if (pending.value) {
      window.$message.create('发现新版本 ' + pending.value.version, {type: 'success', duration: 3000})
    } else {
      window.$message.create('已是最新版本', {type: 'info', duration: 2000})
    }
  } catch (e) {
  } finally {
    checking.value = false
  }
}

const applyUpdate = async () => {
  applying.value = true
  try {
    await apiConfigUpdateApply({panelUrl: updateSettings.value.panelUrl})
    window.$message.create('升级完成，正在重启……', {type: 'success', duration: 4000})
    pending.value = null
  } catch (e) {
  } finally {
    applying.value = false
  }
}

onBeforeMount(() => {
  configQueryFunc()
  loadUpdate()
})

</script>

<template>
  <div class="header-container">
    <div class="header-title">
      GOSTC WEBUI
    </div>
    <div style="height: 100%;display: flex;align-items: end;padding: 14px 10px;box-sizing: border-box;color: #F1F1F1;">
      <span>{{ ` ${state.data.version} ` }}</span>
    </div>
    <div style="flex: 1"></div>
    <!--  更新设置  -->
    <NBadge :show="!!pending" dot :offset="[-2, 2]">
      <NIcon style="margin: 5px;user-select: none;cursor: pointer;color:#F1F1F1" :size="20" @click="openUpdate">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
          <g fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="3"></circle>
            <path
                d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33a1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82a1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
          </g>
        </svg>
      </NIcon>
    </NBadge>
    <!--  夜间模式  -->
    <NIcon style="margin: 5px;user-select: none;cursor: pointer" :size="20"
           @click="localStore().darkTheme = !localStore().darkTheme">
      <svg v-if="localStore().darkTheme" style="color: #F1F1F1" xmlns="http://www.w3.org/2000/svg"
           xmlns:xlink="http://www.w3.org/1999/xlink"
           viewBox="0 0 24 24">
        <g fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="4"></circle>
          <path d="M3 12h1m8-9v1m8 8h1m-9 8v1M5.6 5.6l.7.7m12.1-.7l-.7.7m0 11.4l.7.7m-12.1-.7l-.7.7"></path>
        </g>
      </svg>
      <svg v-else style="color: #F1F1F1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
           viewBox="0 0 24 24">
        <path
            d="M9.5 4c4.41 0 8 3.59 8 8s-3.59 8-8 8c-.34 0-.68-.02-1.01-.07c1.91-2.16 3.01-4.98 3.01-7.93s-1.1-5.77-3.01-7.93C8.82 4.02 9.16 4 9.5 4m0-2c-1.82 0-3.53.5-5 1.35c2.99 1.73 5 4.95 5 8.65s-2.01 6.92-5 8.65c1.47.85 3.18 1.35 5 1.35c5.52 0 10-4.48 10-10S15.02 2 9.5 2z"
            fill="currentColor"></path>
      </svg>
    </NIcon>

    <!--  更新设置抽屉  -->
    <NDrawer v-model:show="showUpdate" :width="360" placement="right">
      <NDrawerContent title="客户端更新" closable>
        <NAlert v-if="pending" type="success" :title="'发现新版本 ' + pending.version" style="margin-bottom: 12px;">
          <div v-if="pending.releaseNote" style="white-space: pre-wrap;">{{ pending.releaseNote }}</div>
          <div style="margin-top: 8px;">
            <NButton type="primary" size="small" :loading="applying" @click="applyUpdate">立即升级</NButton>
          </div>
        </NAlert>
        <NForm label-placement="left" label-width="90" size="small">
          <NFormItem label="更新模式">
            <NSelect v-model:value="updateSettings.mode" :options="modeOptions"/>
          </NFormItem>
          <NFormItem label="更新通道">
            <NSelect v-model:value="updateSettings.channel" :options="channelOptions"/>
          </NFormItem>
          <NFormItem label="检查间隔">
            <NInputNumber v-model:value="updateSettings.intervalHours" :min="1" :max="720">
              <template #suffix>小时</template>
            </NInputNumber>
          </NFormItem>
          <NFormItem label="面板地址">
            <NInput v-model:value="updateSettings.panelUrl" placeholder="留空使用连接地址"/>
          </NFormItem>
          <NFormItem label="升级后重启">
            <NSwitch v-model:value="updateSettings.autoRestart"/>
          </NFormItem>
          <NFormItem label="签名公钥">
            <NInput v-model:value="updateSettings.publicKey" type="textarea" :rows="2"
                    placeholder="ed25519 公钥(hex)，配置后强制校验签名"/>
          </NFormItem>
        </NForm>
        <NSpace>
          <NButton :loading="saving" @click="saveSettings">保存设置</NButton>
          <NButton :loading="checking" @click="checkUpdate">检查更新</NButton>
        </NSpace>
        <div style="margin-top: 12px;color: #999;font-size: 12px;">
          当前版本：{{ state.data.version }}。命令行可用 gostc update [--yes]。
        </div>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style lang="scss" scoped>
.header-container {
  padding: 0 10px;
  background-color: #303133;
  height: 60px;
  display: flex;
  align-items: center;

  .header-title {
    color: #F1F1F1;
    font-size: 1.8em;
    white-space: nowrap;
  }
}

@media screen and (max-width: 350px) {
  .header-title {
    display: none;
  }
}
</style>
