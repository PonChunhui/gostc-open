<script setup>
import {computed, h, nextTick, onBeforeMount, onBeforeUnmount, ref, watch} from "vue";
import * as echarts from "echarts";
import {apiAdminMetricRealtime, apiAdminMetricHistory, apiAdminMetricTop} from "../../../api/admin/metric.js";
import AppCard from "../../../layout/components/AppCard.vue";
import SearchCard from "../../../layout/components/SearchCard.vue";
import SearchItem from "../../../layout/components/SearchItem.vue";

const kindOptions = [
  {key: 1, value: '节点'},
  {key: 2, value: '客户端'},
  {key: 3, value: '隧道'},
  {key: 4, value: '用户'},
]
const metricKeys = ['connections', 'input_rate', 'output_rate', 'tunnel_count', 'client_count']
const metricTextMap = {
  connections: '连接数', input_rate: '上行速率', output_rate: '下行速率',
  tunnel_count: '隧道数', client_count: '客户端数',
}
const metricOptions = metricKeys.map(m => ({key: m, value: metricTextMap[m]}))
const metricText = (m) => metricTextMap[m] || m
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
    autoRefresh: true, gauge: {}, loading: false,
  },
  hist: {
    originKind: 1, originCode: '', metric: 'input_rate',
    startTs: Date.now() - 3600 * 1000, endTs: Date.now(), loading: false,
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
  {title: '排名', key: 'rank', width: 70, render: (row, index) => index + 1},
  {
    title: '名称', key: 'name', ellipsis: {tooltip: true}, minWidth: 180, render(row) {
      return h('div', {}, [
        h('div', {style: 'font-weight:500'}, row.name || '-'),
        h('div', {style: 'font-size:12px;color:#999'}, row.code),
      ])
    }
  },
  {
    title: '数值', key: 'value', width: 140, render(row) {
      return fmtValue(state.value.top.metric, row.value)
    }
  },
  {
    title: '占比', key: 'bar', minWidth: 160, render(row) {
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
      <SearchItem type="select" :label-width="80" label="对象类型" :options="kindOptions"
                  :options-key-value="{key:'key',value:'value'}" :default="state.top.originKind"
                  @onChange="value => state.top.originKind = value"/>
      <SearchItem type="select" :label-width="80" label="指标" :options="metricOptions"
                  :options-key-value="{key:'key',value:'value'}" :default="state.top.metric"
                  @onChange="value => state.top.metric = value"/>
      <SearchItem type="input-number" :label-width="80" label="数量" :default="state.top.limit"
                  :input-number-min="1" :input-number-max="100"
                  @onChange="value => state.top.limit = value"/>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" :loading="state.top.loading" @click="topFunc">查询 TopN</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>
    <AppCard :loading="state.top.loading">
      <n-data-table :columns="topColumns" :data="state.top.list" :bordered="false" :scroll-x="600"/>
    </AppCard>

    <!-- 实时曲线 -->
    <SearchCard :show-border="false" space>
      <SearchItem type="select" :label-width="80" label="对象类型" :options="kindOptions"
                  :options-key-value="{key:'key',value:'value'}" :default="state.rt.originKind"
                  @onChange="value => state.rt.originKind = value"/>
      <SearchItem type="input" :label-width="80" label="对象编码" clearable
                  @onChange="value => state.rt.originCode = value"/>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" :loading="state.rt.loading" @click="realtimeFunc">查询实时</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>
    <AppCard :show-border="false">
      <n-space vertical>
        <n-space align="center">
          <span>指标：</span>
          <n-select v-model:value="state.rt.metrics" multiple style="min-width:300px"
                    :options="metricOptions" :label-field="'value'" :value-field="'key'"/>
          <n-checkbox v-model:checked="state.rt.autoRefresh">自动刷新(10s)</n-checkbox>
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
      <SearchItem type="select" :label-width="80" label="对象类型" :options="kindOptions"
                  :options-key-value="{key:'key',value:'value'}" :default="state.hist.originKind"
                  @onChange="value => state.hist.originKind = value"/>
      <SearchItem type="input" :label-width="80" label="对象编码" clearable
                  @onChange="value => state.hist.originCode = value"/>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" :loading="state.hist.loading" @click="historyFunc">查询历史</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>
    <AppCard :show-border="false">
      <n-space align="center">
        <span>指标：</span>
        <n-select v-model:value="state.hist.metric" style="min-width:160px"
                  :options="metricOptions" :label-field="'value'" :value-field="'key'"/>
        <n-date-picker v-model:value="state.hist.startTs" type="datetime" placeholder="开始时间"/>
        <n-date-picker v-model:value="state.hist.endTs" type="datetime" placeholder="结束时间"/>
      </n-space>
    </AppCard>
    <AppCard>
      <div ref="histChartEl" style="width:100%;height:360px"></div>
    </AppCard>
  </div>
</template>

<style scoped lang="scss">
</style>
