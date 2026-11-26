<script setup>
import {onBeforeMount, onBeforeUnmount, ref} from "vue";
import {useRoute} from "vue-router";
import router from "../../../router/index.js";
import {apiAdminGostClientLoggerPage} from "../../../api/admin/gost_client_logger.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";
import Empty from "../../../components/Empty.vue";
import {NSpace} from "naive-ui";

const route = useRoute()
const clientCode = route.query.clientCode || ''

const state = ref({
  table: {
    searchLoading: false,
    search: {
      page: 1,
      size: 20,
      clientCode: clientCode,
      type: '',
      keyword: '',
    },
    list: [],
    total: 0,
    minWidth: 900,
    error: '',
  },
  autoRefresh: false,
})

let timer = null

const pageFunc = async () => {
  try {
    state.value.table.searchLoading = true
    let res = await apiAdminGostClientLoggerPage(state.value.table.search)
    state.value.table.list = res.data.list || []
    state.value.table.total = res.data.total
    state.value.table.error = ''
  } catch (e) {
    state.value.table.list = []
    state.value.table.total = 0
    state.value.table.error = e?.message || '获取日志失败'
  } finally {
    state.value.table.searchLoading = false
  }
}

const refreshTable = () => {
  pageFunc()
}

const searchTable = () => {
  state.value.table.search.page = 1
  pageFunc()
}

const toggleAutoRefresh = (val) => {
  state.value.autoRefresh = val
  if (timer) {
    clearInterval(timer)
    timer = null
  }
  if (val) {
    timer = setInterval(() => pageFunc(), 5000)
  }
}

const back = () => {
  router.back()
}

const tableColumns = [
  {title: '时间', key: 'timestamp', width: 180},
  {title: '类型', key: 'type', width: 100},
  {title: '内容', key: 'message', ellipsis: {tooltip: true}},
]

onBeforeMount(() => {
  if (!clientCode) {
    state.value.table.error = '缺少客户端参数'
    return
  }
  pageFunc()
})

onBeforeUnmount(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<template>
  <div>
    <SearchCard :show-border="false" space>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" @click="back">返回</n-button>
        </n-space>
      </SearchItem>
      <SearchItem
          type="select"
          :label-width="70"
          label="类型"
          :default="''"
          :options="[
              {value:'全部',key:''},
              {value:'FRP',key:'FRP'},
              {value:'Client',key:'Client'},
              {value:'Node',key:'Node'}
          ]"
          @onChange="value => state.table.search.type=value"
      ></SearchItem>
      <SearchItem
          type="input"
          :label-width="70"
          clearable
          label="关键词"
          @onChange="value => state.table.search.keyword=value"
      ></SearchItem>
      <SearchItem type="custom">
        <n-space align="center">
          <n-button type="info" :focusable="false" @click="searchTable">搜索</n-button>
          <n-button type="info" :focusable="false" @click="refreshTable">刷新</n-button>
          <n-switch :value="state.autoRefresh" @update:value="toggleAutoRefresh">
            <template #checked>自动刷新</template>
            <template #unchecked>自动刷新</template>
          </n-switch>
        </n-space>
      </SearchItem>
    </SearchCard>

    <AppCard>
      <Empty v-if="state.table.list.length===0 && !state.table.searchLoading" border
             :description="state.table.error || '暂无日志'"></Empty>
      <n-data-table
          v-else
          remote
          :columns="tableColumns"
          :data="state.table.list"
          :bordered="false"
          :loading="state.table.searchLoading"
          :scroll-x="state.table.minWidth"
      />
    </AppCard>

    <AppCard :show-border="false">
      <n-pagination
          :page-size="state.table.search.size"
          :page="state.table.search.page"
          :item-count="state.table.total"
          :simple="true"
          :on-update-page="(val) => {state.table.search.page = val;refreshTable()}"
          :on-update-page-size="(val) => {state.table.search.size = val;refreshTable()}"
      />
    </AppCard>
  </div>
</template>

<style scoped lang="scss">

</style>