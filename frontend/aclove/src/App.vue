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
} from "naive-ui";
import {
  LayersOutline,
  MoonOutline,
  SunnyOutline,
  RefreshOutline,
} from "@vicons/ionicons5";
import { baseFetch } from "@/api/client.js";
import Home from "@/components/Home.vue";

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

const fixedMenuItems = [
  {
    label: "首页",
    key: "home",
    icon: () => h(NIcon, null, { default: () => h(HomeOutline) }),
  },
  {
    label: "数据统计",
    key: "statistics",
    icon: () => h(NIcon, null, { default: () => h(AnalyticsOutline) }),
  },
  {
    label: "用户管理",
    key: "users",
    icon: () => h(NIcon, null, { default: () => h(PeopleOutline) }),
  },
  {
    label: "系统设置",
    key: "settings",
    icon: () => h(NIcon, null, { default: () => h(SettingsOutline) }),
  },
];

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
            src="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCIgd2lkdGg9IjI0IiBoZWlnaHQ9IjI0Ij48cGF0aCBmaWxsPSIjMTBhMDU4IiBkPSJNMTIgMkM2LjUgMiAyIDYuNSAyIDEyYzAgNS41IDQuNSAxMCAxMCAxMHMxMC00LjUgMTAtMTBjMC01LjUtNC41LTEwLTEwLTEwem0zIDdoLTZ2NmgtMnY2aC0ydjZoLTJ2LTZoMnY2aDJ2LTZoMnY2aDJ2LTZ6bTYgMGgtNnY2aC0ydjZoLTJ2LTZoMnY2aDJ2LTZoMnY2aDJ2LTZ6bS02IDhoLTZ2NmgtMnY2aC0ydjZoLTJ2LTZoMnY2aDJ2LTZoMnY2aDJ2LTZ6"
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

    <NLayout class="main-layout">
      <NLayoutHeader bordered class="app-header">
        <div class="header-content">
          <div class="header-left">
            <NButton
              quaternary
              circle
              :loading="isFetching"
              @click="handleRefreshCategories"
            >
              <template #icon>
                <NIcon><RefreshOutline /></NIcon>
              </template>
            </NButton>
          </div>
          <div class="header-right">
            <NSpace align="center" :size="16">
              <NText depth="3">{{ currentTime }}</NText>
              <NButton quaternary circle @click="toggleTheme">
                <template #icon>
                  <NIcon>
                    <MoonOutline v-if="!isDark" />
                    <SunnyOutline v-else />
                  </NIcon>
                </template>
              </NButton>
            </NSpace>
          </div>
        </div>
      </NLayoutHeader>

      <NLayoutContent class="app-content" :native-scrollbar="false">
        <Home
          :categories="categories || []"
          :loading="categoriesLoading"
          @refresh-categories="handleRefreshCategories"
        />
      </NLayoutContent>
    </NLayout>
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
