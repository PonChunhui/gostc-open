<script setup>
import {computed} from "vue";
import {useRouter} from "vue-router";
import {useThemeVars} from "naive-ui";
import {appStore} from "../store/app.js";
import {localStore} from "../store/local.js";

const router = useRouter()
const themeVars = useThemeVars()
const logged = computed(() => !!localStore().auth.token)
</script>

<template>
  <header class="public-header">
    <div class="inner">
      <div class="brand" @click="router.push({name: 'Home'})">
        <img v-if="appStore().siteConfig.favicon" :src="appStore().siteConfig.favicon" class="logo" alt="">
        <span class="title">{{ appStore().siteConfig.title }}</span>
      </div>
      <nav class="nav">
        <a class="link" href="/docs/" target="_blank" rel="noopener">帮助文档</a>
        <router-link class="link" :to="{name: 'Download'}">客户端下载</router-link>
        <n-button v-if="!logged" size="small" type="primary" @click="router.push({name: 'Login'})">登录</n-button>
        <n-button v-else size="small" type="primary" @click="router.push({name: 'Dashboard'})">进入控制台</n-button>
      </nav>
    </div>
  </header>
</template>

<style lang="scss" scoped>
.public-header {
  --primary: v-bind('themeVars.primaryColor');
  position: sticky;
  top: 0;
  z-index: 10;
  backdrop-filter: blur(8px);
  border-bottom: 1px solid v-bind('themeVars.borderColor');
  background: color-mix(in srgb, v-bind('themeVars.cardColor') 82%, transparent);

  .inner {
    max-width: 1100px;
    margin: 0 auto;
    padding: 12px 20px;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    user-select: none;

    .logo {
      width: 26px;
      height: 26px;
      border-radius: 6px;
    }

    .title {
      font-size: 1.15em;
      font-weight: 600;
      color: v-bind('themeVars.textColor1');
    }
  }

  .nav {
    display: flex;
    align-items: center;
    gap: 18px;

    .link {
      position: relative;
      color: v-bind('themeVars.textColor2');
      text-decoration: none;
      font-size: .95em;
      padding-bottom: 2px;
      transition: color .2s ease;

      &::after {
        content: '';
        position: absolute;
        left: 0;
        bottom: -2px;
        width: 100%;
        height: 2px;
        border-radius: 1px;
        background: var(--primary);
        transform: scaleX(0);
        transform-origin: left;
        transition: transform .2s ease;
      }

      &:hover {
        color: var(--primary);

        &::after {
          transform: scaleX(1);
        }
      }
    }
  }

  @media (max-width: 600px) {
    .inner {
      padding: 10px 12px;
    }

    .nav {
      gap: 10px;

      .link {
        font-size: .85em;
      }
    }
  }
}
</style>