<script setup>
import { h, ref, computed, onUnmounted } from "vue";
import { useQuery } from "@tanstack/vue-query";
import {
  NLayout,
  NLayoutSider,
  NLayoutContent,
  NLayoutHeader,
  NMenu,
  NButton,
  NIcon,
  NAvatar,
  NSpace,
  NText,
  NCard,
  NEl,
  NGradientText,
  NDivider,
} from "naive-ui";
import {
  LayersOutline,
  MoonOutline,
  SunnyOutline,
  RefreshOutline,
  HeartOutline,
  SparklesOutline,
} from "@vicons/ionicons5";
import { baseFetch } from "@/api/client.js";
import Home from "@/components/Home.vue";
import mikuLogo from "@/assets/logo/miku.svg?url";

const collapsed = ref(false);
const activeKey = ref("home");
const isDark = ref(false);
const currentTime = ref(new Date().toLocaleString());
let timeInterval = null;

function toggleTheme() {
  isDark.value = !isDark.value;
  document.documentElement.classList.toggle("dark", isDark.value);
}

function categoryToMenuItem(category, parentKey = "") {
  const key = parentKey ? `${parentKey}-${category.id}` : `category-${category.id}`;
  const children =
    category.children && category.children.length > 0
      ? category.children.map((child) => categoryToMenuItem(child, key))
      : undefined;

  return {
    label: category.name || "未命名",
    key: key,
    id: category.id,
    icon: () => h(NIcon, null, { default: () => h(LayersOutline) }),
    ...(children && { children }),
  };
}

async function fetchCategories() {
  const response = await baseFetch("/categories");
  return Array.isArray(response) ? response : [];
}

const {
  data: categories,
  isLoading: categoriesLoading,
  refetch: refetchCategories,
  isFetching,
} = useQuery({
  queryKey: ["categories"],
  queryFn: fetchCategories,
});

const menuOptions = computed(() => {
  const categoryData = categories.value || [];
  return categoryData.map((cat) => categoryToMenuItem(cat));
});

function handleMenuUpdate(key) {
  activeKey.value = key;
  console.log("导航到:", key);
}

function handleRefreshCategories() {
  refetchCategories();
}

timeInterval = setInterval(() => {
  currentTime.value = new Date().toLocaleString();
}, 1000);

onUnmounted(() => {
  if (timeInterval) {
    clearInterval(timeInterval);
  }
});
</script>

<template>
  <NLayout has-sider class="app-layout">
    <NLayoutSider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="240"
      :collapsed="collapsed"
      show-trigger
      @collapse="collapsed = true"
      @expand="collapsed = false"
      :native-scrollbar="false"
      class="app-sider"
    >
      <div class="logo-container">
        <div class="logo">
          <NAvatar
            round
            size="small"
            :src="mikuLogo"
          />
          <span v-if="!collapsed" class="logo-text">ACLOVE</span>
        </div>
      </div>
      <NButton quaternary block @click="toggleTheme" class="theme-toggle">
        <template #icon>
          <NIcon>
            <MoonOutline v-if="!isDark" />
            <SunnyOutline v-else />
          </NIcon>
        </template>
        <span v-if="!collapsed">{{ isDark ? "浅色模式" : "深色模式" }}</span>
      </NButton>
      <NMenu
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="22"
        :options="menuOptions"
        :value="activeKey"
        @update:value="handleMenuUpdate"
        class="app-menu"
      />
    </NLayoutSider>

    <NLayoutContent class="app-content">
      <div class="home-container">
        <NCard class="welcome-card" :bordered="false">
          <div class="welcome-content">
            <NSpace vertical align="center" :size="24">
              <div class="icon-wrapper">
                <NIcon size="80" :depth="1">
                  <HeartOutline />
                </NIcon>
              </div>
              
              <NGradientText
                :size="48"
                :font-size="48"
                :weight="800"
                type="linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
              >
                欢迎来到aclove匿名版
              </NGradientText>
              
              <NDivider />
              
              <NText :depth="2" class="subtitle">
                <NSpace vertical align="center" :size="12">
                  <span>你所热爱的就是你的生活</span>
                  <NIcon :depth="3" size="24">
                    <SparklesOutline />
                  </NIcon>
                </NSpace>
              </NText>
            </NSpace>
          </div>
        </NCard>
      </div>
    </NLayoutContent>

  </NLayout>
</template>

<style scoped>
.app-layout {
  height: 100vh;
  width: 100%;
}

.app-sider {
  background: linear-gradient(180deg, #ffffff 0%, #fafafa 100%);
}

.dark .app-sider {
  background: linear-gradient(180deg, #1a1a1a 0%, #141414 100%);
}

.logo-container {
  padding: 16px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.dark .logo-container {
  border-bottom-color: rgba(255, 255, 255, 0.06);
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: #333;
  letter-spacing: 2px;
}

.dark .logo-text {
  color: #fff;
}

.app-menu {
  padding: 12px 8px;
}

.theme-toggle {
  margin: 0 12px 12px;
}

.app-content {
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e8ec 100%);
  min-height: 100%;
}

.dark .app-content {
  background: linear-gradient(135deg, #0f0f23 0%, #1a1a2e 100%);
}

.home-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 40px 24px;
}

.welcome-card {
  max-width: 800px;
  width: 100%;
  border-radius: 24px;
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.1);
  transition: transform 0.4s ease, box-shadow 0.4s ease;
}

.welcome-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 20px 64px rgba(0, 0, 0, 0.16);
}

.dark .welcome-card {
  background: rgba(30, 30, 46, 0.85);
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.4);
}

.welcome-content {
  padding: 64px 48px;
  text-align: center;
}

.icon-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 128px;
  height: 128px;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.12) 0%, rgba(118, 75, 162, 0.12) 100%);
  margin: 0 auto;
}

.dark .icon-wrapper {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.2) 0%, rgba(118, 75, 162, 0.2) 100%);
}

.subtitle {
  font-size: 28px;
  line-height: 1.8;
  letter-spacing: 3px;
}

.app-header {
  background: #fff;
  padding: 0 24px;
  height: 56px;
  display: flex;
  align-items: center;
}

.dark .app-header {
  background: #1a1a1a;
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
}

.main-layout {
  background: #f5f7fa;
}

.dark .main-layout {
  background: #141414;
}

.app-content {
  padding: 24px;
  background: #f5f7fa;
  height: calc(100vh - 56px);
}

.dark .app-content {
  background: #141414;
}
</style>
