<script setup>
import {computed, h, onBeforeMount, ref} from "vue";
import {useI18n} from "vue-i18n";
import {NTag} from "naive-ui";
import {apiNormalTemplateCreate, apiNormalTemplateDelete, apiNormalTemplatePage, apiNormalTemplateApply} from "../../../api/normal/template.js";
import {apiNormalGostNodeList} from "../../../api/normal/gost_node.js";
import {apiNormalGostClientList} from "../../../api/normal/gost_client.js";
import {apiNormalGostNodeConfigList} from "../../../api/normal/gost_node_config.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";
import Modal from "../../../components/Modal.vue";
import {requiredRule} from "../../../utils/formDataRule.js";
import {generateButtonGroup, generateDeleteButton, generateTextButton} from "../../../utils/tableHelper.js";

const {t} = useI18n()

const kindOptions = computed(() => [
  {key: 'host', value: t('template.kindHost')},
  {key: 'forward', value: t('template.kindForward')},
  {key: 'proxy', value: t('template.kindProxy')},
  {key: 'tunnel', value: t('template.kindTunnel')},
  {key: 'p2p', value: t('template.kindP2P')},
])
const kindText = (k) => (kindOptions.value.find(o => o.key === k) || {}).value || k

const state = ref({
  table: {searchLoading: false, search: {page: 1, size: 10, kind: null}, list: [], total: 0, minWidth: 800},
  create: {open: false, loading: false, data: {name: '', kind: 'host', content: '{\n  \n}'}, dataRules: {name: requiredRule(t('common.name'))}},
  apply: {open: false, loading: false, data: {code: '', name: '', nodeCode: null, clientCode: null, configCode: null}},
  nodes: [], clients: [], configs: [],
})

const tableColumns = computed(() => [
  {title: t('common.name'), key: 'Name', ellipsis: {tooltip: true}},
  {title: t('template.kind'), key: 'Kind', width: 120, render: (row) => kindText(row.Kind)},
  {title: t('template.content'), key: 'Content', ellipsis: {tooltip: true}},
  {title: t('common.createdAt'), key: 'CreatedAt', width: 170, ellipsis: {tooltip: true}},
  {
    title: t('common.operate'), key: 'operator', width: 150, fixed: 'right', render(row) {
      return generateButtonGroup(
          generateTextButton(t('common.apply'), row, openApply, 'success'),
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
    const res = await apiNormalTemplatePage(state.value.table.search)
    state.value.table.list = res.data.list || []
    state.value.table.total = res.data.total || 0
  } finally { state.value.table.searchLoading = false }
}

const openCreate = () => { state.value.create.data = {name: '', kind: 'host', content: '{\n  \n}'}; state.value.create.open = true }
const closeCreate = () => { state.value.create.open = false }
const createRef = ref()
const createFunc = () => {
  createRef.value.validate(async valid => {
    if (!valid) {
      let content
      try { content = JSON.parse(state.value.create.data.content) } catch (e) { $message.error('JSON 格式错误'); return }
      try {
        state.value.create.loading = true
        await apiNormalTemplateCreate({name: state.value.create.data.name, kind: state.value.create.data.kind, content})
        searchTable(); closeCreate()
      } finally { state.value.create.loading = false }
    }
  })
}
const deleteFunc = async (row) => { await apiNormalTemplateDelete({code: row.Code}); searchTable() }

const loadNodes = async () => { const res = await apiNormalGostNodeList({}); state.value.nodes = res.data || [] }
const loadClients = async () => { const res = await apiNormalGostClientList({}); state.value.clients = res.data || [] }
const loadConfigs = async (nodeCode) => {
  if (!nodeCode) { state.value.configs = []; return }
  const res = await apiNormalGostNodeConfigList({nodeCode})
  state.value.configs = res.data || []
}
const openApply = (row) => {
  state.value.apply.data = {code: row.Code, name: row.Name, nodeCode: null, clientCode: null, configCode: null}
  state.value.configs = []
  state.value.apply.open = true
}
const closeApply = () => { state.value.apply.open = false }
const onApplyNodeChange = (val) => {
  state.value.apply.data.nodeCode = val
  state.value.apply.data.configCode = null
  loadConfigs(val)
}
const applyFunc = async () => {
  const d = state.value.apply.data
  if (!d.nodeCode || !d.clientCode || !d.configCode) { $message.error(t('common.noData')); return }
  try {
    state.value.apply.loading = true
    await apiNormalTemplateApply({code: d.code, name: d.name, nodeCode: d.nodeCode, clientCode: d.clientCode, configCode: d.configCode})
    $message.success(t('common.apply'))
    closeApply()
  } finally { state.value.apply.loading = false }
}

onBeforeMount(() => { pageFunc(); loadNodes(); loadClients() })
</script>

<template>
  <div>
    <SearchCard :show-border="false" space>
      <SearchItem type="select" :label-width="80" :label="t('template.kind')" :options="kindOptions"
                  :options-key-value="{key:'key',value:'value'}" clearable
                  @onChange="value => state.table.search.kind = value"/>
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

    <Modal :show="state.create.open" :title="t('common.create')" width="520px" :confirm-text="t('common.save')"
           :cancel-text="t('common.cancel')" :confirm-loading="state.create.loading"
           @on-confirm="createFunc" @on-cancel="closeCreate" mask-close>
      <n-form ref="createRef" :rules="state.create.dataRules" :model="state.create.data">
        <n-form-item path="name" :label="t('common.name')">
          <n-input v-model:value.trim="state.create.data.name"/>
        </n-form-item>
        <n-form-item path="kind" :label="t('template.kind')">
          <n-select v-model:value="state.create.data.kind" :options="kindOptions"
                    :label-field="'value'" :value-field="'key'"/>
        </n-form-item>
        <n-form-item path="content" :label="t('template.content')">
          <n-input type="textarea" :autosize="{minRows:6,maxRows:20}" v-model:value="state.create.data.content"/>
        </n-form-item>
        <n-text depth="3">{{ t('template.contentHint') }}</n-text>
      </n-form>
    </Modal>

    <Modal :show="state.apply.open" :title="t('template.applyTitle')" width="440px" :confirm-text="t('common.apply')"
           :cancel-text="t('common.cancel')" :confirm-loading="state.apply.loading"
           @on-confirm="applyFunc" @on-cancel="closeApply" mask-close>
      <n-form :model="state.apply.data">
        <n-form-item :label="t('template.applyName')">
          <n-input v-model:value.trim="state.apply.data.name"/>
        </n-form-item>
        <n-form-item :label="t('template.nodeCode')">
          <n-select :value="state.apply.data.nodeCode" :options="state.nodes"
                    :label-field="'name'" :value-field="'code'" filterable
                    :on-update:value="onApplyNodeChange"/>
        </n-form-item>
        <n-form-item :label="t('template.clientCode')">
          <n-select v-model:value="state.apply.data.clientCode" :options="state.clients"
                    :label-field="'name'" :value-field="'code'" filterable/>
        </n-form-item>
        <n-form-item :label="t('template.configCode')">
          <n-select v-model:value="state.apply.data.configCode" :options="state.configs"
                    :label-field="'name'" :value-field="'code'" filterable/>
        </n-form-item>
      </n-form>
    </Modal>
  </div>
</template>

<style scoped lang="scss">
</style>
