<script setup>
import {computed, h, onBeforeMount, ref} from "vue";
import {useI18n} from "vue-i18n";
import {NTag} from "naive-ui";
import {apiAdminFailoverLogPage} from "../../../api/admin/failover.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";

const {t} = useI18n()

const typeText = (k) => ({
  1: t('failover.typeHost'), 2: t('failover.typeForward'), 3: t('failover.typeProxy'),
  4: t('failover.typeTunnel'), 5: t('failover.typeP2P'),
}[k] || t('common.unknown'))
const reasonText = (r) => ({
  failover: t('failover.reasonFailover'), failback: t('failover.reasonFailback'), port_conflict: t('failover.reasonPortConflict'),
}[r] || r)
const reasonTypeMap = {failover: 'info', failback: 'success', port_conflict: 'warning'}
const reasonOptions = computed(() => [
  {key: 'failover', value: t('failover.reasonFailover')},
  {key: 'failback', value: t('failover.reasonFailback')},
  {key: 'port_conflict', value: t('failover.reasonPortConflict')},
])

const state = ref({
  searchLoading: false,
  search: {page: 1, size: 10, userCode: null, tunnelCode: null, reason: null},
  list: [], total: 0, minWidth: 1300,
})

const columns = computed(() => [
  {title: t('common.time'), key: 'createdAt', width: 170, ellipsis: {tooltip: true}},
  {title: t('common.user'), key: 'account', width: 120, ellipsis: {tooltip: true}, render: (row) => row.account || row.userCode || '-'},
  {title: t('failover.tunnel'), key: 'tunnelCode', ellipsis: {tooltip: true}, minWidth: 140},
  {title: t('failover.type'), key: 'tunnelType', width: 100, render: (row) => typeText(row.tunnelType)},
  {title: t('failover.fromNode'), key: 'fromName', ellipsis: {tooltip: true}, minWidth: 120, render: (row) => row.fromName || row.fromNode || '-'},
  {title: t('failover.toNode'), key: 'toName', ellipsis: {tooltip: true}, minWidth: 120, render: (row) => row.toName || row.toNode || '-'},
  {title: t('failover.oldEntry'), key: 'oldEntry', ellipsis: {tooltip: true}, minWidth: 160},
  {title: t('failover.newEntry'), key: 'newEntry', ellipsis: {tooltip: true}, minWidth: 160},
  {
    title: t('failover.reason'), key: 'reason', width: 120, render(row) {
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
      <SearchItem type="input" :label-width="80" :label="t('common.userCode')" clearable
                  @onChange="value => state.search.userCode = value"/>
      <SearchItem type="select" :label-width="80" :label="t('failover.reason')" :options="reasonOptions"
                  :options-key-value="{key:'key',value:'value'}" clearable
                  @onChange="value => state.search.reason = value"/>
      <SearchItem type="input" :label-width="80" :label="t('common.tunnelCode')" clearable
                  @onChange="value => state.search.tunnelCode = value"/>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" @click="search">{{ t('common.query') }}</n-button>
          <n-button type="info" :focusable="false" @click="refresh">{{ t('common.refresh') }}</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>
    <AppCard>
      <n-data-table remote :columns="columns" :data="state.list" :bordered="false"
                    :loading="state.searchLoading" :scroll-x="state.minWidth">
        <template #empty>
          <n-empty :description="t('common.noData')"/>
        </template>
      </n-data-table>
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
