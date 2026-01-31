<script setup>
import { ref, watch, onMounted } from "vue";
import { Editor } from "@bytemd/vue-next";
import gfm from "@bytemd/plugin-gfm";
import highlight from "@bytemd/plugin-highlight";
import breaks from "@bytemd/plugin-breaks";
import "bytemd/dist/index.css";
import "github-markdown-css/github-markdown.css";

const props = defineProps({
  modelValue: {
    type: String,
    default: "",
  },
  placeholder: {
    type: String,
    default: "请输入内容...",
  },
  maxLength: {
    type: Number,
    default: 10000,
  },
});

const emit = defineEmits(["update:modelValue", "change"]);

const content = ref(props.modelValue);

const plugins = [gfm(), highlight(), breaks()];

watch(
  () => props.modelValue,
  (newVal) => {
    if (newVal !== content.value) {
      content.value = newVal;
    }
  },
);

watch(content, (newVal) => {
  emit("update:modelValue", newVal);
  emit("change", newVal);
});

function handleChange(val) {
  content.value = val;
}
</script>

<template>
  <div class="markdown-editor">
    <Editor
      :value="content"
      :plugins="plugins"
      :placeholder="placeholder"
      @change="handleChange"
    />
    <div class="editor-footer">
      <span
        class="char-count"
        :class="{ 'over-limit': content?.length > maxLength }"
      >
        {{ content?.length || 0 }} / {{ maxLength }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.markdown-editor {
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
  overflow: hidden;
}

.markdown-editor :deep(.bytemd) {
  height: 400px;
  border: none;
}

.markdown-editor :deep(.bytemd-toolbar) {
  border-bottom: 1px solid var(--n-border-color);
}

.markdown-editor :deep(.bytemd-editor) {
  font-family: "JetBrains Mono", "Fira Code", monospace;
  font-size: 14px;
  line-height: 1.6;
}

.markdown-editor :deep(.bytemd-preview) {
  background: var(--n-color);
}

.editor-footer {
  padding: 8px 12px;
  border-top: 1px solid var(--n-border-color);
  background: var(--n-color);
  display: flex;
  justify-content: flex-end;
}

.char-count {
  font-size: 12px;
  color: var(--n-text-color-3);
}

.char-count.over-limit {
  color: #d03050;
}
</style>
