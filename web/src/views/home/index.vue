<script setup>
import {computed} from "vue";
import {useRouter} from "vue-router";
import {useThemeVars} from "naive-ui";
import {appStore} from "../../store/app.js";
import {localStore} from "../../store/local.js";
import PublicHeader from "../../components/PublicHeader.vue";
import HomeIcon from "../../components/HomeIcon.vue";

const router = useRouter()
const themeVars = useThemeVars()
const logged = computed(() => !!localStore().auth.token)
const registerOpen = computed(() => appStore().siteConfig.register === '1')

const entries = computed(() => {
  const list = [
    {title: '登录', desc: '已有账号？登录进入控制台', icon: 'login', action: () => router.push({name: 'Login'})},
  ]
  if (registerOpen.value) {
    list.push({title: '注册', desc: '免费创建账号，开始内网穿透', icon: 'user-plus', action: () => router.push({name: 'Login', query: {tab: 'register'}})})
  }
  list.push(
      {title: '帮助文档', desc: '快速上手、功能说明与常见问题', icon: 'book', href: '/docs/'},
      {title: '客户端下载', desc: 'Windows 图形版 / 命令行版、Linux', icon: 'download', action: () => router.push({name: 'Download'})},
  )
  return list
})

const features = [
  {title: '端口转发', desc: '内网 TCP 端口映射到节点公网端口，远程桌面、SSH 皆可。', icon: 'repeat'},
  {title: '私有隧道', desc: '不暴露公网端口，凭访问密钥由访客端接入，更私密。', icon: 'shield'},
  {title: 'P2P 隧道', desc: '两端尝试打洞直连，失败自动回退中继，延迟更低。', icon: 'link'},
  {title: '域名解析', desc: '自定义域名发布内网 HTTP 服务，配合网关服务使用。', icon: 'globe'},
  {title: '多节点管理', desc: '接入多个公网节点，中心化配置，网页修改实时生效。', icon: 'server'},
  {title: '客户端日志', desc: '在线查看客户端运行日志，按类型与关键词过滤，排障更轻松。', icon: 'terminal'},
]

const steps = [
  {title: '获取账号', desc: '获取账号并登录控制台'},
  {title: '下载客户端', desc: '在内网机器运行客户端并连接'},
  {title: '添加隧道', desc: '配置转发 / 隧道，实时生效'},
]

const goPrimary = () => {
  router.push({name: logged.value ? 'Dashboard' : 'Login'})
}

const openEntry = (item) => {
  if (item.href) {
    window.open(item.href, '_blank', 'noopener')
  } else {
    item.action()
  }
}
</script>

<template>
  <div class="home">
    <PublicHeader/>

    <section class="hero">
      <div class="hero-glow"></div>
      <div class="hero-inner">
        <span class="hero-badge">开源 · 基于 FRP</span>
        <h1 class="hero-title">{{ appStore().siteConfig.title }}</h1>
        <p class="hero-slogan">基于 FRP 的内网穿透管理平台</p>
        <p class="hero-desc">多用户 · 多节点 · 速率限制 · 中心化配置，通过网页修改配置，实时生效</p>
        <div class="hero-actions">
          <n-button type="primary" size="large" class="cta" @click="goPrimary">{{ logged ? '进入控制台' : '立即使用' }}</n-button>
          <n-button v-if="registerOpen && !logged" size="large" @click="router.push({name: 'Login', query: {tab: 'register'}})">免费注册</n-button>
          <n-button size="large" quaternary @click="router.push({name: 'Download'})">客户端下载</n-button>
        </div>
      </div>
    </section>

    <section class="block">
      <n-grid cols="1 s:2 m:4" responsive="screen" :x-gap="16" :y-gap="16">
        <n-gi v-for="item in entries" :key="item.title">
          <n-card class="entry-card" size="small" @click="openEntry(item)">
            <div class="entry-icon"><HomeIcon :name="item.icon" :size="22"/></div>
            <div class="entry-body">
              <div class="entry-title">{{ item.title }}</div>
              <div class="entry-desc">{{ item.desc }}</div>
            </div>
            <span class="entry-arrow">→</span>
          </n-card>
        </n-gi>
      </n-grid>
    </section>

    <section class="block">
      <div class="block-head">
        <h2 class="block-title">功能特性</h2>
        <p class="block-sub">覆盖常见内网穿透场景，按需选用</p>
      </div>
      <n-grid cols="1 s:2 m:3" responsive="screen" :x-gap="16" :y-gap="16">
        <n-gi v-for="item in features" :key="item.title">
          <n-card class="feature-card" size="small">
            <div class="feature-icon"><HomeIcon :name="item.icon" :size="22"/></div>
            <div class="feature-title">{{ item.title }}</div>
            <div class="feature-desc">{{ item.desc }}</div>
          </n-card>
        </n-gi>
      </n-grid>
    </section>

    <section class="block">
      <div class="block-head">
        <h2 class="block-title">三步开始</h2>
        <p class="block-sub">几分钟完成一次内网穿透</p>
      </div>
      <div class="steps">
        <div class="step" v-for="(item, index) in steps" :key="item.title">
          <div class="step-num">{{ index + 1 }}</div>
          <div class="step-title">{{ item.title }}</div>
          <div class="step-desc">{{ item.desc }}</div>
        </div>
      </div>
    </section>

    <footer class="footer">
      <span>{{ appStore().siteConfig.title }}</span>
      <span v-if="appStore().siteConfig.system?.version"> · {{ appStore().siteConfig.system.version }}</span>
      <span> · Apache-2.0 · </span>
      <a href="/docs/" target="_blank" rel="noopener">帮助文档</a>
    </footer>
  </div>
