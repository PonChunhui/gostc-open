<script setup>
import {computed, h, onBeforeMount, ref} from "vue";
import {useI18n} from "vue-i18n";
import {NTag} from "naive-ui";
import {
  apiNormalCertCreate, apiNormalCertUpdate, apiNormalCertDelete, apiNormalCertPage,
  apiNormalCertIssue, apiNormalCertRenew, apiNormalCertUpload
} from "../../../api/normal/cert.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";
import Modal from "../../../components/Modal.vue";
import {requiredRule} from "../../../utils/formDataRule.js";
import {generateButtonGroup, generateUpdateButton, generateDeleteButton, generateTextButton} from "../../../utils/tableHelper.js";
import {useTableBatch, runBatch} from "../../../utils/useTableBatch.js";

const {t} = useI18n()
const {checkedKeys, onCheck, clear} = useTableBatch()

const providerOptions = computed(() => [
  {key: 'dnspod', value: 'DNSPod'},
  {key: 'cloudflare', value: 'Cloudflare'},
  {key: 'alidns', value: 'AliDNS'},
])
const statusMap = () => ({
  0: {type: 'default', text: t('cert.statusPending')},
  1: {type: 'success', text: t('cert.statusValid')},
  2: {type: 'error', text: t('cert.statusFailed')},
  3: {type: 'warning', text: t('cert.statusExpired')},
})

const emptyCert = () => ({
  code: '', domain: '', provider: 'dnspod', autoRenew: 1,
  credential: {loginToken: '', authToken: '', zoneToken: '', authEmail: '', authKey: '', apiKey: '', secretKey: '', regionId: ''},
})

const state = ref({
  table: {searchLoading: false, search: {page: 1, size: 10, domain: '', provider: null}, list: [], total: 0, minWidth: 1100},
  create: {open: false, loading: false, data: emptyCert(), dataRules: {domain: requiredRule(t('cert.domain'))}},
  update: {open: false, loading: false, data: emptyCert(), dataRules: {domain: requiredRule(t('cert.domain'))}},
  upload: {open: false, loading: false, code: '', cert: '', key: ''},
})

