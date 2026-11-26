<script setup>
import {onBeforeMount, ref} from "vue";
import {
  apiAdminSystemUserCreate,
  apiAdminSystemUserDelete,
  apiAdminSystemUserPage,
  apiAdminSystemUserUpdate,
  apiAdminSystemUserBan
} from "../../../api/admin/system_user.js";
import {apiAdminSystemUserGroupList} from "../../../api/admin/system_user_group.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";
import {generateTableColumns} from "./tableColumns.js";
import Modal from "../../../components/Modal.vue";
import {requiredRule} from "../../../utils/formDataRule.js";

const state = ref({
  table: {
    searchLoading: false,
    search: {
      page: 1,
      size: 10,
      account: '',
      admin: 0,
      status: 0,
    },
    list: [],
    total: 0,
    minWidth: 1480,
  },
  create: {
    open: false,
    loading: false,
    data: {
      account: '',
      password: '',
      amount: '0',
      admin: 2,
      remark: '',
      groupCode: '',
    },
    dataRules: {
      account: requiredRule('账号不能为空'),
      password: requiredRule('密码不能为空'),
    },
  },
  update: {
    open: false,
    loading: false,
    data: {
      code: '',
      account: '',
      password: '',
      amount: '0',
      admin: 2,
      remark: '',
      groupCode: '',
    },
    dataRules: {
      account: requiredRule('账号不能为空'),
    },
  },
  ban: {
    open: false,
    loading: false,
    data: {
      code: '',
      account: '',
      reason: '',
    },
  }
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
    let res = await apiAdminSystemUserPage(state.value.table.search)
    state.value.table.list = res.data.list
    state.value.table.total = res.data.total
  } finally {
    state.value.table.searchLoading = false
  }
}

const openCreate = () => {
  state.value.create.data = {
    account: '',
    password: '',
    amount: '0',
    admin: 2,
    remark: '',
    groupCode: '',
  }
  state.value.create.open = true
}

const closeCreate = () => {
  state.value.create.open = false
}

const openUpdate = (row) => {
  state.value.update.open = true
  state.value.update.data = {
    code: row.code,
    account: row.account,
    password: '',
    amount: row.amount,
    admin: row.admin,
    remark: row.remark || '',
    groupCode: row.groupCode || '',
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
        await apiAdminSystemUserCreate(state.value.create.data)
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
        await apiAdminSystemUserUpdate(state.value.update.data)
        searchTable()
        closeUpdate()
      } finally {
        state.value.update.loading = false
      }
    }
  })
}

const deleteFunc = async (row) => {
  try {
    await apiAdminSystemUserDelete({code: row.code})
    searchTable()
  } finally {
  }
}

const openBan = (row) => {
  state.value.ban.data = {
    code: row.code,
    account: row.account,
    reason: '',
  }
  state.value.ban.open = true
}

const closeBan = () => {
  state.value.ban.open = false
}

const banFunc = async () => {
  try {
    state.value.ban.loading = true
    await apiAdminSystemUserBan({
      code: state.value.ban.data.code,
      status: 2,
      reason: state.value.ban.data.reason,
    })
    searchTable()
    closeBan()
  } finally {
    state.value.ban.loading = false
  }
}

const unbanFunc = (row) => {
  $dialog.warning({
    maskClosable: true,
    closable: false,
    showIcon: true,
    title: '解封',
    content: `确认解封用户「${row.account}」吗?`,
    negativeText: '取消',
    positiveText: '确认',
    negativeButtonProps: {focusable: false},
    positiveButtonProps: {type: 'success', focusable: false},
    async onPositiveClick() {
      await apiAdminSystemUserBan({code: row.code, status: 1, reason: ''})
      searchTable()
    }
  })
}

const tableColumns = generateTableColumns(openUpdate, deleteFunc, openBan, unbanFunc)

const groupOptions = ref([{label: '无组(不限制)', value: ''}])
const loadGroups = async () => {
  const res = await apiAdminSystemUserGroupList({})
  groupOptions.value = [{label: '无组(不限制)', value: ''}].concat(
      (res.data || []).map(g => ({label: g.name, value: g.code}))
  )
}

