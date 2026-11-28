<script setup>
import {computed, h, nextTick, onBeforeMount, onBeforeUnmount, ref, watch} from "vue";
import {useI18n} from "vue-i18n";
import * as echarts from "echarts";
import {apiAdminMetricRealtime, apiAdminMetricHistory, apiAdminMetricTop, apiAdminMetricObjects} from "../../../api/admin/metric.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";

const {t} = useI18n()

const kindOptions = computed(() => [
  {key: 1, value: t('metric.kindNode')},
  {key: 2, value: t('metric.kindClient')},
  {key: 3, value: t('metric.kindTunnel')},
  {key: 4, value: t('metric.kindUser')},
])
const metricKeys = ['connections', 'input_rate', 'output_rate', 'tunnel_count', 'client_count']
const metricOptions = computed(() => metricKeys.map(m => ({key: m, value: t('metric.m_' + m)})))
const metricText = (m) => t('metric.m_' + m)
const isRate = (m) => m === 'input_rate' || m === 'output_rate'

const fmtRate = (v) => {
  const n = Number(v) || 0
  if (n >= 1048576) return (n / 1048576).toFixed(2) + ' MB/s'
  if (n >= 1024) return (n / 1024).toFixed(2) + ' KB/s'
  return n.toFixed(0) + ' B/s'
}
const fmtValue = (m, v) => isRate(m) ? fmtRate(v) : (Number(v) || 0)

const state = ref({
  top: {originKind: 1, metric: 'input_rate', limit: 10, list: [], loading: false},
  rt: {
    originKind: 1, originCode: '', metrics: ['input_rate', 'output_rate', 'connections'],
    autoRefresh: true, gauge: {}, loading: false, objects: [],
  },
  hist: {
    originKind: 1, originCode: '', metric: 'input_rate',
    startTs: Date.now() - 3600 * 1000, endTs: Date.now(), loading: false, objects: [],
  },
})

const topMax = computed(() => {
  let max = 0
  for (const it of state.value.top.list) {
    const v = Number(it.value) || 0
    if (v > max) max = v
  }
  return max
})
const topColumns = computed(() => [
  {title: t('metric.rank'), key: 'rank', width: 70, render: (row, index) => index + 1},
  {
    title: t('common.name'), key: 'name', ellipsis: {tooltip: true}, minWidth: 180, render(row) {
      return h('div', {}, [
        h('div', {style: 'font-weight:500'}, row.name || '-'),
        h('div', {style: 'font-size:12px;color:#999'}, row.code),
      ])
    }
  },
  {
    title: t('metric.value'), key: 'value', width: 140, render(row) {
      return fmtValue(state.value.top.metric, row.value)
    }
  },
  {
    title: t('metric.percent'), key: 'bar', minWidth: 160, render(row) {
      const max = topMax.value
      const pct = max > 0 ? Math.round((Number(row.value) || 0) / max * 100) : 0
      return h('div', {style: 'width:100%'}, [
        h('div', {style: `height:8px;border-radius:4px;background:var(--primary-color,#18a058);width:${pct}%`}),
      ])
    }
  },
])

// ---- echarts ----
const rtChartEl = ref(null)
const histChartEl = ref(null)
let rtChart = null
let histChart = null
let timer = null

const renderLine = (chart, seriesList) => {
  if (!chart) return
  const series = seriesList.map(s => ({
    name: metricText(s.metric),
    type: 'line',
    showSymbol: false,
    smooth: true,
    data: (s.points || []).map(p => [p.ts * 1000, p.value]),
  }))
  chart.setOption({
    tooltip: {trigger: 'axis'},
    legend: {data: series.map(s => s.name)},
    grid: {left: 60, right: 20, top: 40, bottom: 40},
    xAxis: {type: 'time'},
    yAxis: {type: 'value'},
    series,
  }, true)
}

const topFunc = async () => {
  try {
    state.value.top.loading = true
    const res = await apiAdminMetricTop({
      originKind: state.value.top.originKind,
      metric: state.value.top.metric,
      limit: state.value.top.limit,
    })
    state.value.top.list = res.data || []
  } finally { state.value.top.loading = false }
}

const realtimeFunc = async () => {
  if (!state.value.rt.originCode) return
  try {
    state.value.rt.loading = true
    const res = await apiAdminMetricRealtime({
      originKind: state.value.rt.originKind,
      originCode: state.value.rt.originCode,
      metrics: state.value.rt.metrics,
    })
    const list = res.data || []
    const gauge = {}
    list.forEach(s => { gauge[s.metric] = s.gauge })
    state.value.rt.gauge = gauge
    renderLine(rtChart, list)
  } finally { state.value.rt.loading = false }
}

const historyFunc = async () => {
  if (!state.value.hist.originCode) return
  try {
    state.value.hist.loading = true
    const res = await apiAdminMetricHistory({
      originKind: state.value.hist.originKind,
      originCode: state.value.hist.originCode,
      metric: state.value.hist.metric,
      startTs: Math.floor(state.value.hist.startTs / 1000),
      endTs: Math.floor(state.value.hist.endTs / 1000),
    })
    const points = res.data || []
    renderLine(histChart, [{metric: state.value.hist.metric, points}])
  } finally { state.value.hist.loading = false }
}