</template>

<style lang="scss" scoped>
.home {
  --primary: v-bind('themeVars.primaryColor');
  --primary-hover: v-bind('themeVars.primaryColorHover');
  --border: v-bind('themeVars.borderColor');
  --text-1: v-bind('themeVars.textColor1');
  --text-2: v-bind('themeVars.textColor2');
  --text-3: v-bind('themeVars.textColor3');

  min-height: 100vh;
  color: var(--text-2);
}

/* ---------- Hero ---------- */
.hero {
  position: relative;
  overflow: hidden;
  padding: 96px 20px 56px;
  text-align: center;

  .hero-glow {
    position: absolute;
    inset: 0;
    pointer-events: none;
    background:
        radial-gradient(600px 300px at 50% -40px, color-mix(in srgb, var(--primary) 20%, transparent), transparent 70%),
        radial-gradient(400px 220px at 85% 10%, color-mix(in srgb, var(--primary) 10%, transparent), transparent 70%);
  }

  .hero-inner {
    position: relative;
    max-width: 820px;
    margin: 0 auto;
    animation: rise .5s ease both;
  }

  .hero-badge {
    display: inline-block;
    padding: 4px 14px;
    border-radius: 999px;
    font-size: .82em;
    letter-spacing: .5px;
    color: var(--primary);
    background: color-mix(in srgb, var(--primary) 12%, transparent);
    border: 1px solid color-mix(in srgb, var(--primary) 28%, transparent);
    margin-bottom: 20px;
  }

  .hero-title {
    font-size: 3em;
    line-height: 1.1;
    letter-spacing: 1px;
    margin: 0;
    color: var(--text-1);
  }

  .hero-slogan {
    font-size: 1.35em;
    font-weight: 600;
    margin: 16px 0 10px;
    background: linear-gradient(90deg, var(--primary), var(--primary-hover));
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
  }

  .hero-desc {
    margin: 0 auto 32px;
    max-width: 620px;
    color: var(--text-3);
  }

  .hero-actions {
    display: flex;
    gap: 12px;
    justify-content: center;
    flex-wrap: wrap;

    .cta {
      box-shadow: 0 6px 18px color-mix(in srgb, var(--primary) 35%, transparent);
    }
  }
}

/* ---------- 区块 ---------- */
.block {
  max-width: 1100px;
  margin: 0 auto;
  padding: 28px 20px;
  animation: rise .5s ease both;

  .block-head {
    text-align: center;
    margin-bottom: 24px;
  }

  .block-title {
    font-size: 1.5em;
    margin: 0;
    color: var(--text-1);
    position: relative;
    display: inline-block;
    padding-bottom: 10px;

    &::after {
      content: '';
      position: absolute;
      left: 50%;
      bottom: 0;
      transform: translateX(-50%);
      width: 36px;
      height: 3px;
      border-radius: 2px;
      background: var(--primary);
    }
  }

  .block-sub {
    margin: 10px 0 0;
    font-size: .92em;
    color: var(--text-3);
  }
}

