<script setup>
import { computed, h } from "vue";
import {
  NCard,
  NSpace,
  NText,
  NTag,
  NIcon,
  NButton,
  NDropdown,
  NDivider,
} from "naive-ui";
import {
  EyeOutline,
  ChatbubbleOutline,
  ThumbsUpOutline,
  TimeOutline,
  CreateOutline,
  TrashOutline,
  EllipsisHorizontal,
  ArrowBackOutline,
} from "@vicons/ionicons5";
import { useUserStore } from "@/stores/user.js";
import RichTextViewer from "@/components/editor/RichTextViewer.vue";

const props = defineProps({
  post: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(["back", "edit", "delete"]);

const userStore = useUserStore();

const isAuthor = computed(() => {
  return userStore.userInfo?.id === props.post.author?.id;
});

const isEdited = computed(() => {
  return props.post.edit_count > 0;
});

const isRichText = computed(() => {
  return props.post.editor_type === "rich_text" || 
    (props.post.content && props.post.content.includes("<"));
});

function formatTime(dateString) {
  if (!dateString) return "";
  const date = new Date(dateString);
  return date.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatRelativeTime(dateString) {
  const date = new Date(dateString);
  const now = new Date();
  const diff = now - date;

  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  if (minutes < 1) return "刚刚";
  if (minutes < 60) return `${minutes}分钟前`;
  if (hours < 24) return `${hours}小时前`;
  if (days < 30) return `${days}天前`;

  return formatTime(dateString);
}

function handleBack() {
  emit("back");
}

function handleEdit() {
  emit("edit", props.post);
}

function handleDelete() {
  emit("delete", props.post);
}

const dropdownOptions = [
  {
    label: "编辑",
    key: "edit",
    icon: () => h(NIcon, null, { default: () => h(CreateOutline) }),
    show: isAuthor.value,
  },
  {
    label: "删除",
    key: "delete",
    icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
    show: isAuthor.value,
  },
].filter((opt) => opt.show);

function handleDropdownSelect(key) {
  if (key === "edit") {
    handleEdit();
  } else if (key === "delete") {
    handleDelete();
  }
}
</script>

<template>
  <div class="post-detail">
    <!-- Back Button -->
    <NButton text class="back-btn" @click="handleBack">
      <NIcon size="20">
        <ArrowBackOutline />
      </NIcon>
      返回列表
    </NButton>

    <NCard class="post-content-card">
      <!-- Header -->
      <div class="post-header">
        <div class="post-meta">
          <NTag size="small" type="info" class="anonymous-tag">
            匿名
          </NTag>
          <NText depth="3" class="post-time">
            <NIcon size="14">
              <TimeOutline />
            </NIcon>
            {{ formatRelativeTime(post.created_at) }}
          </NText>
          <span v-if="isEdited" class="edited-badge">
            已编辑 {{ formatRelativeTime(post.last_edited_at) }}
          </span>
        </div>

        <NDropdown
          v-if="dropdownOptions.length > 0"
          :options="dropdownOptions"
          @select="handleDropdownSelect"
          trigger="click"
        >
          <NButton text class="more-btn">
            <NIcon size="20">
              <EllipsisHorizontal />
            </NIcon>
          </NButton>
        </NDropdown>
      </div>

      <!-- Title -->
      <h1 class="post-title">{{ post.title }}</h1>

      <!-- Category -->
      <div v-if="post.category" class="post-category">
        <NTag size="small" type="success">
          {{ post.category.name }}
        </NTag>
      </div>

      <NDivider />

      <!-- Content -->
      <div class="post-body">
        <RichTextViewer
          v-if="isRichText"
          :content="post.content"
        />
        <div v-else class="plain-content">
          {{ post.content }}
        </div>
      </div>

      <NDivider />

      <!-- Footer Stats -->
      <div class="post-footer">
        <NSpace align="center" :size="24">
          <NText depth="3" class="stat-item">
            <NIcon size="18">
              <EyeOutline />
            </NIcon>
            {{ post.view_count || 0 }} 浏览
          </NText>

          <NText depth="3" class="stat-item">
            <NIcon size="18">
              <ChatbubbleOutline />
            </NIcon>
            {{ post.reply_count || 0 }} 回复
          </NText>

          <NText depth="3" class="stat-item">
            <NIcon size="18">
              <ThumbsUpOutline />
            </NIcon>
            {{ post.upvote_count || 0 }} 赞
          </NText>
        </NSpace>

        <div v-if="post.created_at" class="post-time-detail">
          发布于 {{ formatTime(post.created_at) }}
        </div>
      </div>
    </NCard>
  </div>
</template>

<style scoped>
.post-detail {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

.back-btn {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
}

.post-content-card {
  border-radius: 16px;
}

.post-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.post-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.anonymous-tag {
  font-size: 12px;
}

.post-time {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
}

.edited-badge {
  font-size: 12px;
  color: var(--n-text-color-3);
  background: var(--n-action-color);
  padding: 2px 8px;
  border-radius: 4px;
}

.more-btn {
  opacity: 0.6;
  transition: opacity 0.2s;
}

.more-btn:hover {
  opacity: 1;
}

.post-title {
  font-size: 1.75rem;
  font-weight: 600;
  margin: 0 0 12px 0;
  line-height: 1.4;
  color: var(--n-text-color);
}

.post-category {
  margin-bottom: 8px;
}

.post-body {
  font-size: 16px;
  line-height: 1.8;
  color: var(--n-text-color);
}

.plain-content {
  white-space: pre-wrap;
  word-wrap: break-word;
}

.post-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
}

.post-time-detail {
  font-size: 13px;
  color: var(--n-text-color-3);
}

@media (max-width: 768px) {
  .post-detail {
    padding: 12px;
  }

  .post-title {
    font-size: 1.4rem;
  }

  .post-footer {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
