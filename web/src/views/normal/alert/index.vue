<script setup>
import {computed, h, onBeforeMount, ref} from "vue";
import {useI18n} from "vue-i18n";
import {NTag} from "naive-ui";
import {
  apiNormalAlertRuleCreate, apiNormalAlertRuleUpdate, apiNormalAlertRuleDelete,
  apiNormalAlertRulePage, apiNormalAlertLogPage
} from "../../../api/normal/alert.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";
import Modal from "../../../components/Modal.vue";
import {requiredRule} from "../../../utils/formDataRule.js";
import {generateButtonGroup, generateUpdateButton, generateDeleteButton} from "../../../utils/tableHelper.js";
import {useTableBatch, runBatch} from "../../../utils/useTableBatch.js";

const {t} = useI18n()
const {checkedKeys, onCheck, clear} = useTableBatch()

const eventTypes = ['tunnel_down', 'tunnel_up', 'node_offline', 'traffic_exhaust', 'cert_expire', 'node_failover', 'failover_port_conflict']
const eventOptions = computed(() => eventTypes.map(e => ({key: e, value: t('alert.event_' + e)})))
const targetOptions = computed(() => [
  {key: 1, value: t('alert.targetAll')},
  {key: 2, value: t('alert.targetNode')},
  {key: 3, value: t('alert.targetTunnel')},
  {key: 4, value: t('alert.targetClient')},
])
const channelOptions = computed(() => [
  {key: 'email', value: t('alert.channelEmail')},
  {key: 'webhook', value: t('alert.channelWebhook')},
])
const eventText = (e) => t('alert.event_' + e)
const targetText = (k) => ({1: t('alert.targetAll'), 2: t('alert.targetNode'), 3: t('alert.targetTunnel'), 4: t('alert.targetClient')}[k] || t('common.unknown'))

const emptyRule = () => ({
  code: '', name: '', eventType: 'tunnel_down', targetKind: 1, targetCode: '',
  channels: ['email'], receivers: [], silenceSec: 300, enable: 1,
})

const state = ref({
  tab: 'rule',
  rule: {
    searchLoading: false, search: {page: 1, size: 10, eventType: null, enable: null},
    list: [], total: 0, minWidth: 1000,
  },
  log: {
    searchLoading: false, search: {page: 1, size: 10, eventType: null, result: null},
    list: [], total: 0, minWidth: 900,
  },
  create: {open: false, loading: false, data: emptyRule(), dataRules: {name: requiredRule(t('common.name'))}},
  update: {open: false, loading: false, data: emptyRule(), dataRules: {name: requiredRule(t('common.name'))}},
})

const ruleColumns = computed(() => [
  {type: 'selection', fixed: 'left', width: 50},
  {title: t('alert.ruleName'), key: 'name', ellipsis: {tooltip: true}},
  {title: t('alert.eventType'), key: 'eventType', width: 130, render: (row) => eventText(row.eventType)},
  {title: t('alert.targetKind'), key: 'targetKind', width: 90, render: (row) => targetText(row.targetKind)},
  {title: t('alert.targetCode'), key: 'targetCode', ellipsis: {tooltip: true}, width: 140},
  {
    title: t('alert.channels'), key: 'channels', width: 140, render(row) {
      return h('span', {}, (row.channels || []).map(c => c === 'email' ? t('alert.channelEmail') : t('alert.channelWebhook')).join(', '))
    }
  },
  {title: t('alert.silenceSec'), key: 'silenceSec', width: 110},
  {
    title: t('common.status'), key: 'enable', width: 90, render(row) {
      return h(NTag, {size: 'small', type: row.enable === 1 ? 'success' : 'default'}, {default: () => row.enable === 1 ? t('common.enabled') : t('common.disabled')})
    }
  },
  {
    title: t('common.operate'), key: 'operator', width: 130, fixed: 'right', render(row) {
      return generateButtonGroup(generateUpdateButton(row, openUpdate, t), generateDeleteButton(row, deleteFunc, t))
    }
  },
])

const logColumns = computed(() => [
  {title: t('alert.eventType'), key: 'eventType', width: 130, render: (row) => eventText(row.eventType)},
  {title: t('alert.logObject'), key: 'objectCode', ellipsis: {tooltip: true}},
  {title: t('alert.logChannel'), key: 'channel', width: 100},
  {
    title: t('alert.logResult'), key: 'result', width: 90, render(row) {
      return h(NTag, {size: 'small', type: row.result === 1 ? 'success' : 'error'}, {default: () => row.result === 1 ? t('alert.success') : t('alert.failed')})
    }
  },
  {title: t('alert.logMessage'), key: 'message', ellipsis: {tooltip: true}},
  {title: t('common.createdAt'), key: 'createdAt', width: 170},
])

