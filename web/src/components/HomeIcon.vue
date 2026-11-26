<script setup>
import {computed} from "vue";

const props = defineProps({
  name: {type: String, required: true},
  size: {type: [Number, String], default: 22},
})

// feather 风格线性图标（stroke=currentColor），每个图标由若干基础 SVG 元素组成
const icons = {
  login: [
    {tag: 'path', attrs: {d: 'M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4'}},
    {tag: 'polyline', attrs: {points: '10 17 15 12 10 7'}},
    {tag: 'line', attrs: {x1: 15, y1: 12, x2: 3, y2: 12}},
  ],
  'user-plus': [
    {tag: 'path', attrs: {d: 'M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2'}},
    {tag: 'circle', attrs: {cx: 8.5, cy: 7, r: 4}},
    {tag: 'line', attrs: {x1: 20, y1: 8, x2: 20, y2: 14}},
    {tag: 'line', attrs: {x1: 23, y1: 11, x2: 17, y2: 11}},
  ],
  book: [
    {tag: 'path', attrs: {d: 'M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z'}},
    {tag: 'path', attrs: {d: 'M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z'}},
  ],
  download: [
    {tag: 'path', attrs: {d: 'M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4'}},
    {tag: 'polyline', attrs: {points: '7 10 12 15 17 10'}},
    {tag: 'line', attrs: {x1: 12, y1: 15, x2: 12, y2: 3}},
  ],
  repeat: [
    {tag: 'polyline', attrs: {points: '17 1 21 5 17 9'}},
    {tag: 'path', attrs: {d: 'M3 11V9a4 4 0 0 1 4-4h14'}},
    {tag: 'polyline', attrs: {points: '7 23 3 19 7 15'}},
    {tag: 'path', attrs: {d: 'M21 13v2a4 4 0 0 1-4 4H3'}},
  ],
  shield: [
    {tag: 'path', attrs: {d: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z'}},
  ],
  link: [
    {tag: 'path', attrs: {d: 'M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71'}},
    {tag: 'path', attrs: {d: 'M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71'}},
  ],
  globe: [
    {tag: 'circle', attrs: {cx: 12, cy: 12, r: 10}},
    {tag: 'line', attrs: {x1: 2, y1: 12, x2: 22, y2: 12}},
    {tag: 'path', attrs: {d: 'M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z'}},
  ],
  server: [
    {tag: 'rect', attrs: {x: 2, y: 2, width: 20, height: 8, rx: 2, ry: 2}},
    {tag: 'rect', attrs: {x: 2, y: 14, width: 20, height: 8, rx: 2, ry: 2}},
    {tag: 'line', attrs: {x1: 6, y1: 6, x2: 6.01, y2: 6}},
    {tag: 'line', attrs: {x1: 6, y1: 18, x2: 6.01, y2: 18}},
  ],
  terminal: [
    {tag: 'polyline', attrs: {points: '4 17 10 11 4 5'}},
    {tag: 'line', attrs: {x1: 12, y1: 19, x2: 20, y2: 19}},
  ],
}

const els = computed(() => icons[props.name] || [])
</script>

<template>
  <svg
      :width="size"
      :height="size"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden="true"
      style="display:block"
  >
    <component v-for="(el, i) in els" :key="i" :is="el.tag" v-bind="el.attrs"/>
  </svg>
</template>
