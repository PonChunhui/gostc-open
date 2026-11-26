<script setup>
import {onBeforeMount, ref} from "vue";
import {
  apiAdminSystemCdkBatchCreate,
  apiAdminSystemCdkBatchDelete,
  apiAdminSystemCdkBatchPage,
  apiAdminSystemCdkCardBan,
  apiAdminSystemCdkCardExport,
  apiAdminSystemCdkCardPage,
} from "../../../api/admin/system_cdk.js";
import {apiAdminSystemUserGroupList} from "../../../api/admin/system_user_group.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";
import {generateBatchTableColumns, generateCardTableColumns} from "./tableColumns.js";
import Modal from "../../../components/Modal.vue";
import {requiredNumberRule, requiredRule} from "../../../utils/formDataRule.js";

const emptyData = () => ({
  name: '',
  remark: '',
  groupCode: null,
  groupDays: 0,
  amount: 0,
  count: 10,
})

const state = ref({
  table: {
    searchLoading: false,
    search: {
      page: 1,
      size: 10,
      name: '',
    },
    list: [],
    total: 0,
    minWidth: 1100,
  },
  create: {
    open: false,
    loading: false,
    data: emptyData(),
    dataRules: {
      name: requiredRule('批次名称不能为空'),
      count: requiredNumberRule('生成数量不能为空'),
    },
  },
  card: {
    open: false,
    batchCode: '',
    batchName: '',
    searchLoading: false,
    search: {
      page: 1,
      size: 10,
      batchCode: '',
      status: null,
      secret: '',
      account: '',
    },
    list: [],
    total: 0,
    minWidth: 900,
  },
})

const groupOptions = ref([])
const loadOptions = async () => {
  const res = await apiAdminSystemUserGroupList({})
  groupOptions.value = (res.data || []).map(g => ({label: g.name, value: g.code}))
}

const statusOptions = [
  {label: '未使用', value: 1},
  {label: '已兑换', value: 2},
  {label: '已停用', value: 3},
]

const refreshTable = () => {
  pageFunc()
}

const searchTable = () => {
  state.value.table.search.page = 1
  pageFunc()
}

const pageFunc = async () => {
  try {
    state.value.table.searchLoading = true
    let res = await apiAdminSystemCdkBatchPage(state.value.table.search)
    state.value.table.list = res.data.list
    state.value.table.total = res.data.total
  } finally {
    state.value.table.searchLoading = false
  }
}

const openCreate = () => {
  state.value.create.data = emptyData()
  state.value.create.open = true
}

const closeCreate = () => {
  state.value.create.open = false
}

const createRef = ref()
const createFunc = () => {
  createRef.value.validate(async valid => {
    if (!valid) {
      const data = state.value.create.data
      if (!data.groupCode && !(parseFloat(data.amount) > 0)) {
        $message.create('请至少配置发放用户组或发放积分', {type: 'error'})
        return
      }
      try {
        state.value.create.loading = true
        await apiAdminSystemCdkBatchCreate({
          ...data,
          groupCode: data.groupCode || '',
        })
        searchTable()
        closeCreate()
      } finally {
        state.value.create.loading = false
      }
    }
  })
}

const deleteFunc = async (row) => {
  await apiAdminSystemCdkBatchDelete({code: row.code})
  searchTable()
}

