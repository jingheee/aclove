<script setup>
import { ref, computed, watch } from "vue";
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
  NTag,
  NTooltip,
  NIcon,
} from "naive-ui";
import { useQuery } from "@tanstack/vue-query";
import { InformationCircleOutline, ImageOutline } from "@vicons/ionicons5";
import RichTextEditor from "@/components/editor/RichTextEditor.vue";
import RichTextViewer from "@/components/editor/RichTextViewer.vue";
import { usePostStore } from "@/stores/post.js";
import { baseFetch } from "@/api/client.js";

const props = defineProps({
  categoryId: {
    type: [String, Number],
    default: null,
  },
  editMode: {
    type: Boolean,
    default: false,
  },
  postId: {
    type: [String, Number],
    default: null,
  },
  initialData: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(["submit", "cancel", "error"]);

const postStore = usePostStore();

const formData = ref({
  category_id: props.categoryId,
  title: "",
  content: "",
  editor_type: "rich_text",
});

const formError = ref("");
const isSubmitting = ref(false);
const uploadStatus = ref([]);
const showPreview = ref(false);

const titleRules = [
  { required: true, message: "请输入标题", trigger: "blur" },
  { min: 5, message: "标题至少5个字符", trigger: "blur" },
  { max: 100, message: "标题最多100个字符", trigger: "blur" },
];

const categoryRules = [
  { required: true, type: "number", message: "请选择分类", trigger: "change" },
];

async function fetchCategories() {
  const response = await baseFetch("/categories");
  return response || [];
}

const { data: categories, isLoading: categoriesLoading } = useQuery({
  queryKey: ["categories"],
  queryFn: fetchCategories,
});

const categoryOptions = computed(() => {
  if (!categories.value) return [];
  return flattenCategories(categories.value);
});

const selectedCategoryName = computed(() => {
  if (!formData.value.category_id || !categories.value) return "";
  const cat = findCategory(categories.value, formData.value.category_id);
  return cat?.name || "";
});

const hasImages = computed(() => {
  if (!formData.value.content) return false;
  return formData.value.content.includes("<img");
});

const imageCount = computed(() => {
  if (!formData.value.content) return 0;
  const matches = formData.value.content.match(/<img/g);
  return matches ? matches.length : 0;
});

function flattenCategories(cats, depth = 0) {
  const result = [];
  for (const cat of cats) {
    result.push({
      label: "  ".repeat(depth) + cat.name,
      value: cat.id,
    });
    if (cat.children && cat.children.length > 0) {
      result.push(...flattenCategories(cat.children, depth + 1));
    }
  }
  return result;
}

function findCategory(cats, id) {
  for (const cat of cats) {
    if (cat.id === id) return cat;
    if (cat.children) {
      const found = findCategory(cat.children, id);
      if (found) return found;
    }
  }
  return null;
}

watch(
  () => props.initialData,
  (newVal) => {
    if (newVal) {
      formData.value = {
        category_id: newVal.category_id || props.categoryId,
        title: newVal.title || "",
        content: newVal.content || "",
        editor_type: newVal.editor_type || "rich_text",
      };
    }
  },
  { immediate: true },
);

watch(
  () => props.categoryId,
  (newVal) => {
    if (newVal && !props.editMode) {
      formData.value.category_id = newVal;
    }
  },
);

const canSubmit = computed(() => {
  const hasContent = formData.value.content && 
    formData.value.content.replace(/<[^>]*>/g, "").trim().length >= 10;
  
  return (
    formData.value.category_id &&
    formData.value.title?.length >= 5 &&
    formData.value.title?.length <= 100 &&
    hasContent &&
    !isSubmitting.value &&
    !uploadStatus.value.some(u => u.status === "uploading")
  );
});

async function handleSubmit() {
  if (!canSubmit.value) return;

  formError.value = "";
  isSubmitting.value = true;

  try {
    const payload = {
      category_id: formData.value.category_id,
      title: formData.value.title.trim(),
      content: formData.value.content,
      editor_type: "rich_text",
    };

    let result;
    if (props.editMode && props.postId) {
      const updatePayload = {};
      if (payload.title !== props.initialData?.title) {
        updatePayload.title = payload.title;
      }
      if (payload.content !== props.initialData?.content) {
        updatePayload.content = payload.content;
      }
      if (payload.editor_type !== props.initialData?.editor_type) {
        updatePayload.editor_type = payload.editor_type;
      }
      result = await postStore.updatePost(props.postId, updatePayload);
    } else {
      result = await postStore.createPost(payload);
    }

    emit("submit", result);
  } catch (err) {
    formError.value = err.message || "提交失败，请稍后重试";
    emit("error", err);
  } finally {
    isSubmitting.value = false;
  }
}

function handleCancel() {
  emit("cancel");
}

function handleImageUpload({ file, status, url, error }) {
  const existingIndex = uploadStatus.value.findIndex(u => u.file.name === file.name);
  
  if (existingIndex >= 0) {
    uploadStatus.value[existingIndex] = { file, status, url, error };
  } else {
    uploadStatus.value.push({ file, status, url, error });
  }

  // 清理已完成的上传状态
  setTimeout(() => {
    uploadStatus.value = uploadStatus.value.filter(u => u.status === "uploading");
  }, 3000);
}

function togglePreview() {
  showPreview.value = !showPreview.value;
}

function stripHtml(html) {
  if (!html) return "";
  return html.replace(/<[^>]*>/g, "").trim();
}
</script>

<template>
  <NSpin :show="isSubmitting">
    <div class="rich-post-editor">
      <!-- Header -->
      <div class="editor-header">
        <h2 class="editor-title">
          {{ editMode ? "编辑帖子" : "发布新帖" }}
        </h2>
        <div class="editor-meta">
          <NTooltip>
            <template #trigger>
              <NTag size="small" type="info" class="editor-type-tag">
                <NIcon :component="InformationCircleOutline" class="tag-icon" />
                富文本
              </NTag>
            </template>
            支持图片、表格、代码块等丰富格式
          </NTooltip>
        </div>
      </div>

      <!-- Error Alert -->
      <NAlert
        v-if="formError"
        type="error"
        closable
        class="form-error"
        @close="formError = ''"
      >
        {{ formError }}
      </NAlert>

      <!-- Form -->
      <NForm
        :model="formData"
        label-placement="top"
        require-mark-placement="right-hanging"
        class="editor-form"
      >
        <!-- Category -->
        <NFormItem label="选择分类" path="category_id" :rule="categoryRules">
          <NSelect
            v-model:value="formData.category_id"
            :options="categoryOptions"
            placeholder="请选择分类"
            :loading="categoriesLoading"
            :disabled="editMode"
            size="large"
          />
        </NFormItem>

        <!-- Title -->
        <NFormItem label="标题" path="title" :rule="titleRules">
          <NInput
            v-model:value="formData.title"
            placeholder="请输入标题（5-100字符）"
            maxlength="100"
            show-count
            size="large"
          />
        </NFormItem>

        <!-- Rich Text Editor -->
        <NFormItem label="内容" path="content">
          <div class="editor-container">
            <RichTextEditor
              v-model="formData.content"
              placeholder="开始输入内容，支持富文本格式..."
              :max-length="10000"
              :disabled="isSubmitting"
              upload-endpoint="/api/upload/image"
              @image-upload="handleImageUpload"
            />
            
            <!-- Upload Status -->
            <div v-if="uploadStatus.length > 0" class="upload-status">
              <div
                v-for="(item, index) in uploadStatus"
                :key="index"
                class="upload-item"
                :class="`is-${item.status}`"
              >
                <NIcon :component="ImageOutline" />
                <span class="upload-filename">{{ item.file.name }}</span>
                <span class="upload-state">{{ 
                  item.status === "uploading" ? "上传中..." : 
                  item.status === "success" ? "已上传" : "失败" 
                }}</span>
              </div>
            </div>
          </div>
        </NFormItem>

        <!-- Content Stats -->
        <div class="content-stats">
          <NTag v-if="hasImages" size="small" type="success">
            <NIcon :component="ImageOutline" class="tag-icon" />
            {{ imageCount }} 张图片
          </NTag>
          <span class="content-hint">
            纯文本长度: {{ stripHtml(formData.content).length }} 字符
          </span>
        </div>

        <!-- Preview Toggle -->
        <div class="preview-section">
          <NButton text type="primary" @click="togglePreview">
            {{ showPreview ? "隐藏预览" : "显示预览" }}
          </NButton>
          
          <div v-if="showPreview" class="preview-content">
            <NCard title="预览" size="small">
              <RichTextViewer :content="formData.content" />
            </NCard>
          </div>
        </div>

        <!-- Actions -->
        <NFormItem class="form-actions">
          <NSpace>
            <NButton
              type="primary"
              size="large"
              :disabled="!canSubmit"
              :loading="isSubmitting"
              @click="handleSubmit"
            >
              {{ editMode ? "保存修改" : "发布帖子" }}
            </NButton>
            <NButton size="large" @click="handleCancel">
              取消
            </NButton>
          </NSpace>
        </NFormItem>
      </NForm>
    </div>
  </NSpin>
</template>

<style scoped>
.rich-post-editor {
  max-width: 900px;
  margin: 0 auto;
  padding: 24px;
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--n-border-color);
}

.editor-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0;
  color: var(--n-text-color);
}

