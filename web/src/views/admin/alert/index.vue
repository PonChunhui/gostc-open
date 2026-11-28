<script setup>
import {computed, h, onBeforeMount, ref} from "vue";
import {useI18n} from "vue-i18n";
import {NTag} from "naive-ui";
import {apiAdminAlertRulePage, apiAdminAlertRuleDelete, apiAdminAlertLogPage} from "../../../api/admin/alert.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";
import {generateButtonGroup, generateDeleteButton} from "../../../utils/tableHelper.js";
import {useTableBatch, runBatch} from "../../../utils/useTableBatch.js";

const {t} = useI18n()
const {checkedKeys, onCheck, clear} = useTableBatch()

const eventTypes = ['tunnel_down', 'tunnel_up', 'node_offline', 'traffic_exhaust', 'cert_expire', 'node_failover', 'failover_port_conflict']
const eventOptions = computed(() => eventTypes.map(e => ({key: e, value: t('alert.event_' + e)})))
const eventText = (e) => t('alert.event_' + e)
const targetText = (k) => ({1: t('alert.targetAll'), 2: t('alert.targetNode'), 3: t('alert.targetTunnel'), 4: t('alert.targetClient')}[k] || t('common.unknown'))
const channelText = (c) => (c === 'email' ? t('alert.channelEmail') : t('alert.channelWebhook'))
const enableOptions = computed(() => [{key: 1, value: t('common.enable')}, {key: 2, value: t('common.disable')}])
const resultOptions = computed(() => [{key: 1, value: t('alert.success')}, {key: 2, value: t('alert.failed')}])

const state = ref({
  tab: 'rule',
  rule: {
    searchLoading: false, search: {page: 1, size: 10, userCode: null, eventType: null, enable: null},
    list: [], total: 0, minWidth: 1200,
  },
  log: {
    searchLoading: false, search: {page: 1, size: 10, userCode: null, eventType: null, result: null},
    list: [], total: 0, minWidth: 1100,
  },
})

const ruleColumns = computed(() => [
  {type: 'selection', fixed: 'left', width: 50},
  {title: t('common.name'), key: 'name', ellipsis: {tooltip: true}, minWidth: 140},
  {title: t('common.user'), key: 'account', width: 120, ellipsis: {tooltip: true}, render: (row) => row.account || row.userCode || '-'},
  {title: t('alert.eventType'), key: 'eventType', width: 130, render: (row) => eventText(row.eventType)},
  {title: t('alert.targetKind'), key: 'targetKind', width: 80, render: (row) => targetText(row.targetKind)},
  {
    title: t('alert.channels'), key: 'channels', width: 130, render(row) {
      return h('span', {}, (row.channels || []).map(channelText).join(', '))
    }
  },
  {title: t('alert.silenceSec'), key: 'silenceSec', width: 90},
  {
    title: t('common.status'), key: 'enable', width: 80, render(row) {
      return h(NTag, {size: 'small', type: row.enable === 1 ? 'success' : 'default'}, {default: () => row.enable === 1 ? t('common.enabled') : t('common.disabled')})
    }
  },
  {title: t('common.createdAt'), key: 'createdAt', width: 170, ellipsis: {tooltip: true}},
  {
    title: t('common.operate'), key: 'operator', width: 90, fixed: 'right', render(row) {
      return generateButtonGroup(generateDeleteButton(row, deleteFunc, t))
    }
  },
])

const logColumns = computed(() => [
  {title: t('common.user'), key: 'account', width: 120, ellipsis: {tooltip: true}, render: (row) => row.account || row.userCode || '-'},
  {title: t('alert.eventType'), key: 'eventType', width: 130, render: (row) => eventText(row.eventType)},
  {
    title: t('alert.logObject'), key: 'objectCode', ellipsis: {tooltip: true}, minWidth: 160, render(row) {
      return h('span', {}, `${row.objectKind || ''} ${row.objectCode || ''}`.trim())
    }
  },
  {title: t('alert.logChannel'), key: 'channel', width: 90, render: (row) => channelText(row.channel)},
  {
    title: t('alert.logResult'), key: 'result', width: 80, render(row) {
      return h(NTag, {size: 'small', type: row.result === 1 ? 'success' : 'error'}, {default: () => row.result === 1 ? t('alert.success') : t('alert.failed')})
    }
  },
  {title: t('alert.logMessage'), key: 'message', ellipsis: {tooltip: true}, minWidth: 180},
  {title: t('common.time'), key: 'createdAt', width: 170, ellipsis: {tooltip: true}},
])

const refreshRule = () => rulePageFunc()
const searchRule = () => { state.value.rule.search.page = 1; rulePageFunc() }
const rulePageFunc = async () => {
  try {
    state.value.rule.searchLoading = true
    const res = await apiAdminAlertRulePage(state.value.rule.search)
    state.value.rule.list = res.data.list || []
    state.value.rule.total = res.data.total || 0
  } finally { state.value.rule.searchLoading = false }
}
const refreshLog = () => logPageFunc()
const searchLog = () => { state.value.log.search.page = 1; logPageFunc() }
const logPageFunc = async () => {
  try {
    state.value.log.searchLoading = true
    const res = await apiAdminAlertLogPage(state.value.log.search)
    state.value.log.list = res.data.list || []
    state.value.log.total = res.data.total || 0
  } finally { state.value.log.searchLoading = false }
}

const deleteFunc = async (row) => {
  await apiAdminAlertRuleDelete({code: row.code})
  searchRule()
}
const batchDelete = () => {
  runBatch(t, checkedKeys.value, (code) => apiAdminAlertRuleDelete({code}), () => { clear(); searchRule() })
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
          <SearchItem type="input" :label-width="80" :label="t('common.userCode')" clearable
                      @onChange="value => state.rule.search.userCode = value"/>
          <SearchItem type="select" :label-width="80" :label="t('alert.eventType')" :options="eventOptions"
                      :options-key-value="{key:'key',value:'value'}" clearable
                      @onChange="value => state.rule.search.eventType = value"/>
          <SearchItem type="select" :label-width="80" :label="t('common.status')" :options="enableOptions"
                      :options-key-value="{key:'key',value:'value'}" clearable
                      @onChange="value => state.rule.search.enable = value"/>
          <SearchItem type="custom">
            <n-space>
              <n-button type="info" :focusable="false" @click="searchRule">{{ t('common.search') }}</n-button>
              <n-button type="info" :focusable="false" @click="refreshRule">{{ t('common.refresh') }}</n-button>
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
          <SearchItem type="input" :label-width="80" :label="t('common.userCode')" clearable
                      @onChange="value => state.log.search.userCode = value"/>
          <SearchItem type="select" :label-width="80" :label="t('alert.eventType')" :options="eventOptions"
                      :options-key-value="{key:'key',value:'value'}" clearable
                      @onChange="value => state.log.search.eventType = value"/>
          <SearchItem type="select" :label-width="80" :label="t('alert.logResult')" :options="resultOptions"
                      :options-key-value="{key:'key',value:'value'}" clearable
                      @onChange="value => state.log.search.result = value"/>
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
  </div>
</template>

<style scoped lang="scss">
</style>
