<script setup>
import {computed, h, onBeforeMount, ref} from "vue";
import {NTag, NButton, NPopconfirm, NSpace} from "naive-ui";
import {apiAdminAlertRulePage, apiAdminAlertRuleDelete, apiAdminAlertLogPage} from "../../../api/admin/alert.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";

const eventTypes = ['tunnel_down', 'tunnel_up', 'node_offline', 'traffic_exhaust', 'cert_expire', 'node_failover']
const eventTextMap = {
  tunnel_down: '隧道离线', tunnel_up: '隧道恢复', node_offline: '节点离线',
  traffic_exhaust: '流量耗尽', cert_expire: '证书到期', node_failover: '节点故障转移',
}
const eventOptions = eventTypes.map(e => ({key: e, value: eventTextMap[e]}))
const eventText = (e) => eventTextMap[e] || e
const targetText = (k) => ({1: '全部', 2: '节点', 3: '隧道', 4: '客户端'}[k] || '未知')
const channelText = (c) => (c === 'email' ? '邮件' : 'Webhook')
const enableOptions = [{key: 1, value: '启用'}, {key: 2, value: '停用'}]
const resultOptions = [{key: 1, value: '成功'}, {key: 2, value: '失败'}]

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
  {title: '名称', key: 'name', ellipsis: {tooltip: true}, minWidth: 140},
  {title: '用户', key: 'account', width: 120, ellipsis: {tooltip: true}, render: (row) => row.account || row.userCode || '-'},
  {title: '事件', key: 'eventType', width: 130, render: (row) => eventText(row.eventType)},
  {title: '目标', key: 'targetKind', width: 80, render: (row) => targetText(row.targetKind)},
  {
    title: '渠道', key: 'channels', width: 130, render(row) {
      return h('span', {}, (row.channels || []).map(channelText).join(', '))
    }
  },
  {title: '静默(秒)', key: 'silenceSec', width: 90},
  {
    title: '状态', key: 'enable', width: 80, render(row) {
      return h(NTag, {size: 'small', type: row.enable === 1 ? 'success' : 'default'}, {default: () => row.enable === 1 ? '启用' : '停用'})
    }
  },
  {title: '创建时间', key: 'createdAt', width: 170, ellipsis: {tooltip: true}},
  {
    title: '操作', key: 'operator', width: 90, fixed: 'right', render(row) {
      return h(NPopconfirm, {onPositiveClick: () => deleteFunc(row)}, {
        trigger: () => h(NButton, {size: 'small', type: 'error', focusable: false}, {default: () => '删除'}),
        default: () => '确认删除该告警规则？',
      })
    }
  },
])

const logColumns = computed(() => [
  {title: '用户', key: 'account', width: 120, ellipsis: {tooltip: true}, render: (row) => row.account || row.userCode || '-'},
  {title: '事件', key: 'eventType', width: 130, render: (row) => eventText(row.eventType)},
  {
    title: '对象', key: 'objectCode', ellipsis: {tooltip: true}, minWidth: 160, render(row) {
      return h('span', {}, `${row.objectKind || ''} ${row.objectCode || ''}`.trim())
    }
  },
  {title: '渠道', key: 'channel', width: 90, render: (row) => channelText(row.channel)},
  {
    title: '结果', key: 'result', width: 80, render(row) {
      return h(NTag, {size: 'small', type: row.result === 1 ? 'success' : 'error'}, {default: () => row.result === 1 ? '成功' : '失败'})
    }
  },
  {title: '消息', key: 'message', ellipsis: {tooltip: true}, minWidth: 180},
  {title: '时间', key: 'createdAt', width: 170, ellipsis: {tooltip: true}},
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

const onTabChange = (val) => {
  if (val === 'log' && state.value.log.list.length === 0) logPageFunc()
}

onBeforeMount(() => { rulePageFunc() })
</script>

<template>
  <div>
    <n-tabs type="line" :value="state.tab" @update:value="(v)=>{state.tab=v;onTabChange(v)}">
      <n-tab-pane name="rule" tab="告警规则">
        <SearchCard :show-border="false" space>
          <SearchItem type="input" :label-width="80" label="用户编码" clearable
                      @onChange="value => state.rule.search.userCode = value"/>
          <SearchItem type="select" :label-width="80" label="事件类型" :options="eventOptions"
                      :options-key-value="{key:'key',value:'value'}" clearable
                      @onChange="value => state.rule.search.eventType = value"/>
          <SearchItem type="select" :label-width="80" label="状态" :options="enableOptions"
                      :options-key-value="{key:'key',value:'value'}" clearable
                      @onChange="value => state.rule.search.enable = value"/>
          <SearchItem type="custom">
            <n-space>
              <n-button type="info" :focusable="false" @click="searchRule">搜索</n-button>
              <n-button type="info" :focusable="false" @click="refreshRule">刷新</n-button>
            </n-space>
          </SearchItem>
        </SearchCard>
        <AppCard>
          <n-data-table remote :columns="ruleColumns" :data="state.rule.list" :bordered="false"
                        :loading="state.rule.searchLoading" :scroll-x="state.rule.minWidth"/>
        </AppCard>
        <AppCard :show-border="false">
          <n-pagination :page-size="state.rule.search.size" :page="state.rule.search.page"
                        :item-count="state.rule.total" :simple="true"
                        :on-update-page="(val) => {state.rule.search.page = val;refreshRule()}"
                        :on-update-page-size="(val) => {state.rule.search.size = val;refreshRule()}"/>
        </AppCard>
      </n-tab-pane>

      <n-tab-pane name="log" tab="发送记录">
        <SearchCard :show-border="false" space>
          <SearchItem type="input" :label-width="80" label="用户编码" clearable
                      @onChange="value => state.log.search.userCode = value"/>
          <SearchItem type="select" :label-width="80" label="事件类型" :options="eventOptions"
                      :options-key-value="{key:'key',value:'value'}" clearable
                      @onChange="value => state.log.search.eventType = value"/>
          <SearchItem type="select" :label-width="80" label="结果" :options="resultOptions"
                      :options-key-value="{key:'key',value:'value'}" clearable
                      @onChange="value => state.log.search.result = value"/>
          <SearchItem type="custom">
            <n-space>
              <n-button type="info" :focusable="false" @click="searchLog">搜索</n-button>
              <n-button type="info" :focusable="false" @click="refreshLog">刷新</n-button>
            </n-space>
          </SearchItem>
        </SearchCard>
        <AppCard>
          <n-data-table remote :columns="logColumns" :data="state.log.list" :bordered="false"
                        :loading="state.log.searchLoading" :scroll-x="state.log.minWidth"/>
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
