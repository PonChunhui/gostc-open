<script setup>
import {computed, h, onBeforeMount, ref} from "vue";
import {useI18n} from "vue-i18n";
import {NTag} from "naive-ui";
import {
  apiAdminClientVersionCreate, apiAdminClientVersionUpdate, apiAdminClientVersionDelete, apiAdminClientVersionPage
} from "../../../api/admin/client_version.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";
import Modal from "../../../components/Modal.vue";
import {requiredRule} from "../../../utils/formDataRule.js";
import {generateButtonGroup, generateUpdateButton, generateDeleteButton} from "../../../utils/tableHelper.js";
import {useTableBatch, runBatch} from "../../../utils/useTableBatch.js";

const {t} = useI18n()
const {checkedKeys, onCheck, clear} = useTableBatch()

const platformOptions = [
  {key: 'windows', value: 'Windows'},
  {key: 'linux', value: 'Linux'},
  {key: 'darwin', value: 'macOS'},
]
const archOptions = [
  {key: 'amd64', value: 'amd64'},
  {key: 'arm64', value: 'arm64'},
]
const channelOptions = computed(() => [
  {key: 'stable', value: t('clientVersion.channelStable')},
  {key: 'beta', value: t('clientVersion.channelBeta')},
])
const emptyData = () => ({
  code: '', version: '', platform: 'windows', arch: 'amd64', downloadUrl: '', sha256: '',
  signature: '', channel: 'stable', grayPercent: 100, enable: 1, releaseNote: '',
})

const state = ref({
  table: {searchLoading: false, search: {page: 1, size: 10, platform: null, channel: null}, list: [], total: 0, minWidth: 1100},
  create: {open: false, loading: false, data: emptyData(), dataRules: {version: requiredRule(t('clientVersion.version')), downloadUrl: requiredRule(t('clientVersion.downloadUrl'))}},
  update: {open: false, loading: false, data: emptyData(), dataRules: {version: requiredRule(t('clientVersion.version')), downloadUrl: requiredRule(t('clientVersion.downloadUrl'))}},
})

const tableColumns = computed(() => [
  {type: 'selection', fixed: 'left', width: 50},
  {title: t('clientVersion.version'), key: 'Version', width: 110, ellipsis: {tooltip: true}},
  {title: t('clientVersion.platform'), key: 'Platform', width: 90},
  {title: t('clientVersion.arch'), key: 'Arch', width: 80},
  {
    title: t('clientVersion.channel'), key: 'Channel', width: 90, render(row) {
      return h(NTag, {size: 'small', type: row.Channel === 'stable' ? 'success' : 'warning'},
          {default: () => row.Channel === 'stable' ? t('clientVersion.channelStable') : t('clientVersion.channelBeta')})
    }
  },
  {title: t('clientVersion.grayPercent'), key: 'GrayPercent', width: 90, render: (row) => row.GrayPercent + '%'},
  {
    title: t('common.status'), key: 'Enable', width: 90, render(row) {
      return h(NTag, {size: 'small', type: row.Enable === 1 ? 'success' : 'default'},
          {default: () => row.Enable === 1 ? t('common.enabled') : t('common.disabled')})
    }
  },
  {title: t('clientVersion.downloadUrl'), key: 'DownloadURL', ellipsis: {tooltip: true}},
  {title: t('clientVersion.releaseNote'), key: 'ReleaseNote', ellipsis: {tooltip: true}},
  {title: t('common.createdAt'), key: 'CreatedAt', width: 170, ellipsis: {tooltip: true}},
  {
    title: t('common.operate'), key: 'operator', width: 130, fixed: 'right', render(row) {
      return generateButtonGroup(
          generateUpdateButton(row, openUpdate, t),
          generateDeleteButton(row, deleteFunc, t),
      )
    }
  },
])

const refreshTable = () => pageFunc()
const searchTable = () => { state.value.table.search.page = 1; pageFunc() }
const pageFunc = async () => {
  try {
    state.value.table.searchLoading = true
    const res = await apiAdminClientVersionPage(state.value.table.search)
    state.value.table.list = res.data.list || []
    state.value.table.total = res.data.total || 0
  } finally { state.value.table.searchLoading = false }
}

