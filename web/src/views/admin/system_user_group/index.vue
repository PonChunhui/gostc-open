<script setup>
import {onBeforeMount, ref} from "vue";
import {
  apiAdminSystemUserGroupCreate,
  apiAdminSystemUserGroupDelete,
  apiAdminSystemUserGroupPage,
  apiAdminSystemUserGroupUpdate
} from "../../../api/admin/system_user_group.js";
import {apiAdminGostNodeList} from "../../../api/admin/gost_node.js";
import {apiAdminGostNodeConfigListAll} from "../../../api/admin/gost_node_config.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";
import {generateTableColumns} from "./tableColumns.js";
import Modal from "../../../components/Modal.vue";
import {requiredRule} from "../../../utils/formDataRule.js";

const emptyData = () => ({
  indexValue: 1000,
  name: '',
  remark: '',
  nodeCodes: [],
  configCodes: [],
  tunnelLimit: 0,
  monthTraffic: 0,
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
    minWidth: 1200,
  },
  create: {
    open: false,
    loading: false,
    data: emptyData(),
    dataRules: {
      name: requiredRule('名称不能为空'),
    },
  },
  update: {
    open: false,
    loading: false,
    data: {code: '', ...emptyData()},
    dataRules: {
      name: requiredRule('名称不能为空'),
    },
  }
})

const nodeOptions = ref([])
const configOptions = ref([])
const loadOptions = async () => {
  const nodeRes = await apiAdminGostNodeList({})
  nodeOptions.value = (nodeRes.data || []).map(n => ({label: n.name, value: n.code}))
  const configRes = await apiAdminGostNodeConfigListAll({})
  configOptions.value = (configRes.data || []).map(c => ({
    label: (c.nodeName ? c.nodeName + ' / ' : '') + c.name,
    value: c.code
  }))
}

const buildPayload = (data) => ({
  ...data,
  nodeCodes: (data.nodeCodes || []).join(','),
  configCodes: (data.configCodes || []).join(','),
})

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
    let res = await apiAdminSystemUserGroupPage(state.value.table.search)
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

const openUpdate = (row) => {
  state.value.update.open = true
  state.value.update.data = {
    code: row.code,
    indexValue: row.indexValue,
    name: row.name,
    remark: row.remark || '',
    nodeCodes: row.nodeCodes ? row.nodeCodes.split(',').filter(v => v) : [],
    configCodes: row.configCodes ? row.configCodes.split(',').filter(v => v) : [],
    tunnelLimit: row.tunnelLimit,
    monthTraffic: row.monthTraffic,
  }
}

const closeUpdate = () => {
  state.value.update.open = false
}

const createRef = ref()
const createFunc = () => {
  createRef.value.validate(async valid => {
    if (!valid) {
      try {
        state.value.create.loading = true
        await apiAdminSystemUserGroupCreate(buildPayload(state.value.create.data))
        searchTable()
        closeCreate()
      } finally {
        state.value.create.loading = false
      }
    }
  })
}

const updateRef = ref()
const updateFunc = () => {
  updateRef.value.validate(async valid => {
    if (!valid) {
      try {
        state.value.update.loading = true
        await apiAdminSystemUserGroupUpdate(buildPayload(state.value.update.data))
        searchTable()
        closeUpdate()
      } finally {
        state.value.update.loading = false
      }
    }
  })
}

const deleteFunc = async (row) => {
  await apiAdminSystemUserGroupDelete({code: row.code})
  searchTable()
}

