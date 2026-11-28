<script setup>
import {computed, h, onBeforeMount, ref} from "vue";
import {useI18n} from "vue-i18n";
import {NTag} from "naive-ui";
import {apiNormalPayCreate, apiNormalPayQuery, apiNormalPayPage} from "../../../api/normal/pay.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";
import Modal from "../../../components/Modal.vue";
import {requiredRule} from "../../../utils/formDataRule.js";
import {generateButtonGroup, generateTextButton} from "../../../utils/tableHelper.js";
import {goToUrl} from "../../../utils/browser.js";

const {t} = useI18n()

const providerOptions = computed(() => [
  {key: 'sandbox', value: 'Sandbox'},
  {key: 'alipay', value: 'Alipay'},
  {key: 'wechat', value: 'WeChat'},
  {key: 'stripe', value: 'Stripe'},
])
const kindOptions = computed(() => [
  {key: 'recharge', value: t('pay.kindRecharge')},
  {key: 'group', value: t('pay.kindGroup')},
])
const statusOptions = computed(() => [
  {key: 0, value: t('pay.statusPending')},
  {key: 1, value: t('pay.statusPaid')},
  {key: 2, value: t('pay.statusClosed')},
  {key: 3, value: t('pay.statusFailed')},
])
const statusMap = () => ({
  0: {type: 'warning', text: t('pay.statusPending')},
  1: {type: 'success', text: t('pay.statusPaid')},
  2: {type: 'default', text: t('pay.statusClosed')},
  3: {type: 'error', text: t('pay.statusFailed')},
})
const kindText = (k) => k === 'group' ? t('pay.kindGroup') : t('pay.kindRecharge')

const state = ref({
  create: {
    open: false, loading: false,
    data: {amount: '', kind: 'recharge', targetCode: '', provider: 'sandbox', subject: ''},
    dataRules: {amount: requiredRule(t('pay.amount')), provider: requiredRule(t('pay.provider'))},
  },
  payResult: {open: false, code: '', payUrl: '', amount: ''},
  table: {searchLoading: false, search: {page: 1, size: 10, status: null}, list: [], total: 0, minWidth: 1000},
})

const tableColumns = computed(() => [
  {title: t('pay.subject'), key: 'Subject', ellipsis: {tooltip: true}},
  {title: t('pay.kind'), key: 'Kind', width: 100, render: (row) => kindText(row.Kind)},
  {title: t('pay.amount'), key: 'Amount', width: 100},
  {title: t('pay.provider'), key: 'Provider', width: 100},
  {
    title: t('common.status'), key: 'Status', width: 90, render(row) {
      const m = statusMap()[row.Status] || statusMap()[0]
      return h(NTag, {size: 'small', type: m.type}, {default: () => m.text})
    }
  },
  {title: t('pay.tradeNo'), key: 'TradeNo', ellipsis: {tooltip: true}, width: 160},
  {title: t('common.createdAt'), key: 'CreatedAt', width: 170, ellipsis: {tooltip: true}},
  {
    title: t('common.operate'), key: 'operator', width: 150, fixed: 'right', render(row) {
      return generateButtonGroup(
          generateTextButton(t('pay.queryOrder'), row, queryFunc, 'info'),
          generateTextButton(t('pay.gotoPay'), row, gotoPayFunc, 'success'),
      )
    }
  },
])

const refreshTable = () => pageFunc()
const searchTable = () => { state.value.table.search.page = 1; pageFunc() }
const pageFunc = async () => {
  try {
    state.value.table.searchLoading = true
    const res = await apiNormalPayPage(state.value.table.search)
    state.value.table.list = res.data.list || []
    state.value.table.total = res.data.total || 0
  } finally { state.value.table.searchLoading = false }
}

const openCreate = () => {
  state.value.create.data = {amount: '', kind: 'recharge', targetCode: '', provider: 'sandbox', subject: ''}
  state.value.create.open = true
}
const closeCreate = () => { state.value.create.open = false }
const createRef = ref()
const createFunc = () => {
  createRef.value.validate(async valid => {
    if (!valid) {
      try {
        state.value.create.loading = true
        const res = await apiNormalPayCreate(state.value.create.data)
        closeCreate()
        state.value.payResult = {open: true, code: res.data.code, payUrl: res.data.payUrl, amount: res.data.amount}
        searchTable()
      } finally { state.value.create.loading = false }
    }
  })
}
const closePayResult = () => { state.value.payResult.open = false }
const gotoPayFunc = (row) => {
  // 通过查单获取最新支付链接（待支付订单）
  apiNormalPayQuery({code: row.Code}).then(() => searchTable())
}
const gotoPayResult = () => {
  if (state.value.payResult.payUrl) goToUrl(state.value.payResult.payUrl)
}
const queryFunc = async (row) => {
  await apiNormalPayQuery({code: row.Code})
  searchTable()
}

onBeforeMount(() => { pageFunc() })
</script>

<template>
  <div>
    <SearchCard :show-border="false" space>
      <SearchItem type="select" :label-width="60" :label="t('common.status')" :options="statusOptions"
                  :options-key-value="{key:'key',value:'value'}" clearable
                  @onChange="value => state.table.search.status = value"/>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" @click="searchTable">{{ t('common.search') }}</n-button>
          <n-button type="info" :focusable="false" @click="refreshTable">{{ t('common.refresh') }}</n-button>
          <n-button type="success" :focusable="false" @click="openCreate">{{ t('pay.createOrder') }}</n-button>
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

    <Modal :show="state.create.open" :title="t('pay.createOrder')" width="440px" :confirm-text="t('pay.createOrder')"
           :cancel-text="t('common.cancel')" :confirm-loading="state.create.loading"
           @on-confirm="createFunc" @on-cancel="closeCreate" mask-close>
      <n-alert type="info" :show-icon="false" style="margin-bottom:12px">{{ t('pay.sandboxHint') }}</n-alert>
      <n-form ref="createRef" :rules="state.create.dataRules" :model="state.create.data">
        <n-form-item path="amount" :label="t('pay.amount')">
          <n-input v-model:value.trim="state.create.data.amount"/>
        </n-form-item>
        <n-form-item path="kind" :label="t('pay.kind')">
          <n-select v-model:value="state.create.data.kind" :options="kindOptions"
                    :label-field="'value'" :value-field="'key'"/>
        </n-form-item>
        <n-form-item v-if="state.create.data.kind==='group'" path="targetCode" :label="t('pay.targetCode')">
          <n-input v-model:value.trim="state.create.data.targetCode"/>
        </n-form-item>
        <n-form-item path="provider" :label="t('pay.provider')">
          <n-select v-model:value="state.create.data.provider" :options="providerOptions"
                    :label-field="'value'" :value-field="'key'"/>
        </n-form-item>
        <n-form-item path="subject" :label="t('pay.subject')">
          <n-input v-model:value.trim="state.create.data.subject"/>
        </n-form-item>
      </n-form>
    </Modal>

    <Modal :show="state.payResult.open" :title="t('pay.payUrl')" width="480px" :confirm-text="t('pay.gotoPay')"
           :cancel-text="t('common.close')" @on-confirm="gotoPayResult" @on-cancel="closePayResult" mask-close>
      <n-space vertical>
        <span>{{ t('pay.amount') }}: {{ state.payResult.amount }}</span>
        <n-input :value="state.payResult.payUrl" readonly type="textarea" :autosize="{minRows:2,maxRows:4}"/>
      </n-space>
    </Modal>
  </div>
</template>

<style scoped lang="scss">
</style>
