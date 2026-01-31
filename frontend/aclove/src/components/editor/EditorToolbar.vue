<script setup>
import { computed } from "vue";
import {
  AddCircleOutline,
  RemoveCircleOutline,
  CodeSlashOutline,
  TextOutline,
  ListOutline,
  ListCircleOutline,
  ChatbubbleOutline,
  TerminalOutline,
  ArrowUndoOutline,
  ArrowRedoOutline,
  LinkOutline,
  UnlinkOutline,
  ImageOutline,
  GridOutline,
  ReturnDownBackOutline,
  ReturnDownForwardOutline,
} from "@vicons/ionicons5";

const props = defineProps({
  editor: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(["action", "insert-image-url"]);

const isActive = computed(() => {
  if (!props.editor) return () => false;
  return (name, attrs = {}) => props.editor.isActive(name, attrs);
});

const canRun = computed(() => {
  if (!props.editor) return () => false;
  return (name) => props.editor.can()[name]();
});

const toolbarGroups = computed(() => [
  {
    name: "history",
    items: [
      { action: "undo", icon: ArrowUndoOutline, title: "撤销", disabled: !canRun.value("undo") },
      { action: "redo", icon: ArrowRedoOutline, title: "重做", disabled: !canRun.value("redo") },
    ],
  },
  {
    name: "marks",
    items: [
      { action: "bold", icon: AddCircleOutline, title: "粗体", active: isActive.value("bold") },
      { action: "italic", icon: RemoveCircleOutline, title: "斜体", active: isActive.value("italic") },
      { action: "underline", icon: TextOutline, title: "下划线", active: isActive.value("underline") },
      { action: "strike", icon: CodeSlashOutline, title: "删除线", active: isActive.value("strike") },
      { action: "code", icon: TerminalOutline, title: "行内代码", active: isActive.value("code") },
    ],
  },
  {
    name: "headings",
    items: [
      { action: "paragraph", icon: TextOutline, title: "正文", active: isActive.value("paragraph") },
      { action: "h1", icon: TextOutline, title: "标题 1", active: isActive.value("heading", { level: 1 }) },
      { action: "h2", icon: TextOutline, title: "标题 2", active: isActive.value("heading", { level: 2 }) },
      { action: "h3", icon: TextOutline, title: "标题 3", active: isActive.value("heading", { level: 3 }) },
    ],
  },
  {
    name: "lists",
    items: [
      { action: "bulletList", icon: ListOutline, title: "无序列表", active: isActive.value("bulletList") },
      { action: "orderedList", icon: ListCircleOutline, title: "有序列表", active: isActive.value("orderedList") },
    ],
  },
  {
    name: "align",
    items: [
      { action: "alignLeft", icon: ReturnDownBackOutline, title: "左对齐", active: isActive.value({ textAlign: "left" }) },
      { action: "alignCenter", icon: GridOutline, title: "居中", active: isActive.value({ textAlign: "center" }) },
      { action: "alignRight", icon: ReturnDownForwardOutline, title: "右对齐", active: isActive.value({ textAlign: "right" }) },
    ],
  },
  {
    name: "insert",
    items: [
      { action: "blockquote", icon: ChatbubbleOutline, title: "引用", active: isActive.value("blockquote") },
      { action: "codeBlock", icon: TerminalOutline, title: "代码块", active: isActive.value("codeBlock") },
      { action: "horizontalRule", icon: RemoveCircleOutline, title: "分隔线" },
    ],
  },
  {
    name: "link",
    items: [
      { action: "link", icon: LinkOutline, title: "插入链接", active: isActive.value("link") },
      { action: "unsetLink", icon: UnlinkOutline, title: "移除链接", disabled: !isActive.value("link") },
    ],
  },
  {
    name: "image",
    items: [
      { action: "image", icon: ImageOutline, title: "上传图片" },
    ],
  },
]);

function handleAction(action) {
  emit("action", action);
}

function handleInsertImageUrl() {
  emit("insert-image-url");
}
</script>

<template>
  <div class="editor-toolbar">
    <div class="toolbar-scroll">
      <div v-for="group in toolbarGroups" :key="group.name" class="toolbar-group">
        <button
          v-for="item in group.items"
          :key="item.action"
          class="toolbar-btn"
          :class="{
            'is-active': item.active,
            'is-disabled': item.disabled,
          }"
          :title="item.title"
          :disabled="item.disabled"
          @click="handleAction(item.action)"
        >
          <component :is="item.icon" class="toolbar-icon" />
        </button>
      </div>
    </div>

    <div class="toolbar-extra">
      <button class="toolbar-btn" title="URL 插入图片" @click="handleInsertImageUrl">
        <LinkOutline class="toolbar-icon" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.editor-toolbar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  background: var(--n-action-color);
  border-bottom: 1px solid var(--n-border-color);
  flex-wrap: wrap;
}

.toolbar-scroll {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
  flex: 1;
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 0 4px;
  border-right: 1px solid var(--n-border-color);
}

.toolbar-group:last-of-type {
  border-right: none;
}

.toolbar-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--n-text-color-2);
  cursor: pointer;
  transition: all 0.15s ease;
}

.toolbar-btn:hover:not(:disabled) {
  background: var(--n-button-color-2-hover);
  color: var(--n-text-color);
}

.toolbar-btn.is-active {
  background: var(--n-primary-color);
  color: white;
}

.toolbar-btn.is-disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.toolbar-btn.is-danger {
  color: var(--n-error-color);
}

.toolbar-btn.is-danger:hover {
  background: var(--n-error-color-hover);
  color: white;
}

.toolbar-icon {
  width: 18px;
  height: 18px;
}

.toolbar-extra {
  display: flex;
  align-items: center;
  gap: 4px;
  padding-left: 8px;
  border-left: 1px solid var(--n-border-color);
  margin-left: auto;
}

@media (max-width: 768px) {
  .editor-toolbar {
    padding: 6px 8px;
  }

  .toolbar-btn {
    width: 28px;
    height: 28px;
  }

  .toolbar-icon {
    width: 16px;
    height: 16px;
  }
}
</style>