const tableColumns = computed(() => [
  {type: 'selection', fixed: 'left', width: 50},
  {title: t('cert.domain'), key: 'domain', ellipsis: {tooltip: true}},
  {title: t('cert.provider'), key: 'provider', width: 110},
  {
    title: t('common.status'), key: 'status', width: 90, render(row) {
      const m = statusMap()[row.status] || statusMap()[0]
      return h(NTag, {size: 'small', type: m.type}, {default: () => m.text})
    }
  },
  {title: t('cert.expAt'), key: 'expAt', width: 170, ellipsis: {tooltip: true}},
  {
    title: t('cert.autoRenew'), key: 'autoRenew', width: 90, render(row) {
      return h(NTag, {size: 'small', type: row.autoRenew === 1 ? 'success' : 'default'}, {default: () => row.autoRenew === 1 ? t('common.yes') : t('common.no')})
    }
  },
  {title: t('health.lastError'), key: 'lastError', ellipsis: {tooltip: true}},
  {title: t('cert.pushSummary'), key: 'pushSummary', ellipsis: {tooltip: true}},
  {
    title: t('common.operate'), key: 'operator', width: 280, fixed: 'right', render(row) {
      return generateButtonGroup(
          generateUpdateButton(row, openUpdate, t),
          generateTextButton(t('cert.issue'), row, issueFunc, 'info'),
          generateTextButton(t('cert.renew'), row, renewFunc, 'warning'),
          generateTextButton(t('cert.upload'), row, openUpload, 'default'),
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
    const res = await apiNormalCertPage(state.value.table.search)
    state.value.table.list = res.data.list || []
    state.value.table.total = res.data.total || 0
  } finally { state.value.table.searchLoading = false }
}

const openCreate = () => { state.value.create.data = emptyCert(); state.value.create.open = true }
const closeCreate = () => { state.value.create.open = false }
const openUpdate = (row) => {
  const d = JSON.parse(JSON.stringify(row))
  d.credential = Object.assign(emptyCert().credential, d.credential || {})
  state.value.update.data = d
  state.value.update.open = true
}
const closeUpdate = () => { state.value.update.open = false }

const createRef = ref()
const createFunc = () => {
  createRef.value.validate(async valid => {
    if (!valid) {
      try {
        state.value.create.loading = true
        await apiNormalCertCreate(state.value.create.data)
        searchTable(); closeCreate()
      } finally { state.value.create.loading = false }
    }
  })
}
const updateRef = ref()
const updateFunc = () => {
  updateRef.value.validate(async valid => {
    if (!valid) {
      try {
        state.value.update.loading = true
        await apiNormalCertUpdate(state.value.update.data)
        searchTable(); closeUpdate()
      } finally { state.value.update.loading = false }
    }
  })
}
const deleteFunc = async (row) => { await apiNormalCertDelete({code: row.code}); searchTable() }
const batchDelete = () => {
  runBatch(t, checkedKeys.value, (code) => apiNormalCertDelete({code}), () => { clear(); searchTable() })
}

const issueFunc = (row) => {
  $dialog.warning({
    title: t('cert.issue'), content: t('cert.issueConfirm'),
    negativeText: t('common.cancel'), positiveText: t('common.confirm'),
    positiveButtonProps: {focusable: false},
    async onPositiveClick() { await apiNormalCertIssue({code: row.code}); searchTable() }
  })
}
const renewFunc = (row) => {
  $dialog.warning({
    title: t('cert.renew'), content: t('cert.renewConfirm'),
    negativeText: t('common.cancel'), positiveText: t('common.confirm'),
    positiveButtonProps: {focusable: false},
    async onPositiveClick() { await apiNormalCertRenew({code: row.code}); searchTable() }
  })
}
const openUpload = (row) => {
  state.value.upload = {open: true, loading: false, code: row.code, cert: '', key: ''}
}
const closeUpload = () => { state.value.upload.open = false }
const uploadFunc = async () => {
  try {
    state.value.upload.loading = true
    await apiNormalCertUpload({code: state.value.upload.code, cert: state.value.upload.cert, key: state.value.upload.key})
    searchTable(); closeUpload()
  } finally { state.value.upload.loading = false }
}

onBeforeMount(() => { pageFunc() })
</script>

<template>
  <div>
    <SearchCard :show-border="false" space>
      <SearchItem type="input" :label-width="60" :label="t('cert.domain')" clearable
                  @onChange="value => state.table.search.domain = value"/>
      <SearchItem type="select" :label-width="80" :label="t('cert.provider')" :options="providerOptions"
                  :options-key-value="{key:'key',value:'value'}" clearable
                  @onChange="value => state.table.search.provider = value"/>
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
                    :row-key="row => row.code" :checked-row-keys="checkedKeys"
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

    <Modal :show="state.create.open" :title="t('common.create')" width="480px" :confirm-text="t('common.save')"
           :cancel-text="t('common.cancel')" :confirm-loading="state.create.loading"
           @on-confirm="createFunc" @on-cancel="closeCreate" mask-close>
      <n-form ref="createRef" :rules="state.create.dataRules" :model="state.create.data">
        <n-form-item path="domain" :label="t('cert.domain')">
          <n-input v-model:value.trim="state.create.data.domain"/>
        </n-form-item>
        <n-form-item path="provider" :label="t('cert.provider')">
          <n-select v-model:value="state.create.data.provider" :options="providerOptions"
                    :label-field="'value'" :value-field="'key'"/>
        </n-form-item>
        <template v-if="state.create.data.provider==='dnspod'">
          <n-form-item :label="t('cert.loginToken')">
            <n-input v-model:value.trim="state.create.data.credential.loginToken"/>
          </n-form-item>
        </template>
        <template v-else-if="state.create.data.provider==='cloudflare'">
          <n-form-item :label="t('cert.authToken')">
            <n-input v-model:value.trim="state.create.data.credential.authToken"/>
          </n-form-item>
          <n-form-item :label="t('cert.zoneToken')">
            <n-input v-model:value.trim="state.create.data.credential.zoneToken"/>
          </n-form-item>
          <n-form-item :label="t('cert.authEmail')">
            <n-input v-model:value.trim="state.create.data.credential.authEmail"/>
          </n-form-item>
          <n-form-item :label="t('cert.authKey')">
            <n-input v-model:value.trim="state.create.data.credential.authKey"/>
          </n-form-item>
        </template>
        <template v-else-if="state.create.data.provider==='alidns'">
          <n-form-item :label="t('cert.apiKey')">
            <n-input v-model:value.trim="state.create.data.credential.apiKey"/>
          </n-form-item>
          <n-form-item :label="t('cert.secretKey')">
            <n-input v-model:value.trim="state.create.data.credential.secretKey"/>
          </n-form-item>
          <n-form-item :label="t('cert.regionId')">
            <n-input v-model:value.trim="state.create.data.credential.regionId"/>
          </n-form-item>
        </template>
        <n-form-item :label="t('cert.autoRenew')">
          <n-switch :value="state.create.data.autoRenew" :checked-value="1" :unchecked-value="2"
                    :on-update:value="value => state.create.data.autoRenew = value"/>
        </n-form-item>
      </n-form>
    </Modal>

    <Modal :show="state.update.open" :title="t('common.update')" width="480px" :confirm-text="t('common.save')"
           :cancel-text="t('common.cancel')" :confirm-loading="state.update.loading"
           @on-confirm="updateFunc" @on-cancel="closeUpdate" mask-close>
      <n-form ref="updateRef" :rules="state.update.dataRules" :model="state.update.data">
        <n-form-item path="domain" :label="t('cert.domain')">
          <n-input v-model:value.trim="state.update.data.domain"/>
        </n-form-item>
        <n-form-item path="provider" :label="t('cert.provider')">
          <n-select v-model:value="state.update.data.provider" :options="providerOptions"
                    :label-field="'value'" :value-field="'key'"/>
        </n-form-item>
        <template v-if="state.update.data.provider==='dnspod'">
          <n-form-item :label="t('cert.loginToken')">
            <n-input v-model:value.trim="state.update.data.credential.loginToken"/>
          </n-form-item>
        </template>
        <template v-else-if="state.update.data.provider==='cloudflare'">
          <n-form-item :label="t('cert.authToken')">
            <n-input v-model:value.trim="state.update.data.credential.authToken"/>
          </n-form-item>
          <n-form-item :label="t('cert.zoneToken')">
            <n-input v-model:value.trim="state.update.data.credential.zoneToken"/>
          </n-form-item>
          <n-form-item :label="t('cert.authEmail')">
            <n-input v-model:value.trim="state.update.data.credential.authEmail"/>
          </n-form-item>
          <n-form-item :label="t('cert.authKey')">
            <n-input v-model:value.trim="state.update.data.credential.authKey"/>
          </n-form-item>
        </template>
        <template v-else-if="state.update.data.provider==='alidns'">
          <n-form-item :label="t('cert.apiKey')">
            <n-input v-model:value.trim="state.update.data.credential.apiKey"/>
          </n-form-item>
          <n-form-item :label="t('cert.secretKey')">
            <n-input v-model:value.trim="state.update.data.credential.secretKey"/>
          </n-form-item>
          <n-form-item :label="t('cert.regionId')">
            <n-input v-model:value.trim="state.update.data.credential.regionId"/>
          </n-form-item>
        </template>
        <n-form-item :label="t('cert.autoRenew')">
          <n-switch :value="state.update.data.autoRenew" :checked-value="1" :unchecked-value="2"
                    :on-update:value="value => state.update.data.autoRenew = value"/>
        </n-form-item>
      </n-form>
    </Modal>

    <Modal :show="state.upload.open" :title="t('cert.upload')" width="560px" :confirm-text="t('common.save')"
           :cancel-text="t('common.cancel')" :confirm-loading="state.upload.loading"
           @on-confirm="uploadFunc" @on-cancel="closeUpload" mask-close>
      <n-form :model="state.upload">
        <n-form-item :label="t('cert.certContent')">
          <n-input type="textarea" :autosize="{minRows:4,maxRows:12}" v-model:value="state.upload.cert"/>
        </n-form-item>
        <n-form-item :label="t('cert.keyContent')">
          <n-input type="textarea" :autosize="{minRows:4,maxRows:12}" v-model:value="state.upload.key"/>
        </n-form-item>
      </n-form>
    </Modal>
  </div>
</template>

<style scoped lang="scss">
</style>