const refreshRule = () => rulePageFunc()
const searchRule = () => { state.value.rule.search.page = 1; rulePageFunc() }
const rulePageFunc = async () => {
  try {
    state.value.rule.searchLoading = true
    const res = await apiNormalAlertRulePage(state.value.rule.search)
    state.value.rule.list = res.data.list || []
    state.value.rule.total = res.data.total || 0
  } finally { state.value.rule.searchLoading = false }
}
const refreshLog = () => logPageFunc()
const searchLog = () => { state.value.log.search.page = 1; logPageFunc() }
const logPageFunc = async () => {
  try {
    state.value.log.searchLoading = true
    const res = await apiNormalAlertLogPage(state.value.log.search)
    state.value.log.list = res.data.list || []
    state.value.log.total = res.data.total || 0
  } finally { state.value.log.searchLoading = false }
}

const openCreate = () => { state.value.create.data = emptyRule(); state.value.create.open = true }
const closeCreate = () => { state.value.create.open = false }
const openUpdate = (row) => {
  const d = JSON.parse(JSON.stringify(row))
  d.channels = d.channels || []
  d.receivers = d.receivers || []
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
        await apiNormalAlertRuleCreate(state.value.create.data)
        searchRule(); closeCreate()
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
        await apiNormalAlertRuleUpdate(state.value.update.data)
        searchRule(); closeUpdate()
      } finally { state.value.update.loading = false }
    }
  })
}
const deleteFunc = async (row) => {
  await apiNormalAlertRuleDelete({code: row.code})
  searchRule()
}
const batchDelete = () => {
  runBatch(t, checkedKeys.value, (code) => apiNormalAlertRuleDelete({code}), () => { clear(); searchRule() })
}

const onTabChange = (val) => {
  if (val === 'log' && state.value.log.list.length === 0) logPageFunc()
}

onBeforeMount(() => { rulePageFunc() })
</script>

