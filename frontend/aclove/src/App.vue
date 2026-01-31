<script setup>
import { h, ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useQuery } from "@tanstack/vue-query";
import {
  NLayout,
  NLayoutSider,
  NLayoutContent,
  NMenu,
  NButton,
  NIcon,
  NAvatar,
  NModal,
} from "naive-ui";
import {
  LayersOutline,
  MoonOutline,
  SunnyOutline,
  HomeOutline,
} from "@vicons/ionicons5";
import { baseFetch } from "@/api/client.js";
import { useUserStore } from "@/stores/user.js";
import RichPostEditor from "@/components/post/RichPostEditor.vue";
import mikuLogo from "@/assets/logo/miku.svg?url";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();

const collapsed = ref(false);
const isDark = ref(false);
const editingPost = ref(null);
const showEditor = ref(false);

onMounted(async () => {
  try {
    await userStore.initSession();
  } catch (err) {
    console.error("Session 初始化失败:", err);
  }
});

function toggleTheme() {
  isDark.value = !isDark.value;
  document.documentElement.classList.toggle("dark", isDark.value);
}

function categoryToMenuItem(category, parentKey = "") {
  const key = parentKey
    ? `${parentKey}-${category.id}`
    : `category-${category.id}`;
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
  return response || [];
}

const { data: categories } = useQuery({
  queryKey: ["categories"],
  queryFn: fetchCategories,
});

const menuOptions = computed(() => {
  const baseOptions = [
    {
      label: "首页",
      key: "home",
      icon: () => h(NIcon, null, { default: () => h(HomeOutline) }),
    },
  ];

  const categoryData = categories.value || [];
  const categoryOptions = categoryData.map((cat) => categoryToMenuItem(cat));

  return [...baseOptions, ...categoryOptions];
});

const activeKey = computed(() => {
  if (route.name === "home") {
    return "home";
  }
  if (route.name === "category" && route.params.categoryId) {
    return findCategoryKey(route.params.categoryId, categories.value);
  }
  return null;
});

function findCategoryKey(categoryId, categoryList, parentKey = "") {
  if (!categoryList) return null;

  for (const cat of categoryList) {
    const currentKey = parentKey
      ? `${parentKey}-${cat.id}`
      : `category-${cat.id}`;
    if (String(cat.id) === String(categoryId)) {
      return currentKey;
    }
    if (cat.children && cat.children.length > 0) {
      const found = findCategoryKey(categoryId, cat.children, currentKey);
      if (found) return found;
    }
  }
  return null;
}

function handleMenuUpdate(key) {
  if (key === "home") {
    router.push({ name: "home" });
  } else if (key.startsWith("category-")) {
    const categoryId = extractCategoryId(key);
    router.push({ name: "category", params: { categoryId } });
  }
}

function extractCategoryId(key) {
  const parts = key.split("-");
  return parts[parts.length - 1];
}

function handleCreatePost() {
  editingPost.value = null;
  showEditor.value = true;
}

function handleEditorSubmit(post) {
  showEditor.value = false;
  editingPost.value = null;

  if (post.id) {
    router.push({ name: "post-detail", params: { postId: post.id } });
  }
}

function handleEditorCancel() {
  showEditor.value = false;
  editingPost.value = null;
}


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
          <NAvatar round size="small" :src="mikuLogo" />
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
      <router-view v-slot="{ Component }">
        <component :is="Component" @create-post="handleCreatePost" />
      </router-view>

      <NModal
        v-model:show="showEditor"
        preset="card"
        :title="editingPost ? '编辑帖子' : '发布新帖'"
        style="width: 900px; max-width: 95vw"
        :mask-closable="false"
      >
        <RichPostEditor
          :category-id="route.params.categoryId"
          :edit-mode="!!editingPost"
          :post-id="editingPost?.id"
          :initial-data="editingPost"
          @submit="handleEditorSubmit"
          @cancel="handleEditorCancel"
        />
      </NModal>
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
  overflow-y: auto;
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

.content-container {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.welcome-card {
  max-width: 800px;
  width: 100%;
  border-radius: 24px;
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.1);
  transition:
    transform 0.4s ease,
    box-shadow 0.4s ease;
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
  background: linear-gradient(
    135deg,
    rgba(102, 126, 234, 0.12) 0%,
    rgba(118, 75, 162, 0.12) 100%
  );
  margin: 0 auto;
}

.dark .icon-wrapper {
  background: linear-gradient(
    135deg,
    rgba(102, 126, 234, 0.2) 0%,
    rgba(118, 75, 162, 0.2) 100%
  );
}

.subtitle {
  font-size: 28px;
  line-height: 1.8;
  letter-spacing: 3px;
}
</style>
