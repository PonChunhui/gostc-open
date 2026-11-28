<script setup>
import {computed, h, onBeforeMount, ref} from "vue";
import {NTag} from "naive-ui";
import {apiAdminFailoverLogPage} from "../../../api/admin/failover.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";

const typeTextMap = {1: '域名解析', 2: '端口转发', 3: '私有代理', 4: '私有隧道', 5: 'P2P'}
const typeText = (k) => typeTextMap[k] || '未知'
const reasonTextMap = {failover: '故障转移', failback: '自动回切', port_conflict: '端口冲突放弃'}
const reasonTypeMap = {failover: 'info', failback: 'success', port_conflict: 'warning'}
const reasonText = (r) => reasonTextMap[r] || r
const reasonOptions = Object.keys(reasonTextMap).map(k => ({key: k, value: reasonTextMap[k]}))

const state = ref({
  searchLoading: false,
  search: {page: 1, size: 10, userCode: null, tunnelCode: null, reason: null},
  list: [], total: 0, minWidth: 1300,
})

const columns = computed(() => [
  {title: '时间', key: 'createdAt', width: 170, ellipsis: {tooltip: true}},
  {title: '用户', key: 'account', width: 120, ellipsis: {tooltip: true}, render: (row) => row.account || row.userCode || '-'},
  {title: '隧道', key: 'tunnelCode', ellipsis: {tooltip: true}, minWidth: 140},
  {title: '类型', key: 'tunnelType', width: 100, render: (row) => typeText(row.tunnelType)},
  {title: '原节点', key: 'fromName', ellipsis: {tooltip: true}, minWidth: 120, render: (row) => row.fromName || row.fromNode || '-'},
  {title: '新节点', key: 'toName', ellipsis: {tooltip: true}, minWidth: 120, render: (row) => row.toName || row.toNode || '-'},
  {title: '原入口', key: 'oldEntry', ellipsis: {tooltip: true}, minWidth: 160},
  {title: '新入口', key: 'newEntry', ellipsis: {tooltip: true}, minWidth: 160},
  {
    title: '原因', key: 'reason', width: 120, render(row) {
      return h(NTag, {size: 'small', type: reasonTypeMap[row.reason] || 'default'}, {default: () => reasonText(row.reason)})
    }
  },
])

const refresh = () => pageFunc()
const search = () => { state.value.search.page = 1; pageFunc() }
const pageFunc = async () => {
  try {
    state.value.searchLoading = true
    const res = await apiAdminFailoverLogPage(state.value.search)
    state.value.list = res.data.list || []
    state.value.total = res.data.total || 0
  } finally { state.value.searchLoading = false }
}

onBeforeMount(() => { pageFunc() })
</script>

<template>
  <div>
    <SearchCard :show-border="false" space>
      <SearchItem type="input" :label-width="80" label="用户编码" clearable
                  @onChange="value => state.search.userCode = value"/>
      <SearchItem type="select" :label-width="80" label="原因" :options="reasonOptions"
                  :options-key-value="{key:'key',value:'value'}" clearable
                  @onChange="value => state.search.reason = value"/>
      <SearchItem type="input" :label-width="80" label="隧道编码" clearable
                  @onChange="value => state.search.tunnelCode = value"/>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" @click="search">查询</n-button>
          <n-button type="info" :focusable="false" @click="refresh">刷新</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>
    <AppCard>
      <n-data-table remote :columns="columns" :data="state.list" :bordered="false"
                    :loading="state.searchLoading" :scroll-x="state.minWidth"/>
    </AppCard>
    <AppCard :show-border="false">
      <n-pagination :page-size="state.search.size" :page="state.search.page"
                    :item-count="state.total" :simple="true"
                    :on-update-page="(val) => {state.search.page = val;refresh()}"
                    :on-update-page-size="(val) => {state.search.size = val;refresh()}"/>
    </AppCard>
  </div>
</template>

<style scoped lang="scss">
</style>