<template>
  <div>
    <n-tabs type="line" :value="state.tab" @update:value="(v)=>{state.tab=v;onTabChange(v)}">
      <n-tab-pane name="rule" :tab="t('alert.ruleTab')">
        <SearchCard :show-border="false" space>
          <SearchItem type="select" :label-width="80" :label="t('alert.eventType')" :options="eventOptions"
                      :options-key-value="{key:'key',value:'value'}" clearable
                      @onChange="value => state.rule.search.eventType = value"/>
          <SearchItem type="custom">
            <n-space>
              <n-button type="info" :focusable="false" @click="searchRule">{{ t('common.search') }}</n-button>
              <n-button type="info" :focusable="false" @click="refreshRule">{{ t('common.refresh') }}</n-button>
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
          <n-data-table remote :columns="ruleColumns" :data="state.rule.list" :bordered="false"
                        :loading="state.rule.searchLoading" :scroll-x="state.rule.minWidth"
                        :row-key="row => row.code" :checked-row-keys="checkedKeys"
                        @update:checked-row-keys="onCheck">
            <template #empty>
              <n-empty :description="t('common.noData')"/>
            </template>
          </n-data-table>
        </AppCard>
        <AppCard :show-border="false">
          <n-pagination :page-size="state.rule.search.size" :page="state.rule.search.page"
                        :item-count="state.rule.total" :simple="true"
                        :on-update-page="(val) => {state.rule.search.page = val;refreshRule()}"
                        :on-update-page-size="(val) => {state.rule.search.size = val;refreshRule()}"/>
        </AppCard>
      </n-tab-pane>

      <n-tab-pane name="log" :tab="t('alert.logTab')">
        <SearchCard :show-border="false" space>
          <SearchItem type="select" :label-width="80" :label="t('alert.eventType')" :options="eventOptions"
                      :options-key-value="{key:'key',value:'value'}" clearable
                      @onChange="value => state.log.search.eventType = value"/>
          <SearchItem type="custom">
            <n-space>
              <n-button type="info" :focusable="false" @click="searchLog">{{ t('common.search') }}</n-button>
              <n-button type="info" :focusable="false" @click="refreshLog">{{ t('common.refresh') }}</n-button>
            </n-space>
          </SearchItem>
        </SearchCard>
        <AppCard>
          <n-data-table remote :columns="logColumns" :data="state.log.list" :bordered="false"
                        :loading="state.log.searchLoading" :scroll-x="state.log.minWidth">
            <template #empty>
              <n-empty :description="t('common.noData')"/>
            </template>
          </n-data-table>
        </AppCard>
        <AppCard :show-border="false">
          <n-pagination :page-size="state.log.search.size" :page="state.log.search.page"
                        :item-count="state.log.total" :simple="true"
                        :on-update-page="(val) => {state.log.search.page = val;refreshLog()}"
                        :on-update-page-size="(val) => {state.log.search.size = val;refreshLog()}"/>
        </AppCard>
      </n-tab-pane>
    </n-tabs>

    <Modal :show="state.create.open" :title="t('common.create')" width="480px" :confirm-text="t('common.save')"
           :cancel-text="t('common.cancel')" :confirm-loading="state.create.loading"
           @on-confirm="createFunc" @on-cancel="closeCreate" mask-close>
      <n-form ref="createRef" :rules="state.create.dataRules" :model="state.create.data">
        <n-form-item path="name" :label="t('alert.ruleName')">
          <n-input v-model:value.trim="state.create.data.name"/>
        </n-form-item>
        <n-form-item path="eventType" :label="t('alert.eventType')">
          <n-select v-model:value="state.create.data.eventType" :options="eventOptions"
                    :label-field="'value'" :value-field="'key'"/>
        </n-form-item>
        <n-form-item path="targetKind" :label="t('alert.targetKind')">
          <n-select v-model:value="state.create.data.targetKind" :options="targetOptions"
                    :label-field="'value'" :value-field="'key'"/>
        </n-form-item>
        <n-form-item path="targetCode" :label="t('alert.targetCode')">
          <n-input v-model:value.trim="state.create.data.targetCode"/>
        </n-form-item>
        <n-form-item path="channels" :label="t('alert.channels')">
          <n-select v-model:value="state.create.data.channels" multiple :options="channelOptions"
                    :label-field="'value'" :value-field="'key'"/>
        </n-form-item>
        <n-form-item path="receivers" :label="t('alert.receivers')">
          <n-select v-model:value="state.create.data.receivers" multiple filterable tag :options="[]"
                    :placeholder="t('alert.receiversHint')"/>
        </n-form-item>
        <n-form-item path="silenceSec" :label="t('alert.silenceSec')">
          <n-input-number v-model:value="state.create.data.silenceSec" :min="0"/>
        </n-form-item>
        <n-form-item path="enable" :label="t('common.status')">
          <n-switch :value="state.create.data.enable" :checked-value="1" :unchecked-value="2"
                    :on-update:value="value => state.create.data.enable = value">
            <template #checked>{{ t('common.enabled') }}</template>
            <template #unchecked>{{ t('common.disabled') }}</template>
          </n-switch>
        </n-form-item>
      </n-form>
    </Modal>

    <Modal :show="state.update.open" :title="t('common.update')" width="480px" :confirm-text="t('common.save')"
           :cancel-text="t('common.cancel')" :confirm-loading="state.update.loading"
           @on-confirm="updateFunc" @on-cancel="closeUpdate" mask-close>
      <n-form ref="updateRef" :rules="state.update.dataRules" :model="state.update.data">
        <n-form-item path="name" :label="t('alert.ruleName')">
          <n-input v-model:value.trim="state.update.data.name"/>
        </n-form-item>
        <n-form-item path="eventType" :label="t('alert.eventType')">
          <n-select v-model:value="state.update.data.eventType" :options="eventOptions"
                    :label-field="'value'" :value-field="'key'"/>
        </n-form-item>
        <n-form-item path="targetKind" :label="t('alert.targetKind')">
          <n-select v-model:value="state.update.data.targetKind" :options="targetOptions"
                    :label-field="'value'" :value-field="'key'"/>
        </n-form-item>
        <n-form-item path="targetCode" :label="t('alert.targetCode')">
          <n-input v-model:value.trim="state.update.data.targetCode"/>
        </n-form-item>
        <n-form-item path="channels" :label="t('alert.channels')">
          <n-select v-model:value="state.update.data.channels" multiple :options="channelOptions"
                    :label-field="'value'" :value-field="'key'"/>
        </n-form-item>
        <n-form-item path="receivers" :label="t('alert.receivers')">
          <n-select v-model:value="state.update.data.receivers" multiple filterable tag :options="[]"
                    :placeholder="t('alert.receiversHint')"/>
        </n-form-item>
        <n-form-item path="silenceSec" :label="t('alert.silenceSec')">
          <n-input-number v-model:value="state.update.data.silenceSec" :min="0"/>
        </n-form-item>
        <n-form-item path="enable" :label="t('common.status')">
          <n-switch :value="state.update.data.enable" :checked-value="1" :unchecked-value="2"
                    :on-update:value="value => state.update.data.enable = value">
            <template #checked>{{ t('common.enabled') }}</template>
            <template #unchecked>{{ t('common.disabled') }}</template>
          </n-switch>
        </n-form-item>
      </n-form>
    </Modal>
  </div>
</template>

<style scoped lang="scss">
</style>