onBeforeMount(() => {
  pageFunc()
  loadGroups()
})

</script>

<template>
  <div>
    <SearchCard :show-border="false" space>
      <SearchItem
          type="input"
          :label-width="70"
          clearable
          label="账号"
          @onChange="value => state.table.search.account=value"
      ></SearchItem>
      <SearchItem
          type="select"
          :label-width="70"
          label="角色"
          :default="0"
          :options="[
              {value:'全部',key:0},
              {value:'管理员',key:1},
              {value:'普通用户',key:2}
          ]"
          @onChange="value => state.table.search.admin=value"
      ></SearchItem>
      <SearchItem
          type="select"
          :label-width="70"
          label="状态"
          :default="0"
          :options="[
              {value:'全部',key:0},
              {value:'正常',key:1},
              {value:'已封禁',key:2}
          ]"
          @onChange="value => state.table.search.status=value"
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
        title="新增用户"
        width="400px"
        confirm-text="保存"
        cancel-text="取消"
        :confirm-loading="state.create.loading"
        @on-confirm="createFunc"
        @on-cancel="closeCreate"
        mask-close
    >
      <n-form ref="createRef" :rules="state.create.dataRules" :model="state.create.data">
        <n-form-item path="account" label="账号">
          <n-input v-model:value.trim="state.create.data.account"></n-input>
        </n-form-item>
        <n-form-item path="password" label="密码">
          <n-input v-model:value.trim="state.create.data.password"></n-input>
        </n-form-item>
        <n-form-item path="admin" label="角色(重新登录生效)">
          <n-select
              v-model:value="state.create.data.admin"
              :default-value="state.create.data.admin"
              :options="[{label:'管理员',value:1},{label:'普通用户',value:2}]"
          ></n-select>
        </n-form-item>
        <n-form-item path="amount" label="积分">
          <n-input v-model:value.trim="state.create.data.amount"></n-input>
        </n-form-item>
        <n-form-item path="groupCode" label="用户组">
          <n-select
              v-model:value="state.create.data.groupCode"
              :options="groupOptions"
          ></n-select>
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
        title="修改用户"
        width="400px"
        confirm-text="保存"
        cancel-text="取消"
        :confirm-loading="state.update.loading"
        @on-confirm="updateFunc"
        @on-cancel="closeUpdate"
        mask-close
    >
      <n-form ref="updateRef" :rules="state.update.dataRules" :model="state.update.data">
        <n-form-item path="account" label="账号">
          <n-input v-model:value.trim="state.update.data.account"></n-input>
        </n-form-item>
        <n-form-item path="password" label="密码">
          <n-input v-model:value.trim="state.update.data.password"></n-input>
        </n-form-item>
        <n-form-item path="admin" label="角色(重新登录生效)">
          <n-select
              v-model:value="state.update.data.admin"
              :default-value="state.update.data.admin"
              :options="[{label:'管理员',value:1},{label:'普通用户',value:2}]"
          ></n-select>
        </n-form-item>
        <n-form-item path="amount" label="积分">
          <n-input v-model:value.trim="state.update.data.amount"></n-input>
        </n-form-item>
        <n-form-item path="groupCode" label="用户组">
          <n-select
              v-model:value="state.update.data.groupCode"
              :options="groupOptions"
          ></n-select>
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

    <Modal
        :show="state.ban.open"
        title="封禁用户"
        width="400px"
        confirm-text="确认封禁"
        cancel-text="取消"
        :confirm-loading="state.ban.loading"
        @on-confirm="banFunc"
        @on-cancel="closeBan"
        mask-close
    >
      <n-form :model="state.ban.data">
        <n-form-item label="账号">
          <n-input :value="state.ban.data.account" disabled></n-input>
        </n-form-item>
        <n-form-item label="封禁原因">
          <n-input
              v-model:value="state.ban.data.reason"
              type="textarea"
              placeholder="选填，最多500字，将展示给被封禁用户"
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