.editor-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.editor-type-tag {
  display: flex;
  align-items: center;
  gap: 4px;
}

.tag-icon {
  font-size: 14px;
}

.form-error {
  margin-bottom: 20px;
}

.editor-form :deep(.n-form-item-label) {
  font-weight: 500;
  font-size: 0.95rem;
}

.editor-container {
  border-radius: 12px;
  overflow: hidden;
}

.upload-status {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.upload-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 0.9em;
  background: var(--n-action-color);
}

.upload-item.is-uploading {
  background: var(--n-primary-color-hover);
  color: var(--n-primary-color);
}

.upload-item.is-success {
  background: var(--n-success-color-hover);
  color: var(--n-success-color);
}

.upload-item.is-error {
  background: var(--n-error-color-hover);
  color: var(--n-error-color);
}

.upload-filename {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.upload-state {
  font-size: 0.85em;
  opacity: 0.8;
}

.content-stats {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding: 12px 16px;
  background: var(--n-action-color);
  border-radius: 8px;
}

.content-hint {
  font-size: 0.85em;
  color: var(--n-text-color-3);
}

.preview-section {
  margin-bottom: 24px;
}

.preview-content {
  margin-top: 16px;
}

.form-actions {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid var(--n-border-color);
}

@media (max-width: 768px) {
  .rich-post-editor {
    padding: 16px;
  }

  .editor-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .editor-title {
    font-size: 1.25rem;
  }
}
</style>