const openCreate = () => { state.value.create.data = emptyData(); state.value.create.open = true }
const closeCreate = () => { state.value.create.open = false }
const createRef = ref()
const createFunc = () => {
  createRef.value.validate(async valid => {
    if (!valid) {
      try {
        state.value.create.loading = true
        await apiAdminClientVersionCreate(state.value.create.data)
        searchTable(); closeCreate()
      } finally { state.value.create.loading = false }
    }
  })
}
const openUpdate = (row) => {
  state.value.update.data = {
    code: row.Code, version: row.Version, platform: row.Platform, arch: row.Arch, downloadUrl: row.DownloadURL,
    sha256: row.Sha256, signature: row.Signature, channel: row.Channel, grayPercent: row.GrayPercent,
    enable: row.Enable, releaseNote: row.ReleaseNote,
  }
  state.value.update.open = true
}
const closeUpdate = () => { state.value.update.open = false }
const updateRef = ref()
const updateFunc = () => {
  updateRef.value.validate(async valid => {
    if (!valid) {
      try {
        state.value.update.loading = true
        await apiAdminClientVersionUpdate(state.value.update.data)
        refreshTable(); closeUpdate()
      } finally { state.value.update.loading = false }
    }
  })
}
const deleteFunc = async (row) => { await apiAdminClientVersionDelete({code: row.Code}); searchTable() }
const batchDelete = () => {
  runBatch(t, checkedKeys.value, (code) => apiAdminClientVersionDelete({code}), () => { clear(); searchTable() })
}

onBeforeMount(() => { pageFunc() })
</script>

