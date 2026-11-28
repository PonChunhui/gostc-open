<script setup>
import {computed, h, onBeforeMount, ref} from "vue";
import {useI18n} from "vue-i18n";
import {NTag} from "naive-ui";
import {
  apiAdminNodeGroupCreate, apiAdminNodeGroupUpdate, apiAdminNodeGroupDelete, apiAdminNodeGroupPage
} from "../../../api/admin/gost_node_group.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";
import Modal from "../../../components/Modal.vue";
import {requiredRule} from "../../../utils/formDataRule.js";
import {generateButtonGroup, generateUpdateButton, generateDeleteButton} from "../../../utils/tableHelper.js";

const {t} = useI18n()

const strategyOptions = computed(() => [
  {key: 'priority', value: t('nodeGroup.strategyPriority')},
  {key: 'weight', value: t('nodeGroup.strategyWeight')},
  {key: 'random', value: t('nodeGroup.strategyRandom')},
])
const strategyText = (s) => (strategyOptions.value.find(o => o.key === s) || {}).value || s
const emptyData = () => ({code: '', name: '', strategy: 'priority', remark: ''})

const state = ref({
  table: {searchLoading: false, search: {page: 1, size: 10, name: null}, list: [], total: 0, minWidth: 700},
  create: {open: false, loading: false, data: emptyData(), dataRules: {name: requiredRule(t('common.name'))}},
  update: {open: false, loading: false, data: emptyData(), dataRules: {name: requiredRule(t('common.name'))}},
})

const tableColumns = computed(() => [
  {title: t('common.name'), key: 'name', ellipsis: {tooltip: true}},
  {
    title: t('nodeGroup.strategy'), key: 'strategy', width: 120, render(row) {
      return h(NTag, {size: 'small', type: 'info'}, {default: () => strategyText(row.strategy)})
    }
  },
  {title: t('nodeGroup.nodeCount'), key: 'nodeCount', width: 90},
  {title: t('common.remark'), key: 'remark', ellipsis: {tooltip: true}},
  {title: t('common.createdAt'), key: 'createdAt', width: 170, ellipsis: {tooltip: true}},
  {
    title: t('common.operate'), key: 'operator', width: 130, fixed: 'right', render(row) {
      return generateButtonGroup(
          generateUpdateButton(row, openUpdate, t),
          generateDeleteButton(row, deleteFunc, t),
      )
    }
  },
])

const refreshTable = () => pageFunc()
const searchTable = () => { state.value.table.search.page = 1; pageFunc() }
const pageFunc = async () => {
  try {
    state.value.table.searchLoading = true
    const res = await apiAdminNodeGroupPage(state.value.table.search)
    state.value.table.list = res.data.list || []
    state.value.table.total = res.data.total || 0
  } finally { state.value.table.searchLoading = false }
}

const openCreate = () => { state.value.create.data = emptyData(); state.value.create.open = true }
const closeCreate = () => { state.value.create.open = false }
const createRef = ref()
const createFunc = () => {
  createRef.value.validate(async valid => {
    if (!valid) {
      try {
        state.value.create.loading = true
        await apiAdminNodeGroupCreate(state.value.create.data)
        searchTable(); closeCreate()
      } finally { state.value.create.loading = false }
    }
  })
}
const openUpdate = (row) => {
  state.value.update.data = {code: row.code, name: row.name, strategy: row.strategy, remark: row.remark}
  state.value.update.open = true
}
const closeUpdate = () => { state.value.update.open = false }
const updateRef = ref()
const updateFunc = () => {
  updateRef.value.validate(async valid => {
    if (!valid) {
      try {
        state.value.update.loading = true
        await apiAdminNodeGroupUpdate(state.value.update.data)
        refreshTable(); closeUpdate()
      } finally { state.value.update.loading = false }
    }
  })
}
const deleteFunc = async (row) => { await apiAdminNodeGroupDelete({code: row.code}); searchTable() }

onBeforeMount(() => { pageFunc() })
</script>

<template>
  <div>
    <SearchCard :show-border="false" space>
      <SearchItem type="input" :label-width="80" :label="t('common.name')" clearable
                  @onChange="value => state.table.search.name = value"/>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" @click="searchTable">{{ t('common.search') }}</n-button>
          <n-button type="info" :focusable="false" @click="refreshTable">{{ t('common.refresh') }}</n-button>
          <n-button type="info" :focusable="false" @click="openCreate">{{ t('common.create') }}</n-button>
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

    <Modal :show="state.create.open" :title="t('common.create')" width="400px" :confirm-text="t('common.save')"
           :cancel-text="t('common.cancel')" :confirm-loading="state.create.loading"
           @on-confirm="createFunc" @on-cancel="closeCreate" mask-close>
      <n-form ref="createRef" :rules="state.create.dataRules" :model="state.create.data">
        <n-form-item path="name" :label="t('common.name')"><n-input v-model:value.trim="state.create.data.name"/></n-form-item>
        <n-form-item path="strategy" :label="t('nodeGroup.strategy')"><n-select v-model:value="state.create.data.strategy" :options="strategyOptions" :label-field="'value'" :value-field="'key'"/></n-form-item>
        <n-form-item path="remark" :label="t('common.remark')"><n-input type="textarea" :autosize="{minRows:2,maxRows:4}" v-model:value="state.create.data.remark"/></n-form-item>
      </n-form>
    </Modal>

    <Modal :show="state.update.open" :title="t('common.update')" width="400px" :confirm-text="t('common.save')"
           :cancel-text="t('common.cancel')" :confirm-loading="state.update.loading"
           @on-confirm="updateFunc" @on-cancel="closeUpdate" mask-close>
      <n-form ref="updateRef" :rules="state.update.dataRules" :model="state.update.data">
        <n-form-item path="name" :label="t('common.name')"><n-input v-model:value.trim="state.update.data.name"/></n-form-item>
        <n-form-item path="strategy" :label="t('nodeGroup.strategy')"><n-select v-model:value="state.update.data.strategy" :options="strategyOptions" :label-field="'value'" :value-field="'key'"/></n-form-item>
        <n-form-item path="remark" :label="t('common.remark')"><n-input type="textarea" :autosize="{minRows:2,maxRows:4}" v-model:value="state.update.data.remark"/></n-form-item>
      </n-form>
    </Modal>
  </div>
</template>

<style scoped lang="scss">
</style>