const exportFunc = async (batchCode, batchName) => {
  const res = await apiAdminSystemCdkCardExport({batchCode})
  const secrets = res.data || []
  if (!secrets.length) {
    $message.create('该批次没有可导出的未使用卡密', {type: 'warning'})
    return
  }
  const blob = new Blob([secrets.join('\n')], {type: 'text/plain;charset=utf-8'})
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${batchName || batchCode}-卡密.txt`
  a.click()
  URL.revokeObjectURL(url)
}

const openCard = (row) => {
  state.value.card.batchCode = row.code
  state.value.card.batchName = row.name
  state.value.card.search = {
    page: 1,
    size: 10,
    batchCode: row.code,
    status: null,
    secret: '',
    account: '',
  }
  state.value.card.open = true
  cardPageFunc()
}

const closeCard = () => {
  state.value.card.open = false
}

const cardSearchTable = () => {
  state.value.card.search.page = 1
  cardPageFunc()
}

const cardRefreshTable = () => {
  cardPageFunc()
}

const cardPageFunc = async () => {
  try {
    state.value.card.searchLoading = true
    const search = state.value.card.search
    let res = await apiAdminSystemCdkCardPage({
      ...search,
      status: search.status || 0,
    })
    state.value.card.list = res.data.list
    state.value.card.total = res.data.total
  } finally {
    state.value.card.searchLoading = false
  }
}

const banFunc = async (row, toStatus) => {
  await apiAdminSystemCdkCardBan({code: row.code, status: toStatus})
  cardPageFunc()
}

const batchTableColumns = generateBatchTableColumns(openCard, (row) => exportFunc(row.code, row.name), deleteFunc)
const cardTableColumns = generateCardTableColumns(banFunc)

onBeforeMount(() => {
  pageFunc()
  loadOptions()
})

</script>

<template>
  <div>
    <SearchCard :show-border="false" space>
      <SearchItem
          type="input"
          :label-width="70"
          clearable
          label="批次名称"
          @onChange="value => state.table.search.name=value"
      ></SearchItem>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" @click="searchTable">搜索</n-button>
          <n-button type="info" :focusable="false" @click="refreshTable">刷新</n-button>
          <n-button type="info" :focusable="false" @click="openCreate">新增批次</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>

    <AppCard>
      <n-data-table
          remote
          :columns="batchTableColumns"
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

    <Modal
        :show="state.create.open"
        title="新增卡密批次"
        width="560px"
        confirm-text="生成"
        cancel-text="取消"
        :confirm-loading="state.create.loading"
        @on-confirm="createFunc"
        @on-cancel="closeCreate"
        mask-close
    >
      <n-form ref="createRef" :rules="state.create.dataRules" :model="state.create.data">
        <n-form-item path="name" label="批次名称">
          <n-input v-model:value.trim="state.create.data.name" placeholder="如：Pro月卡-2026年1月"></n-input>
        </n-form-item>
        <n-form-item path="groupCode" label="发放用户组(留空=不发)">
          <n-select
              v-model:value="state.create.data.groupCode"
              filterable
              clearable
              :options="groupOptions"
              placeholder="不选择则不发放用户组"
          ></n-select>
        </n-form-item>
        <n-form-item v-if="state.create.data.groupCode" path="groupDays" label="用户组有效天数(0=永久)">
          <n-input-number v-model:value="state.create.data.groupDays" :min="0" style="width: 100%"></n-input-number>
        </n-form-item>
        <n-form-item path="amount" label="发放积分(0=不发)">
          <n-input-number v-model:value="state.create.data.amount" :min="0" style="width: 100%"></n-input-number>
        </n-form-item>
        <n-form-item path="count" label="生成数量(1-1000)">
          <n-input-number v-model:value="state.create.data.count" :min="1" :max="1000" style="width: 100%"></n-input-number>
        </n-form-item>
        <n-form-item path="remark" label="备注">
          <n-input
              v-model:value="state.create.data.remark"
              type="textarea"
              placeholder="选填，最多500字"
              maxlength="500"
              :autosize="{minRows: 2, maxRows: 4}"
          ></n-input>
        </n-form-item>
      </n-form>
    </Modal>

    <Modal
        :show="state.card.open"
        :title="`卡密管理 - ${state.card.batchName}`"
        width="960px"
        confirm-text=""
        cancel-text="关闭"
        @on-cancel="closeCard"
        mask-close
    >
      <n-space vertical>
        <n-space>
          <n-select
              v-model:value="state.card.search.status"
              clearable
              :options="statusOptions"
              placeholder="状态"
              style="width: 120px"
          ></n-select>
          <n-input v-model:value="state.card.search.secret" clearable placeholder="卡密" style="width: 200px"></n-input>
          <n-input v-model:value="state.card.search.account" clearable placeholder="兑换账号" style="width: 150px"></n-input>
          <n-button type="info" :focusable="false" @click="cardSearchTable">搜索</n-button>
          <n-button type="info" :focusable="false" @click="cardRefreshTable">刷新</n-button>
          <n-button type="success" :focusable="false" @click="exportFunc(state.card.batchCode, state.card.batchName)">导出未使用卡密</n-button>
        </n-space>

        <n-data-table
            remote
            :columns="cardTableColumns"
            :data="state.card.list"
            :bordered="false"
            :loading="state.card.searchLoading"
            :scroll-x="state.card.minWidth"
        />

        <n-pagination
            :page-size="state.card.search.size"
            :page="state.card.search.page"
            :item-count="state.card.total"
            :simple="true"
            :on-update-page="(val) => {state.card.search.page = val;cardRefreshTable()}"
            :on-update-page-size="(val) => {state.card.search.size = val;cardRefreshTable()}"
        />
      </n-space>
    </Modal>
  </div>
</template>

<style scoped lang="scss">

</style>