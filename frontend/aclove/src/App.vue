<script setup>
import { h, ref, computed, onUnmounted, onMounted, watch } from "vue";
import { useQuery } from "@tanstack/vue-query";
import {
  NLayout,
  NLayoutSider,
  NLayoutContent,
  NMenu,
  NButton,
  NIcon,
  NAvatar,
  NSpace,
  NCard,
  NGradientText,
  NDivider,
  NModal,
} from "naive-ui";
import {
  LayersOutline,
  MoonOutline,
  SunnyOutline,
  HeartOutline,
  SparklesOutline,
  HomeOutline,
} from "@vicons/ionicons5";
import { baseFetch } from "@/api/client.js";
import { useUserStore } from "@/stores/user.js";
import { usePostStore } from "@/stores/post.js";
import PostList from "@/components/post/PostList.vue";
import PostDetail from "@/components/post/PostDetail.vue";
import PostEditor from "@/components/post/PostEditor.vue";
import mikuLogo from "@/assets/logo/miku.svg?url";

const userStore = useUserStore();
const postStore = usePostStore();

const collapsed = ref(false);
const activeKey = ref("home");
const isDark = ref(false);
const currentView = ref("home");
const selectedCategoryId = ref(null);
const viewingPostId = ref(null);
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

function handleMenuUpdate(key) {
  activeKey.value = key;

  if (key === "home") {
    currentView.value = "home";
    selectedCategoryId.value = null;
  } else if (key.startsWith("category-")) {
    const categoryId = extractCategoryId(key);
    selectedCategoryId.value = categoryId;
    currentView.value = "category";
  }
}

function extractCategoryId(key) {
  const parts = key.split("-");
  return parts[parts.length - 1];
}

function handleCreatePost(categoryId) {
  selectedCategoryId.value = categoryId;
  editingPost.value = null;
  showEditor.value = true;
}

function handleViewPost(post) {
  viewingPostId.value = post.id;
  postStore.fetchPostDetail(post.id);
  currentView.value = "detail";
}

function handleEditPost(post) {
  editingPost.value = post;
  showEditor.value = true;
}

function handleDeletePost(post) {
  postStore.deletePost(post.id);
}

function handleEditorSubmit(post) {
  showEditor.value = false;
  editingPost.value = null;

  if (currentView.value === "detail" && post.id) {
    postStore.fetchPostDetail(post.id);
  } else if (post.id) {
    viewingPostId.value = post.id;
    postStore.fetchPostDetail(post.id);
    currentView.value = "detail";
  }
}

function handleEditorCancel() {
  showEditor.value = false;
  editingPost.value = null;
}

function handleBackFromDetail() {
  if (selectedCategoryId.value) {
    currentView.value = "category";
  } else {
    currentView.value = "home";
  }
  viewingPostId.value = null;
}

function handleBackFromDelete() {
  handleBackFromDetail();
  postStore.resetPosts();
  if (selectedCategoryId.value) {
    currentView.value = "category";
  } else {
    currentView.value = "home";
  }
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
      <div v-if="currentView === 'home'" class="home-container">
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

              <NSpace :size="16" style="margin-top: 24px">
                <NButton type="primary" size="large" @click="currentView = 'category'">
                  浏览帖子
                </NButton>
                <NButton size="large" @click="handleCreatePost">
                  发布帖子
                </NButton>
              </NSpace>
            </NSpace>
          </div>
        </NCard>
      </div>

      <div v-else-if="currentView === 'category'" class="content-container">
        <PostList
          :category-id="selectedCategoryId"
          @create-post="handleCreatePost"
          @view-post="handleViewPost"
          @edit-post="handleEditPost"
          @delete-post="handleDeletePost"
        />
      </div>

      <div v-else-if="currentView === 'detail'" class="content-container">
        <PostDetail
          :post-id="viewingPostId"
          @back="handleBackFromDetail"
          @edit="handleEditPost"
          @delete="handleBackFromDelete"
        />
      </div>

      <NModal
        v-model:show="showEditor"
        preset="card"
        :title="editingPost ? '编辑帖子' : '发布新帖'"
        style="width: 900px; max-width: 95vw"
        :mask-closable="false"
      >
        <PostEditor
          :category-id="selectedCategoryId"
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
</style>