const tableColumns = generateTableColumns(openUpdate, deleteFunc)

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
          label="名称"
          @onChange="value => state.table.search.name=value"
      ></SearchItem>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" @click="searchTable">搜索</n-button>
          <n-button type="info" :focusable="false" @click="refreshTable">刷新</n-button>
          <n-button type="info" :focusable="false" @click="openCreate">新增</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>

    <AppCard>
      <n-data-table
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

    <Modal
        :show="state.create.open"
        title="新增用户组"
        width="560px"
        confirm-text="保存"
        cancel-text="取消"
        :confirm-loading="state.create.loading"
        @on-confirm="createFunc"
        @on-cancel="closeCreate"
        mask-close
    >
      <n-form ref="createRef" :rules="state.create.dataRules" :model="state.create.data">
        <n-form-item path="name" label="名称">
          <n-input v-model:value.trim="state.create.data.name"></n-input>
        </n-form-item>
        <n-form-item path="indexValue" label="排序(越小越靠前)">
          <n-input-number v-model:value="state.create.data.indexValue" :min="0" style="width: 100%"></n-input-number>
        </n-form-item>
        <n-form-item path="nodeCodes" label="节点白名单(留空=全部)">
          <n-select
              v-model:value="state.create.data.nodeCodes"
              multiple
              filterable
              clearable
              :options="nodeOptions"
              placeholder="不选择则允许使用全部节点"
          ></n-select>
        </n-form-item>
        <n-form-item path="configCodes" label="套餐白名单(留空=全部)">
          <n-select
              v-model:value="state.create.data.configCodes"
              multiple
              filterable
              clearable
              :options="configOptions"
              placeholder="不选择则允许使用全部套餐"
          ></n-select>
        </n-form-item>
        <n-form-item path="tunnelLimit" label="隧道数量上限(0=不限)">
          <n-input-number v-model:value="state.create.data.tunnelLimit" :min="0" style="width: 100%"></n-input-number>
        </n-form-item>
        <n-form-item path="monthTraffic" label="每月流量上限GB(0=不限,近30天)">
          <n-input-number v-model:value="state.create.data.monthTraffic" :min="0" style="width: 100%"></n-input-number>
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
        :show="state.update.open"
        title="修改用户组"
        width="560px"
        confirm-text="保存"
        cancel-text="取消"
        :confirm-loading="state.update.loading"
        @on-confirm="updateFunc"
        @on-cancel="closeUpdate"
        mask-close
    >
      <n-form ref="updateRef" :rules="state.update.dataRules" :model="state.update.data">
        <n-form-item path="name" label="名称">
          <n-input v-model:value.trim="state.update.data.name"></n-input>
        </n-form-item>
        <n-form-item path="indexValue" label="排序(越小越靠前)">
          <n-input-number v-model:value="state.update.data.indexValue" :min="0" style="width: 100%"></n-input-number>
        </n-form-item>
        <n-form-item path="nodeCodes" label="节点白名单(留空=全部)">
          <n-select
              v-model:value="state.update.data.nodeCodes"
              multiple
              filterable
              clearable
              :options="nodeOptions"
              placeholder="不选择则允许使用全部节点"
          ></n-select>
        </n-form-item>
        <n-form-item path="configCodes" label="套餐白名单(留空=全部)">
          <n-select
              v-model:value="state.update.data.configCodes"
              multiple
              filterable
              clearable
              :options="configOptions"
              placeholder="不选择则允许使用全部套餐"
          ></n-select>
        </n-form-item>
        <n-form-item path="tunnelLimit" label="隧道数量上限(0=不限)">
          <n-input-number v-model:value="state.update.data.tunnelLimit" :min="0" style="width: 100%"></n-input-number>
        </n-form-item>
        <n-form-item path="monthTraffic" label="每月流量上限GB(0=不限,近30天)">
          <n-input-number v-model:value="state.update.data.monthTraffic" :min="0" style="width: 100%"></n-input-number>
        </n-form-item>
        <n-form-item path="remark" label="备注">
          <n-input
              v-model:value="state.update.data.remark"
              type="textarea"
              placeholder="选填，最多500字"
              maxlength="500"
              :autosize="{minRows: 2, maxRows: 4}"
          ></n-input>
        </n-form-item>
      </n-form>
    </Modal>
  </div>
</template>

<style scoped lang="scss">

</style>
