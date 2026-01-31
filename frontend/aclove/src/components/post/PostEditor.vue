<script setup>
import { ref, computed, watch } from 'vue'
import {
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NButton,
  NSpace,
  NAlert,
  NCard,
  NSpin,
} from 'naive-ui'
import { useQuery } from '@tanstack/vue-query'
import MarkdownEditor from './MarkdownEditor.vue'
import { usePostStore } from '@/stores/post.js'
import { baseFetch } from '@/api/client.js'

const props = defineProps({
  categoryId: {
    type: Number,
    default: null,
  },
  editMode: {
    type: Boolean,
    default: false,
  },
  postId: {
    type: Number,
    default: null,
  },
  initialData: {
    type: Object,
    default: null,
  },
})

const emit = defineEmits(['submit', 'cancel', 'error'])

const postStore = usePostStore()

const formData = ref({
  category_id: props.categoryId,
  title: '',
  content: '',
  media_attachments: [],
})

const formError = ref('')
const isSubmitting = ref(false)

const titleRules = [
  { required: true, message: '请输入标题', trigger: 'blur' },
  { min: 5, message: '标题至少5个字符', trigger: 'blur' },
  { max: 100, message: '标题最多100个字符', trigger: 'blur' },
]

const contentRules = [
  { required: true, message: '请输入内容', trigger: 'blur' },
  { min: 10, message: '内容至少10个字符', trigger: 'blur' },
  { max: 10000, message: '内容最多10000个字符', trigger: 'blur' },
]

const categoryRules = [
  { required: true, type: 'number', message: '请选择分类', trigger: 'change' },
]

async function fetchCategories() {
  const response = await baseFetch('/categories')
  return Array.isArray(response) ? response : []
}

const { data: categories, isLoading: categoriesLoading } = useQuery({
  queryKey: ['categories'],
  queryFn: fetchCategories,
})

const categoryOptions = computed(() => {
  if (!categories.value) return []
  return flattenCategories(categories.value)
})

function flattenCategories(cats, depth = 0) {
  const result = []
  for (const cat of cats) {
    result.push({
      label: '  '.repeat(depth) + cat.name,
      value: cat.id,
    })
    if (cat.children && cat.children.length > 0) {
      result.push(...flattenCategories(cat.children, depth + 1))
    }
  }
  return result
}

watch(() => props.initialData, (newVal) => {
  if (newVal) {
    formData.value = {
      category_id: newVal.category_id || props.categoryId,
      title: newVal.title || '',
      content: newVal.content || '',
      media_attachments: newVal.media_attachments || [],
    }
  }
}, { immediate: true })

watch(() => props.categoryId, (newVal) => {
  if (newVal && !props.editMode) {
    formData.value.category_id = newVal
  }
})

const canSubmit = computed(() => {
  return (
    formData.value.category_id &&
    formData.value.title?.length >= 5 &&
    formData.value.title?.length <= 100 &&
    formData.value.content?.length >= 10 &&
    formData.value.content?.length <= 10000 &&
    !isSubmitting.value
  )
})

async function handleSubmit() {
  if (!canSubmit.value) return

  formError.value = ''
  isSubmitting.value = true

  try {
    const payload = {
      category_id: formData.value.category_id,
      title: formData.value.title.trim(),
      content: formData.value.content,
      media_attachments: formData.value.media_attachments,
    }

    let result
    if (props.editMode && props.postId) {
      const updatePayload = {}
      if (payload.title !== props.initialData?.title) {
        updatePayload.title = payload.title
      }
      if (payload.content !== props.initialData?.content) {
        updatePayload.content = payload.content
      }
      if (JSON.stringify(payload.media_attachments) !== JSON.stringify(props.initialData?.media_attachments)) {
        updatePayload.media_attachments = payload.media_attachments
      }
      result = await postStore.updatePost(props.postId, updatePayload)
    } else {
      result = await postStore.createPost(payload)
    }

    emit('submit', result)
  } catch (err) {
    formError.value = err.message || '提交失败，请稍后重试'
    emit('error', err)
  } finally {
    isSubmitting.value = false
  }
}

function handleCancel() {
  emit('cancel')
}

function handleContentChange(val) {
  formData.value.content = val
}
</script>

<template>
  <NSpin :show="isSubmitting">
    <NCard :title="editMode ? '编辑帖子' : '发布新帖'" class="post-editor-card">
      <NAlert v-if="formError" type="error" closable class="form-error" @close="formError = ''">
        {{ formError }}
      </NAlert>

      <NForm
        :model="formData"
        label-placement="left"
        label-width="80"
        require-mark-placement="right-hanging"
      >
        <NFormItem label="分类" path="category_id" :rule="categoryRules">
          <NSelect
            v-model:value="formData.category_id"
            :options="categoryOptions"
            placeholder="请选择分类"
            :loading="categoriesLoading"
            :disabled="editMode"
          />
        </NFormItem>

        <NFormItem label="标题" path="title" :rule="titleRules">
          <NInput
            v-model:value="formData.title"
            placeholder="请输入标题（5-100字符）"
            maxlength="100"
            show-count
          />
        </NFormItem>

        <NFormItem label="内容" path="content" :rule="contentRules">
          <MarkdownEditor
            v-model="formData.content"
            placeholder="请输入内容，支持 Markdown 格式（10-10000字符）"
            :max-length="10000"
            @change="handleContentChange"
          />
        </NFormItem>

        <NFormItem>
          <NSpace>
            <NButton
              type="primary"
              :disabled="!canSubmit"
              :loading="isSubmitting"
              @click="handleSubmit"
            >
              {{ editMode ? '保存修改' : '发布帖子' }}
            </NButton>
            <NButton @click="handleCancel">
              取消
            </NButton>
          </NSpace>
        </NFormItem>
      </NForm>
    </NCard>
  </NSpin>
</template>

<style scoped>
.post-editor-card {
  max-width: 900px;
  margin: 0 auto;
}

.form-error {
  margin-bottom: 16px;
}
</style>
