<script setup>
import {computed, h, onBeforeMount, ref} from "vue";
import {NTag} from "naive-ui";
import {apiAdminHealthPage, apiAdminHealthOverview} from "../../../api/admin/health.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";

const typeOptions = [
  {key: 1, value: '域名解析'},
  {key: 2, value: '端口转发'},
  {key: 3, value: '代理隧道'},
  {key: 4, value: '私有隧道'},
  {key: 5, value: 'P2P'},
]
const statusOptions = [
  {key: 1, value: '健康'},
  {key: 2, value: '异常'},
  {key: 0, value: '未知'},
]
const typeText = (v) => ({1: '域名解析', 2: '端口转发', 3: '代理隧道', 4: '私有隧道', 5: 'P2P'}[v] || '未知')

const state = ref({
  overview: {total: 0, healthy: 0, abnormal: 0, unknown: 0, clientOffline: 0, unregistered: 0, backendBad: 0},
  table: {
    searchLoading: false,
    search: {page: 1, size: 10, userCode: null, nodeCode: null, tunnelType: null, healthStatus: null, keyword: null},
    list: [],
    total: 0,
    minWidth: 1400,
  },
})

const overviewCards = computed(() => [
  {label: '总数', value: state.value.overview.total, type: 'default'},
  {label: '健康', value: state.value.overview.healthy, type: 'success'},
  {label: '异常', value: state.value.overview.abnormal, type: 'error'},
  {label: '未知', value: state.value.overview.unknown, type: 'default'},
  {label: '客户端离线', value: state.value.overview.clientOffline, type: 'warning'},
  {label: '未注册', value: state.value.overview.unregistered, type: 'warning'},
  {label: '后端不可达', value: state.value.overview.backendBad, type: 'error'},
])

const tableColumns = computed(() => [
  {
    title: '隧道', key: 'name', ellipsis: {tooltip: true}, minWidth: 180, render(row) {
      return h('div', {}, [
        h('div', {style: 'font-weight:500'}, row.name || '-'),
        h('div', {style: 'font-size:12px;color:#999'}, row.tunnelCode),
      ])
    }
  },
  {title: '类型', key: 'tunnelType', width: 100, render: (row) => typeText(row.tunnelType)},
  {title: '用户', key: 'account', width: 120, ellipsis: {tooltip: true}, render: (row) => row.account || row.userCode || '-'},
  {title: '节点', key: 'nodeName', width: 120, ellipsis: {tooltip: true}, render: (row) => row.nodeName || row.nodeCode || '-'},
  {
    title: '客户端在线', key: 'clientOnline', width: 100, render(row) {
      return h(NTag, {size: 'small', type: row.clientOnline === 1 ? 'success' : 'error'}, {default: () => row.clientOnline === 1 ? '在线' : '离线'})
    }
  },
  {
    title: '已注册', key: 'tunnelRegistered', width: 90, render(row) {
      return h(NTag, {size: 'small', type: row.tunnelRegistered === 1 ? 'success' : 'warning'}, {default: () => row.tunnelRegistered === 1 ? '已注册' : '未注册'})
    }
  },
  {
    title: '后端可达', key: 'backendReachable', width: 100, render(row) {
      const map = {0: {type: 'default', text: '未探测'}, 1: {type: 'success', text: '可达'}, 2: {type: 'error', text: '不可达'}}
      const m = map[row.backendReachable] || map[0]
      return h(NTag, {size: 'small', type: m.type}, {default: () => m.text})
    }
  },
  {
    title: '健康状态', key: 'healthStatus', width: 90, render(row) {
      const map = {0: {type: 'default', text: '未知'}, 1: {type: 'success', text: '健康'}, 2: {type: 'error', text: '异常'}}
      const m = map[row.healthStatus] || map[0]
      return h(NTag, {size: 'small', type: m.type}, {default: () => m.text})
    }
  },
  {title: '最近检查', key: 'lastCheckAt', width: 170, ellipsis: {tooltip: true}},
  {title: '错误', key: 'lastError', ellipsis: {tooltip: true}, minWidth: 160},
])

const overviewFunc = async () => {
  const s = state.value.table.search
  const res = await apiAdminHealthOverview({userCode: s.userCode, nodeCode: s.nodeCode, tunnelType: s.tunnelType})
  if (res.data) {
    state.value.overview = res.data
  }
}
const refreshTable = () => {
  pageFunc()
  overviewFunc()
}
const searchTable = () => {
  state.value.table.search.page = 1
  pageFunc()
  overviewFunc()
}
const pageFunc = async () => {
  try {
    state.value.table.searchLoading = true
    const res = await apiAdminHealthPage(state.value.table.search)
    state.value.table.list = res.data.list || []
    state.value.table.total = res.data.total || 0
  } finally {
    state.value.table.searchLoading = false
  }
}

onBeforeMount(() => {
  pageFunc()
  overviewFunc()
})
</script>

<template>
  <div>
    <AppCard :show-border="false">
      <n-grid cols="2 520:4 900:7" responsive="screen" :x-gap="12" :y-gap="12">
        <n-grid-item v-for="card in overviewCards" :key="card.label">
          <n-statistic :label="card.label" :value="card.value">
            <template #prefix>
              <n-tag size="small" :type="card.type" :bordered="false" style="margin-right:6px">●</n-tag>
            </template>
          </n-statistic>
        </n-grid-item>
      </n-grid>
    </AppCard>

    <SearchCard :show-border="false" space>
      <SearchItem type="input" :label-width="80" label="用户编码" clearable
                  @onChange="value => state.table.search.userCode = value"/>
      <SearchItem type="input" :label-width="80" label="节点编码" clearable
                  @onChange="value => state.table.search.nodeCode = value"/>
      <SearchItem type="select" :label-width="80" label="隧道类型" :options="typeOptions"
                  :options-key-value="{key:'key',value:'value'}" clearable
                  @onChange="value => state.table.search.tunnelType = value"/>
      <SearchItem type="select" :label-width="80" label="健康状态" :options="statusOptions"
                  :options-key-value="{key:'key',value:'value'}" clearable
                  @onChange="value => state.table.search.healthStatus = value"/>
      <SearchItem type="input" :label-width="80" label="关键字" clearable
                  @onChange="value => state.table.search.keyword = value"/>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" @click="searchTable">搜索</n-button>
          <n-button type="info" :focusable="false" @click="refreshTable">刷新</n-button>
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
  </div>
</template>

<style scoped lang="scss">
</style>