<template>
  <div>
    <SearchCard :show-border="false" space>
      <SearchItem type="select" :label-width="80" :label="t('clientVersion.platform')" :options="platformOptions"
                  :options-key-value="{key:'key',value:'value'}" clearable
                  @onChange="value => state.table.search.platform = value"/>
      <SearchItem type="select" :label-width="80" :label="t('clientVersion.channel')" :options="channelOptions"
                  :options-key-value="{key:'key',value:'value'}" clearable
                  @onChange="value => state.table.search.channel = value"/>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" @click="searchTable">{{ t('common.search') }}</n-button>
          <n-button type="info" :focusable="false" @click="refreshTable">{{ t('common.refresh') }}</n-button>
          <n-button type="info" :focusable="false" @click="openCreate">{{ t('common.create') }}</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>

    <AppCard :show-border="false" v-if="checkedKeys.length">
      <n-space align="center">
        <span>{{ t('common.selected', {n: checkedKeys.length}) }}</span>
        <n-button size="small" type="error" :focusable="false" @click="batchDelete">{{ t('common.batchDelete') }}</n-button>
        <n-button size="small" :focusable="false" @click="clear">{{ t('common.clearSelection') }}</n-button>
      </n-space>
    </AppCard>
    <AppCard>
      <n-data-table remote :columns="tableColumns" :data="state.table.list" :bordered="false"
                    :loading="state.table.searchLoading" :scroll-x="state.table.minWidth"
                    :row-key="row => row.Code" :checked-row-keys="checkedKeys"
                    @update:checked-row-keys="onCheck">
        <template #empty>
          <n-empty :description="t('common.noData')"/>
        </template>
      </n-data-table>
    </AppCard>
    <AppCard :show-border="false">
      <n-pagination :page-size="state.table.search.size" :page="state.table.search.page"
                    :item-count="state.table.total" :simple="true"
                    :on-update-page="(val) => {state.table.search.page = val;refreshTable()}"
                    :on-update-page-size="(val) => {state.table.search.size = val;refreshTable()}"/>
    </AppCard>

    <Modal :show="state.create.open" :title="t('common.create')" width="520px" :confirm-text="t('common.save')"
           :cancel-text="t('common.cancel')" :confirm-loading="state.create.loading"
           @on-confirm="createFunc" @on-cancel="closeCreate" mask-close>
      <n-form ref="createRef" :rules="state.create.dataRules" :model="state.create.data">
        <n-form-item path="version" :label="t('clientVersion.version')"><n-input v-model:value.trim="state.create.data.version"/></n-form-item>
        <n-form-item path="platform" :label="t('clientVersion.platform')"><n-select v-model:value="state.create.data.platform" :options="platformOptions" :label-field="'value'" :value-field="'key'"/></n-form-item>
        <n-form-item path="arch" :label="t('clientVersion.arch')"><n-select v-model:value="state.create.data.arch" :options="archOptions" :label-field="'value'" :value-field="'key'"/></n-form-item>
        <n-form-item path="downloadUrl" :label="t('clientVersion.downloadUrl')"><n-input v-model:value.trim="state.create.data.downloadUrl"/></n-form-item>
        <n-form-item path="sha256" :label="t('clientVersion.sha256')"><n-input v-model:value.trim="state.create.data.sha256"/></n-form-item>
        <n-form-item path="signature" :label="t('clientVersion.signature')"><n-input v-model:value.trim="state.create.data.signature"/></n-form-item>
        <n-form-item path="channel" :label="t('clientVersion.channel')"><n-select v-model:value="state.create.data.channel" :options="channelOptions" :label-field="'value'" :value-field="'key'"/></n-form-item>
        <n-form-item path="grayPercent" :label="t('clientVersion.grayPercent')"><n-input-number v-model:value="state.create.data.grayPercent" :min="0" :max="100"/></n-form-item>
        <n-form-item path="enable" :label="t('common.enable')"><n-switch v-model:value="state.create.data.enable" :checked-value="1" :unchecked-value="2"/></n-form-item>
        <n-form-item path="releaseNote" :label="t('clientVersion.releaseNote')"><n-input type="textarea" :autosize="{minRows:2,maxRows:6}" v-model:value="state.create.data.releaseNote"/></n-form-item>
      </n-form>
    </Modal>

    <Modal :show="state.update.open" :title="t('common.update')" width="520px" :confirm-text="t('common.save')"
           :cancel-text="t('common.cancel')" :confirm-loading="state.update.loading"
           @on-confirm="updateFunc" @on-cancel="closeUpdate" mask-close>
      <n-form ref="updateRef" :rules="state.update.dataRules" :model="state.update.data">
        <n-form-item path="version" :label="t('clientVersion.version')"><n-input v-model:value.trim="state.update.data.version"/></n-form-item>
        <n-form-item path="platform" :label="t('clientVersion.platform')"><n-select v-model:value="state.update.data.platform" :options="platformOptions" :label-field="'value'" :value-field="'key'"/></n-form-item>
        <n-form-item path="arch" :label="t('clientVersion.arch')"><n-select v-model:value="state.update.data.arch" :options="archOptions" :label-field="'value'" :value-field="'key'"/></n-form-item>
        <n-form-item path="downloadUrl" :label="t('clientVersion.downloadUrl')"><n-input v-model:value.trim="state.update.data.downloadUrl"/></n-form-item>
        <n-form-item path="sha256" :label="t('clientVersion.sha256')"><n-input v-model:value.trim="state.update.data.sha256"/></n-form-item>
        <n-form-item path="signature" :label="t('clientVersion.signature')"><n-input v-model:value.trim="state.update.data.signature"/></n-form-item>
        <n-form-item path="channel" :label="t('clientVersion.channel')"><n-select v-model:value="state.update.data.channel" :options="channelOptions" :label-field="'value'" :value-field="'key'"/></n-form-item>
        <n-form-item path="grayPercent" :label="t('clientVersion.grayPercent')"><n-input-number v-model:value="state.update.data.grayPercent" :min="0" :max="100"/></n-form-item>
        <n-form-item path="enable" :label="t('common.enable')"><n-switch v-model:value="state.update.data.enable" :checked-value="1" :unchecked-value="2"/></n-form-item>
        <n-form-item path="releaseNote" :label="t('clientVersion.releaseNote')"><n-input type="textarea" :autosize="{minRows:2,maxRows:6}" v-model:value="state.update.data.releaseNote"/></n-form-item>
      </n-form>
    </Modal>
  </div>
</template>

<style scoped lang="scss">
</style>
