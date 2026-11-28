<script setup>
import {computed, h, onBeforeMount, ref} from "vue";
import {useI18n} from "vue-i18n";
import {apiNormalApiTokenCreate, apiNormalApiTokenDelete, apiNormalApiTokenPage} from "../../../api/normal/api_token.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";
import Modal from "../../../components/Modal.vue";
import {requiredRule} from "../../../utils/formDataRule.js";
import {generateButtonGroup, generateDeleteButton} from "../../../utils/tableHelper.js";

const {t} = useI18n()

const fmtTs = (ts) => {
  if (!ts) return t('common.never')
  const d = new Date(ts * 1000)
  const p = (n) => (n < 10 ? '0' + n : n)
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

const state = ref({
  table: {searchLoading: false, search: {page: 1, size: 10}, list: [], total: 0, minWidth: 800},
  create: {open: false, loading: false, data: {name: '', expDays: 0}, dataRules: {name: requiredRule(t('apiToken.tokenName'))}},
  token: {open: false, value: ''},
})

const tableColumns = computed(() => [
  {title: t('apiToken.tokenName'), key: 'name', ellipsis: {tooltip: true}},
  {title: t('apiToken.tokenPrefix'), key: 'tokenPrefix', width: 160},
  {title: t('apiToken.expAt'), key: 'expAt', width: 170, render: (row) => fmtTs(row.expAt)},
  {title: t('apiToken.lastUsedAt'), key: 'lastUsedAt', width: 170, render: (row) => row.lastUsedAt ? fmtTs(row.lastUsedAt) : '-'},
  {title: t('common.createdAt'), key: 'createdAt', width: 170, render: (row) => fmtTs(row.createdAt)},
  {
    title: t('common.operate'), key: 'operator', width: 90, fixed: 'right', render(row) {
      return generateButtonGroup(generateDeleteButton(row, deleteFunc, t))
    }
  },
])

const refreshTable = () => pageFunc()
const pageFunc = async () => {
  try {
    state.value.table.searchLoading = true
    const res = await apiNormalApiTokenPage(state.value.table.search)
    state.value.table.list = res.data.list || []
    state.value.table.total = res.data.total || 0
  } finally { state.value.table.searchLoading = false }
}

const openCreate = () => { state.value.create.data = {name: '', expDays: 0}; state.value.create.open = true }
const closeCreate = () => { state.value.create.open = false }
const createRef = ref()
const createFunc = () => {
  createRef.value.validate(async valid => {
    if (!valid) {
      try {
        state.value.create.loading = true
        const res = await apiNormalApiTokenCreate(state.value.create.data)
        closeCreate()
        state.value.token = {open: true, value: res.data.token}
        pageFunc()
      } finally { state.value.create.loading = false }
    }
  })
}
const deleteFunc = async (row) => { await apiNormalApiTokenDelete({code: row.code}); pageFunc() }
const closeToken = () => { state.value.token.open = false }
const copyToken = async () => {
  try {
    await navigator.clipboard.writeText(state.value.token.value)
    $message.success(t('common.copied'))
  } catch (e) {
    $message.error(t('common.copy'))
  }
}

onBeforeMount(() => { pageFunc() })
</script>

<template>
  <div>
    <SearchCard :show-border="false" space>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" @click="refreshTable">{{ t('common.refresh') }}</n-button>
          <n-button type="info" :focusable="false" @click="openCreate">{{ t('common.create') }}</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>

    <AppCard>
      <n-data-table remote :columns="tableColumns" :data="state.table.list" :bordered="false"
                    :loading="state.table.searchLoading" :scroll-x="state.table.minWidth"/>
    </AppCard>
    <AppCard :show-border="false">
      <n-pagination :page-size="state.table.search.size" :page="state.table.search.page"
                    :item-count="state.table.total" :simple="true"
                    :on-update-page="(val) => {state.table.search.page = val;refreshTable()}"
                    :on-update-page-size="(val) => {state.table.search.size = val;refreshTable()}"/>
    </AppCard>

    <Modal :show="state.create.open" :title="t('apiToken.newToken')" width="400px" :confirm-text="t('common.save')"
           :cancel-text="t('common.cancel')" :confirm-loading="state.create.loading"
           @on-confirm="createFunc" @on-cancel="closeCreate" mask-close>
      <n-form ref="createRef" :rules="state.create.dataRules" :model="state.create.data">
        <n-form-item path="name" :label="t('apiToken.tokenName')">
          <n-input v-model:value.trim="state.create.data.name"/>
        </n-form-item>
        <n-form-item path="expDays" :label="t('apiToken.expDays')">
          <n-input-number v-model:value="state.create.data.expDays" :min="0"/>
        </n-form-item>
        <n-text depth="3">{{ t('apiToken.expDaysHint') }}</n-text>
      </n-form>
    </Modal>

    <Modal :show="state.token.open" :title="t('apiToken.newToken')" width="520px" :confirm-text="t('common.copy')"
           :cancel-text="t('common.close')" @on-confirm="copyToken" @on-cancel="closeToken">
      <n-alert type="warning" :show-icon="false" style="margin-bottom:12px">{{ t('apiToken.tokenOnce') }}</n-alert>
      <n-input :value="state.token.value" readonly type="textarea" :autosize="{minRows:2,maxRows:4}"/>
    </Modal>
  </div>
</template>

<style scoped lang="scss">
</style>