/* ---------- 入口卡片 ---------- */
.entry-card {
  height: 100%;
  cursor: pointer;
  border: 1px solid var(--border);
  transition: transform .2s ease, box-shadow .2s ease, border-color .2s ease;

  :deep(.n-card__content) {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  &:hover {
    transform: translateY(-3px);
    border-color: color-mix(in srgb, var(--primary) 55%, transparent);
    box-shadow: 0 8px 24px color-mix(in srgb, var(--primary) 18%, transparent);

    .entry-arrow {
      transform: translateX(3px);
      color: var(--primary);
    }
  }

  .entry-icon {
    flex: none;
    width: 42px;
    height: 42px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--primary);
    background: color-mix(in srgb, var(--primary) 12%, transparent);
  }

  .entry-body {
    flex: 1;
    min-width: 0;
  }

  .entry-title {
    font-size: 1.05em;
    font-weight: 600;
    margin-bottom: 4px;
    color: var(--text-1);
  }

  .entry-desc {
    font-size: .85em;
    color: var(--text-3);
  }

  .entry-arrow {
    flex: none;
    color: var(--text-3);
    transition: transform .2s ease, color .2s ease;
  }
}

/* ---------- 功能卡片 ---------- */
.feature-card {
  height: 100%;
  border: 1px solid var(--border);
  transition: transform .2s ease, box-shadow .2s ease, border-color .2s ease;

  &:hover {
    transform: translateY(-3px);
    border-color: color-mix(in srgb, var(--primary) 55%, transparent);
    box-shadow: 0 8px 24px color-mix(in srgb, var(--primary) 16%, transparent);
  }

  .feature-icon {
    width: 42px;
    height: 42px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--primary);
    background: color-mix(in srgb, var(--primary) 12%, transparent);
    margin-bottom: 14px;
  }

  .feature-title {
    font-weight: 600;
    margin-bottom: 6px;
    color: var(--text-1);
  }

  .feature-desc {
    font-size: .88em;
    line-height: 1.6;
    color: var(--text-3);
  }
}

/* ---------- 三步 ---------- */
.steps {
  display: flex;
  gap: 16px;
  position: relative;

  &::before {
    content: '';
    position: absolute;
    top: 20px;
    left: 16%;
    right: 16%;
    height: 2px;
    background: color-mix(in srgb, var(--primary) 22%, transparent);
  }

  .step {
    flex: 1;
    text-align: center;
    position: relative;
    padding: 0 8px;
  }

  .step-num {
    width: 40px;
    height: 40px;
    margin: 0 auto 14px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    color: #fff;
    background: var(--primary);
    box-shadow: 0 0 0 5px color-mix(in srgb, var(--primary) 15%, transparent);
    position: relative;
    z-index: 1;
  }

  .step-title {
    font-weight: 600;
    margin-bottom: 4px;
    color: var(--text-1);
  }

  .step-desc {
    font-size: .88em;
    color: var(--text-3);
  }
}

/* ---------- 页脚 ---------- */
.footer {
  max-width: 1100px;
  margin: 24px auto 0;
  padding: 28px 20px 40px;
  text-align: center;
  font-size: .9em;
  color: var(--text-3);
  border-top: 1px solid var(--border);

  a {
    color: var(--primary);
    text-decoration: none;

    &:hover {
      text-decoration: underline;
    }
  }
}

@keyframes rise {
  from {
    opacity: 0;
    transform: translateY(14px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .hero .hero-inner,
  .block {
    animation: none;
  }
  .entry-card,
  .feature-card,
  .entry-arrow {
    transition: none;
  }
}

@media (max-width: 600px) {
  .hero {
    padding: 56px 16px 32px;

    .hero-title {
      font-size: 2.2em;
    }

    .hero-slogan {
      font-size: 1.15em;
    }
  }

  .steps {
    flex-direction: column;
    gap: 24px;

    &::before {
      display: none;
    }
  }
}
</style>
