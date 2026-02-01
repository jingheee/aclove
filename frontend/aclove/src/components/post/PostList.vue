<script setup>
import { ref, computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  NSpace,
  NButton,
  NEmpty,
  NSpin,
  NAlert,
  NSelect,
  NPagination,
} from "naive-ui";
import { AddOutline, RefreshOutline } from "@vicons/ionicons5";
import PostCard from "./PostCard.vue";
import { usePostStore } from "@/stores/post.js";

const props = defineProps({
  categoryId: {
    type: [String, Number],
    default: null,
  },
});

const emit = defineEmits(["create-post"]);

const route = useRoute();
const router = useRouter();
const postStore = usePostStore();

const sortOptions = [
  { label: "最新发布", value: "newest" },
  { label: "热度排序", value: "hot" },
  { label: "最多点赞", value: "top" },
];

const currentSort = ref("newest");

const hasError = computed(() => !!postStore.error);
const isLoading = computed(() => postStore.loading);
const posts = computed(() => postStore.posts);
const total = computed(() => postStore.total);
const currentPage = computed(() => postStore.currentPage);
const pageSize = computed(() => postStore.pageSize);

const totalPages = computed(() => Math.ceil(total.value / pageSize.value));

const effectiveCategoryId = computed(() => {
  return props.categoryId || route.params.categoryId;
});

async function loadPosts() {
  try {
    const categoryId = effectiveCategoryId.value;
    await postStore.fetchPosts({
      category_id: categoryId === "all" ? null : categoryId,
      sort: currentSort.value,
    });
  } catch (err) {
    console.error("加载帖子失败:", err);
  }
}

function handlePageChange(page) {
  postStore.setPage(page);
  loadPosts();
}

function handleSortChange(value) {
  currentSort.value = value;
  postStore.setPage(1);
  loadPosts();
}

function handleRefresh() {
  postStore.setPage(1);
  loadPosts();
}

function handleCreatePost() {
  emit("create-post", effectiveCategoryId.value);
}

function handleViewPost(post) {
  router.push({ name: "post-detail", params: { postId: post.id } });
}

function handleEditPost(post) {
  emit("edit-post", post);
}

function handleDeletePost(post) {
  postStore.deletePost(post.id);
}

watch(
  () => effectiveCategoryId.value,
  () => {
    postStore.resetPosts();
    loadPosts();
  },
  { immediate: true },
);
</script>

<template>
  <div class="content-container">
    <div class="post-list">
      <div class="post-list-header">
        <NSpace align="center">
          <NSelect
            v-model:value="currentSort"
            :options="sortOptions"
            style="width: 120px"
            @update:value="handleSortChange"
          />
          <NButton quaternary circle @click="handleRefresh">
            <template #icon>
              <RefreshOutline />
            </template>
          </NButton>
        </NSpace>

        <NButton type="primary" @click="handleCreatePost">
          <template #icon>
            <AddOutline />
          </template>
          发布帖子
        </NButton>
      </div>

      <NAlert
        v-if="hasError"
        type="error"
        closable
        class="error-alert"
        @close="postStore.error = null"
      >
        {{ postStore.error }}
      </NAlert>

      <NSpin :show="isLoading && posts.length === 0">
        <div v-if="posts.length > 0" class="posts-container">
          <PostCard
            v-for="post in posts"
            :key="post.id"
            :post="post"
            @click="handleViewPost"
            @edit="handleEditPost"
            @delete="handleDeletePost"
          />
        </div>

        <NEmpty
          v-else-if="!isLoading"
          description="暂无帖子"
          class="empty-state"
        >
          <template #extra>
            <NButton type="primary" @click="handleCreatePost">
              发布第一个帖子
            </NButton>
          </template>
        </NEmpty>
      </NSpin>

      <div v-if="totalPages > 1" class="pagination-wrapper">
        <NPagination
          :page="currentPage"
          :page-count="totalPages"
          :page-size="pageSize"
          show-size-picker
          :page-sizes="[10, 20, 50]"
          @update:page="handlePageChange"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.content-container {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.post-list {
  padding: 16px 0;
}

.post-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 0 8px;
}

.error-alert {
  margin-bottom: 16px;
}

.posts-container {
  min-height: 200px;
}

.empty-state {
  padding: 60px 0;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--n-border-color);
}
</style>
