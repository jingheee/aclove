<script setup>
import { ref, computed, h } from 'vue'
import {
  NCard,
  NSpace,
  NText,
  NTag,
  NButton,
  NIcon,
  NDivider,
  NAlert,
  NSpin,
  NModal,
  NPopconfirm,
} from 'naive-ui'
import {
  EyeOutline,
  ChatbubbleOutline,
  ThumbsUpOutline,
  ThumbsDownOutline,
  ArrowBackOutline,
  CreateOutline,
  TrashOutline,
  TimeOutline,
  CalendarOutline,
} from '@vicons/ionicons5'
import MarkdownPreview from './MarkdownPreview.vue'
import { useUserStore } from '@/stores/user.js'
import { usePostStore } from '@/stores/post.js'

const props = defineProps({
  postId: {
    type: Number,
    required: true,
  },
})

const emit = defineEmits(['back', 'edit', 'delete'])

const userStore = useUserStore()
const postStore = usePostStore()

const showDeleteConfirm = ref(false)

const post = computed(() => postStore.currentPost)
const loading = computed(() => postStore.loading)
const error = computed(() => postStore.error)

const isAuthor = computed(() => {
  return userStore.userInfo?.id === post.value?.author?.id
})

const isEdited = computed(() => {
  return post.value?.edit_count > 0
})

function formatDate(dateString) {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatTimeAgo(dateString) {
  const date = new Date(dateString)
  const now = new Date()
  const diff = now - date

  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 30) return `${days}天前`

  return formatDate(dateString)
}

function handleBack() {
  emit('back')
}

function handleEdit() {
  emit('edit', post.value)
}

function handleDelete() {
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  try {
    await postStore.deletePost(props.postId)
    showDeleteConfirm.value = false
    emit('delete', post.value)
  } catch (err) {
    console.error('删除失败:', err)
  }
}
</script>

<template>
  <div class="post-detail">
    <NButton
      quaternary
      class="back-btn"
      @click="handleBack"
    >
      <template #icon>
        <NIcon>
          <ArrowBackOutline />
        </NIcon>
      </template>
      返回列表
    </NButton>

    <NAlert
      v-if="error"
      type="error"
      closable
      class="error-alert"
      @close="postStore.error = null"
    >
      {{ error }}
    </NAlert>

    <NSpin :show="loading && !post">
      <template v-if="post">
        <NCard class="detail-card">
          <div class="detail-header">
            <div class="header-top">
              <NSpace align="center" :size="12">
                <NTag type="info" size="small">
                  匿名用户
                </NTag>
                <NText depth="3" class="time-text">
                  <NIcon size="14">
                    <TimeOutline />
                  </NIcon>
                  {{ formatTimeAgo(post.created_at) }}
                </NText>
                <NText v-if="isEdited" depth="3" class="edited-text">
                  (已编辑 {{ post.edit_count }} 次)
                </NText>
              </NSpace>

              <NSpace v-if="isAuthor">
                <NButton
                  quaternary
                  size="small"
                  @click="handleEdit"
                >
                  <template #icon>
                    <NIcon>
                      <CreateOutline />
                    </NIcon>
                  </template>
                  编辑
                </NButton>
                <NButton
                  quaternary
                  size="small"
                  type="error"
                  @click="handleDelete"
                >
                  <template #icon>
                    <NIcon>
                      <TrashOutline />
                    </NIcon>
                  </template>
                  删除
                </NButton>
              </NSpace>
            </div>

            <h1 class="detail-title">{{ post.title }}</h1>

            <div class="meta-info">
              <NSpace :size="20">
                <NText depth="3" class="meta-item">
                  <NIcon size="16">
                    <CalendarOutline />
                  </NIcon>
                  发布于 {{ formatDate(post.created_at) }}
                </NText>
                <NText v-if="isEdited" depth="3" class="meta-item">
                  <NIcon size="16">
                    <CreateOutline />
                  </NIcon>
                  最后编辑于 {{ formatDate(post.last_edited_at) }}
                </NText>
              </NSpace>
            </div>
          </div>

          <NDivider />

          <div class="detail-content">
            <MarkdownPreview :content="post.content" />
          </div>

          <NDivider />

          <div class="detail-footer">
            <NSpace align="center" :size="24">
              <NButton quaternary>
                <template #icon>
                  <NIcon>
                    <ThumbsUpOutline />
                  </NIcon>
                </template>
                {{ post.upvote_count || 0 }}
              </NButton>

              <NButton quaternary>
                <template #icon>
                  <NIcon>
                    <ThumbsDownOutline />
                  </NIcon>
                </template>
                {{ post.downvote_count || 0 }}
              </NButton>

              <NText depth="3" class="stat-item">
                <NIcon size="16">
                  <EyeOutline />
                </NIcon>
                {{ post.view_count || 0 }} 浏览
              </NText>

              <NText depth="3" class="stat-item">
                <NIcon size="16">
                  <ChatbubbleOutline />
                </NIcon>
                {{ post.reply_count || 0 }} 回复
              </NText>
            </NSpace>
          </div>
        </NCard>

        <NCard title="评论" class="comments-card">
          <NEmpty description="评论功能即将上线，敬请期待" />
        </NCard>
      </template>

      <NEmpty
        v-else-if="!loading"
        description="帖子不存在或已被删除"
        class="empty-state"
      >
        <template #extra>
          <NButton @click="handleBack">
            返回列表
          </NButton>
        </template>
      </NEmpty>
    </NSpin>

    <NModal
      v-model:show="showDeleteConfirm"
      preset="dialog"
      title="确认删除"
      type="warning"
      positive-text="确认删除"
      negative-text="取消"
      @positive-click="confirmDelete"
    >
      确定要删除这篇帖子吗？此操作不可恢复。
    </NModal>
  </div>
</template>

<style scoped>
.post-detail {
  max-width: 900px;
  margin: 0 auto;
  padding: 16px;
}

.back-btn {
  margin-bottom: 16px;
}

.error-alert {
  margin-bottom: 16px;
}

.detail-card {
  margin-bottom: 16px;
}

.detail-header {
  padding: 8px 0;
}

.header-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.time-text {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
}

.edited-text {
  font-size: 12px;
  color: #999;
}

.detail-title {
  font-size: 24px;
  font-weight: 600;
  line-height: 1.4;
  margin: 0 0 12px 0;
  color: var(--n-text-color);
}

.meta-info {
  margin-top: 8px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
}

.detail-content {
  padding: 16px 0;
  font-size: 15px;
  line-height: 1.8;
}

.detail-footer {
  padding-top: 8px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
}

.comments-card {
  margin-top: 16px;
}

.empty-state {
  padding: 60px 0;
}
</style>
