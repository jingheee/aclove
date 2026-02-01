<script setup>
import { computed } from "vue";
import {
  NCard,
  NSpace,
  NText,
  NTag,
  NIcon,
  NButton,
  NDropdown,
} from "naive-ui";
import {
  EyeOutline,
  ChatbubbleOutline,
  ThumbsUpOutline,
  EllipsisHorizontal,
  CreateOutline,
  TrashOutline,
  TimeOutline,
} from "@vicons/ionicons5";
import { useUserStore } from "@/stores/user.js";

const props = defineProps({
  post: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(["click", "edit", "delete"]);

const userStore = useUserStore();

const isAuthor = computed(() => {
  return userStore.userInfo?.id === props.post.author?.id;
});

const hasMedia = computed(() => {
  return props.post.media_count > 0;
});

const isEdited = computed(() => {
  return props.post.edit_count > 0;
});

function formatTime(dateString) {
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

  return date.toLocaleDateString("zh-CN");
}

function handleClick() {
  emit("click", props.post);
}

function handleEdit(e) {
  e.stopPropagation();
  emit("edit", props.post);
}

function handleDelete(e) {
  e.stopPropagation();
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
    emit("edit", props.post);
  } else if (key === "delete") {
    emit("delete", props.post);
  }
}
</script>

<template>
  <NCard
    class="post-card"
    :class="{ 'has-media': hasMedia }"
    hoverable
    @click="handleClick"
  >
    <div class="post-header">
      <div class="post-meta">
        <NTag size="small" type="info" class="anonymous-tag"> 匿名 </NTag>
        <NText depth="3" class="post-time">
          <NIcon size="14">
            <TimeOutline />
          </NIcon>
          {{ formatTime(post.created_at) }}
          <span v-if="isEdited" class="edited-tag">(已编辑)</span>
        </NText>
      </div>

      <NDropdown
        v-if="dropdownOptions.length > 0"
        :options="dropdownOptions"
        @select="handleDropdownSelect"
        trigger="click"
      >
        <NButton text class="more-btn" @click.stop>
          <NIcon size="20">
            <EllipsisHorizontal />
          </NIcon>
        </NButton>
      </NDropdown>
    </div>

    <h3 class="post-title">{{ post.title }}</h3>

    <p class="post-summary">{{ post.summary }}</p>

    <div class="post-footer">
      <NSpace align="center" :size="16">
        <NText depth="3" class="stat-item">
          <NIcon size="16">
            <EyeOutline />
          </NIcon>
          {{ post.view_count || 0 }}
        </NText>

        <NText depth="3" class="stat-item">
          <NIcon size="16">
            <ChatbubbleOutline />
          </NIcon>
          {{ post.reply_count || 0 }}
        </NText>

        <NText depth="3" class="stat-item">
          <NIcon size="16">
            <ThumbsUpOutline />
          </NIcon>
          {{ post.upvote_count || 0 }}
        </NText>

        <NTag v-if="hasMedia" size="small" type="success"> 有附件 </NTag>
      </NSpace>
    </div>
  </NCard>
</template>

<style scoped>
.post-card {
  cursor: pointer;
  transition: all 0.2s ease;
  margin-bottom: 12px;
}

.post-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.post-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.post-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.anonymous-tag {
  font-size: 12px;
}

.post-time {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
}

.edited-tag {
  color: #999;
  font-size: 12px;
}

.more-btn {
  opacity: 0;
  transition: opacity 0.2s;
}

.post-card:hover .more-btn {
  opacity: 1;
}

.post-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 8px 0;
  line-height: 1.5;
  color: var(--n-text-color);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.post-summary {
  font-size: 14px;
  color: var(--n-text-color-2);
  margin: 0 0 12px 0;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.post-footer {
  padding-top: 12px;
  border-top: 1px solid var(--n-border-color);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
}
</style>
