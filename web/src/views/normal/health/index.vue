<script setup>
import {computed, h, onBeforeMount, ref} from "vue";
import {useI18n} from "vue-i18n";
import {NTag} from "naive-ui";
import {apiNormalHealthPage, apiNormalHealthOverview} from "../../../api/normal/health.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";

const {t} = useI18n()

const state = ref({
  overview: {total: 0, healthy: 0, abnormal: 0, unknown: 0, clientOffline: 0, unregistered: 0, backendBad: 0},
  table: {
    searchLoading: false,
    search: {page: 1, size: 10, tunnelType: null, healthStatus: null},
    list: [],
    total: 0,
    minWidth: 1000,
  },
})

const typeOptions = computed(() => [
  {key: 1, value: t('health.typeHost')},
  {key: 2, value: t('health.typeForward')},
  {key: 3, value: t('health.typeProxy')},
  {key: 4, value: t('health.typeTunnel')},
  {key: 5, value: t('health.typeP2P')},
])
const statusOptions = computed(() => [
  {key: 1, value: t('health.healthy')},
  {key: 2, value: t('health.abnormal')},
  {key: 0, value: t('health.unknown')},
])

const typeText = (v) => {
  const map = {1: t('health.typeHost'), 2: t('health.typeForward'), 3: t('health.typeProxy'), 4: t('health.typeTunnel'), 5: t('health.typeP2P')}
  return map[v] || t('common.unknown')
}

const overviewCards = computed(() => [
  {label: t('health.total'), value: state.value.overview.total, type: 'default'},
  {label: t('health.healthy'), value: state.value.overview.healthy, type: 'success'},
  {label: t('health.abnormal'), value: state.value.overview.abnormal, type: 'error'},
  {label: t('health.unknown'), value: state.value.overview.unknown, type: 'default'},
  {label: t('health.clientOffline'), value: state.value.overview.clientOffline, type: 'warning'},
  {label: t('health.unregistered'), value: state.value.overview.unregistered, type: 'warning'},
  {label: t('health.backendBad'), value: state.value.overview.backendBad, type: 'error'},
])

const tableColumns = computed(() => [
  {title: t('health.tunnelName'), key: 'name', ellipsis: {tooltip: true}},
  {title: t('health.tunnelType'), key: 'tunnelType', width: 110, render: (row) => typeText(row.tunnelType)},
  {
    title: t('health.clientOnline'), key: 'clientOnline', width: 100, render(row) {
      return h(NTag, {size: 'small', type: row.clientOnline === 1 ? 'success' : 'error'}, {default: () => row.clientOnline === 1 ? t('health.online') : t('health.offline')})
    }
  },
  {
    title: t('health.tunnelRegistered'), key: 'tunnelRegistered', width: 100, render(row) {
      return h(NTag, {size: 'small', type: row.tunnelRegistered === 1 ? 'success' : 'warning'}, {default: () => row.tunnelRegistered === 1 ? t('health.registered') : t('health.notRegistered')})
    }
  },
  {
    title: t('health.backendReachable'), key: 'backendReachable', width: 110, render(row) {
      const map = {0: {type: 'default', text: t('health.reachableUnknown')}, 1: {type: 'success', text: t('health.reachableYes')}, 2: {type: 'error', text: t('health.reachableNo')}}
      const m = map[row.backendReachable] || map[0]
      return h(NTag, {size: 'small', type: m.type}, {default: () => m.text})
    }
  },
  {
    title: t('health.healthStatus'), key: 'healthStatus', width: 90, render(row) {
      const map = {0: {type: 'default', text: t('health.unknown')}, 1: {type: 'success', text: t('health.healthy')}, 2: {type: 'error', text: t('health.abnormal')}}
      const m = map[row.healthStatus] || map[0]
      return h(NTag, {size: 'small', type: m.type}, {default: () => m.text})
    }
  },
  {title: t('health.lastCheckAt'), key: 'lastCheckAt', width: 170, ellipsis: {tooltip: true}},
  {title: t('health.lastError'), key: 'lastError', ellipsis: {tooltip: true}},
])

const refreshTable = () => {
  pageFunc()
  overviewFunc()
}
const searchTable = () => {
  state.value.table.search.page = 1
  pageFunc()
}
const pageFunc = async () => {
  try {
    state.value.table.searchLoading = true
    const res = await apiNormalHealthPage(state.value.table.search)
    state.value.table.list = res.data.list || []
    state.value.table.total = res.data.total || 0
  } finally {
    state.value.table.searchLoading = false
  }
}
const overviewFunc = async () => {
  const res = await apiNormalHealthOverview({})
  if (res.data) {
    state.value.overview = res.data
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
      <SearchItem type="select" :label-width="80" :label="t('health.tunnelType')" :options="typeOptions"
                  :options-key-value="{key:'key',value:'value'}" clearable
                  @onChange="value => state.table.search.tunnelType = value"/>
      <SearchItem type="select" :label-width="80" :label="t('health.healthStatus')" :options="statusOptions"
                  :options-key-value="{key:'key',value:'value'}" clearable
                  @onChange="value => state.table.search.healthStatus = value"/>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" @click="searchTable">{{ t('common.search') }}</n-button>
          <n-button type="info" :focusable="false" @click="refreshTable">{{ t('common.refresh') }}</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>

    <AppCard>
      <n-data-table remote :columns="tableColumns" :data="state.table.list" :bordered="false"
                    :loading="state.table.searchLoading" :scroll-x="state.table.minWidth">
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
  </div>
</template>

<style scoped lang="scss">
</style>
