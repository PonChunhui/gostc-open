<script setup>
import {computed, nextTick, onBeforeMount, onBeforeUnmount, ref, watch} from "vue";
import {useI18n} from "vue-i18n";
import * as echarts from "echarts";
import {apiNormalMetricRealtime, apiNormalMetricHistory} from "../../../api/normal/metric.js";
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

const state = ref({
  form: {
    originKind: 1, originCode: '', mode: 'realtime',
    metrics: ['connections'], metric: 'connections',
    startTs: Date.now() - 3600 * 1000, endTs: Date.now(),
    autoRefresh: true,
  },
  loading: false,
  gauge: {},
})

const chartEl = ref(null)
let chart = null
let timer = null

const fmtTime = (ts) => {
  const d = new Date(ts * 1000)
  const p = (n) => (n < 10 ? '0' + n : n)
  return `${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

const renderChart = (seriesList) => {
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
    grid: {left: 50, right: 20, top: 40, bottom: 40},
    xAxis: {type: 'time'},
    yAxis: {type: 'value'},
    series,
  }, true)
}

const realtimeFunc = async () => {
  if (!state.value.form.originCode) return
  try {
    state.value.loading = true
    const res = await apiNormalMetricRealtime({
      originKind: state.value.form.originKind,
      originCode: state.value.form.originCode,
      metrics: state.value.form.metrics,
    })
    const list = res.data || []
    const gauge = {}
    list.forEach(s => { gauge[s.metric] = s.gauge })
    state.value.gauge = gauge
    renderChart(list)
  } finally { state.value.loading = false }
}

const historyFunc = async () => {
  if (!state.value.form.originCode) return
  try {
    state.value.loading = true
    const res = await apiNormalMetricHistory({
      originKind: state.value.form.originKind,
      originCode: state.value.form.originCode,
      metric: state.value.form.metric,
      startTs: Math.floor(state.value.form.startTs / 1000),
      endTs: Math.floor(state.value.form.endTs / 1000),
    })
    const points = res.data || []
    renderChart([{metric: state.value.form.metric, points}])
  } finally { state.value.loading = false }
}

const queryFunc = () => {
  if (state.value.form.mode === 'realtime') realtimeFunc()
  else historyFunc()
}

const setupTimer = () => {
  if (timer) { clearInterval(timer); timer = null }
  if (state.value.form.mode === 'realtime' && state.value.form.autoRefresh) {
    timer = setInterval(realtimeFunc, 5000)
  }
}

watch(() => [state.value.form.mode, state.value.form.autoRefresh], setupTimer)

onBeforeMount(async () => {
  await nextTick()
  if (chartEl.value) {
    chart = echarts.init(chartEl.value)
    window.addEventListener('resize', resizeChart)
  }
  setupTimer()
})
const resizeChart = () => { if (chart) chart.resize() }
onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
  window.removeEventListener('resize', resizeChart)
  if (chart) { chart.dispose(); chart = null }
})
</script>

<template>
  <div>
    <SearchCard :show-border="false" space>
      <SearchItem type="select" :label-width="80" :label="t('metric.originKind')" :options="kindOptions"
                  :options-key-value="{key:'key',value:'value'}" :default="state.form.originKind"
                  @onChange="value => state.form.originKind = value"/>
      <SearchItem type="input" :label-width="80" :label="t('metric.originCode')" clearable
                  @onChange="value => state.form.originCode = value"/>
      <SearchItem type="radio-btn" :label-width="60" :label="t('metric.realtime')"
                  :options="[{key:'realtime',value:t('metric.realtime')},{key:'history',value:t('metric.history')}]"
                  :options-key-value="{key:'key',value:'value'}" :default="state.form.mode"
                  @onChange="value => state.form.mode = value"/>
      <SearchItem type="custom">
        <n-space>
          <n-button type="info" :focusable="false" :loading="state.loading" @click="queryFunc">{{ t('metric.query') }}</n-button>
        </n-space>
      </SearchItem>
    </SearchCard>

    <AppCard :show-border="false">
      <n-space vertical>
        <n-space v-if="state.form.mode==='realtime'" align="center">
          <span>{{ t('metric.metric') }}：</span>
          <n-select v-model:value="state.form.metrics" multiple style="min-width:280px"
                    :options="metricOptions" :label-field="'value'" :value-field="'key'"/>
          <n-checkbox v-model:checked="state.form.autoRefresh">{{ t('metric.autoRefresh') }}</n-checkbox>
        </n-space>
        <n-space v-else align="center">
          <span>{{ t('metric.metric') }}：</span>
          <n-select v-model:value="state.form.metric" style="min-width:160px"
                    :options="metricOptions" :label-field="'value'" :value-field="'key'"/>
          <n-date-picker v-model:value="state.form.startTs" type="datetime" :placeholder="t('metric.startTs')"/>
          <n-date-picker v-model:value="state.form.endTs" type="datetime" :placeholder="t('metric.endTs')"/>
        </n-space>
        <n-space v-if="state.form.mode==='realtime'">
          <n-tag v-for="(v,k) in state.gauge" :key="k" size="small" type="info">
            {{ metricText(k) }}: {{ v }}
          </n-tag>
        </n-space>
      </n-space>
    </AppCard>

    <AppCard>
      <div ref="chartEl" style="width:100%;height:420px"></div>
    </AppCard>
  </div>
</template>

<style scoped lang="scss">
</style>