// 加载某查询区(rt/hist)的可选对象，默认选中第一个，避免手填编码导致空曲线。
const loadObjects = async (which) => {
  const st = state.value[which]
  try {
    const res = await apiAdminMetricObjects({originKind: st.originKind})
    st.objects = res.data || []
  } catch (e) { st.objects = [] }
  if (!st.objects.find(o => o.code === st.originCode)) {
    st.originCode = st.objects.length ? st.objects[0].code : ''
  }
}
const onRtKind = async (v) => { state.value.rt.originKind = v; await loadObjects('rt'); realtimeFunc() }
const onHistKind = async (v) => { state.value.hist.originKind = v; await loadObjects('hist'); historyFunc() }

const setupTimer = () => {
  if (timer) { clearInterval(timer); timer = null }
  if (state.value.rt.autoRefresh) {
    timer = setInterval(realtimeFunc, 10000)
  }
}
watch(() => state.value.rt.autoRefresh, setupTimer)

const resizeCharts = () => {
  if (rtChart) rtChart.resize()
  if (histChart) histChart.resize()
}

onBeforeMount(async () => {
  await nextTick()
  if (rtChartEl.value) rtChart = echarts.init(rtChartEl.value)
  if (histChartEl.value) histChart = echarts.init(histChartEl.value)
  window.addEventListener('resize', resizeCharts)
  setupTimer()
  topFunc()
  await loadObjects('rt'); realtimeFunc()
  await loadObjects('hist'); historyFunc()
})
onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
  window.removeEventListener('resize', resizeCharts)
  if (rtChart) { rtChart.dispose(); rtChart = null }
  if (histChart) { histChart.dispose(); histChart = null }
})
</script>

<template>
  <div>
    <!-- 全站 TopN -->
    <SearchCard :show-border="false" space>
      <SearchItem type="select" :label-width="80" :label="t('metric.originKind')" :options="kindOptions"
                  :options-key-value="{key:'key',value:'value'}" :default="state.top.originKind"
                  @onChange="value => state.top.originKind = value"/>
      <SearchItem type="select" :label-width="80" :label="t('metric.metric')" :options="metricOptions"
                  :options-key-value="{key:'key',value:'value'}" :default="state.top.metric"
                  @onChange="value => state.top.metric = value"/>
      <SearchItem type="input-number" :label-width="80" :label="t('metric.limit')" :default="state.top.limit"
                  :input-number-min="1" :input-number-max="100"
                  @onChange="value => state.top.limit = value"/>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" :loading="state.top.loading" @click="topFunc">{{ t('metric.topN') }}</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>
    <AppCard :loading="state.top.loading">
      <n-data-table :columns="topColumns" :data="state.top.list" :bordered="false" :scroll-x="600">
        <template #empty>
          <n-empty :description="t('common.noData')"/>
        </template>
      </n-data-table>
    </AppCard>

    <!-- 实时曲线 -->
    <SearchCard :show-border="false" space>
      <SearchItem type="select" :label-width="80" :label="t('metric.originKind')" :options="kindOptions"
                  :options-key-value="{key:'key',value:'value'}" :default="state.rt.originKind"
                  @onChange="onRtKind"/>
      <SearchItem type="custom" :label-width="80" :label="t('metric.object')">
        <n-select v-model:value="state.rt.originCode" :options="state.rt.objects"
                  label-field="name" value-field="code" filterable clearable :placeholder="t('metric.selectObject')"
                  style="min-width:220px" @update:value="realtimeFunc"/>
      </SearchItem>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" :loading="state.rt.loading" @click="realtimeFunc">{{ t('metric.queryRealtime') }}</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>
    <AppCard :show-border="false">
      <n-space vertical>
        <n-space align="center">
          <span>{{ t('metric.metric') }}:</span>
          <n-select v-model:value="state.rt.metrics" multiple style="min-width:300px"
                    :options="metricOptions" :label-field="'value'" :value-field="'key'"/>
          <n-checkbox v-model:checked="state.rt.autoRefresh">{{ t('metric.autoRefreshInterval') }}</n-checkbox>
        </n-space>
        <n-space>
          <n-tag v-for="(v,k) in state.rt.gauge" :key="k" size="small" type="info">
            {{ metricText(k) }}: {{ fmtValue(k, v) }}
          </n-tag>
        </n-space>
      </n-space>
    </AppCard>
    <AppCard>
      <div ref="rtChartEl" style="width:100%;height:360px"></div>
    </AppCard>

    <!-- 历史曲线 -->
    <SearchCard :show-border="false" space>
      <SearchItem type="select" :label-width="80" :label="t('metric.originKind')" :options="kindOptions"
                  :options-key-value="{key:'key',value:'value'}" :default="state.hist.originKind"
                  @onChange="onHistKind"/>
      <SearchItem type="custom" :label-width="80" :label="t('metric.object')">
        <n-select v-model:value="state.hist.originCode" :options="state.hist.objects"
                  label-field="name" value-field="code" filterable clearable :placeholder="t('metric.selectObject')"
                  style="min-width:220px" @update:value="historyFunc"/>
      </SearchItem>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" :loading="state.hist.loading" @click="historyFunc">{{ t('metric.queryHistory') }}</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>
    <AppCard :show-border="false">
      <n-space align="center">
        <span>{{ t('metric.metric') }}:</span>
        <n-select v-model:value="state.hist.metric" style="min-width:160px"
                  :options="metricOptions" :label-field="'value'" :value-field="'key'"/>
        <n-date-picker v-model:value="state.hist.startTs" type="datetime" :placeholder="t('metric.startTs')"/>
        <n-date-picker v-model:value="state.hist.endTs" type="datetime" :placeholder="t('metric.endTs')"/>
      </n-space>
    </AppCard>
    <AppCard>
      <div ref="histChartEl" style="width:100%;height:360px"></div>
    </AppCard>
  </div>
</template>

<style scoped lang="scss">
</style>
