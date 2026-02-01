<script setup>
import { computed, watch, ref } from "vue";
import { useEditor, EditorContent } from "@tiptap/vue-3";
import StarterKit from "@tiptap/starter-kit";
import Image from "@tiptap/extension-image";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import TextAlign from "@tiptap/extension-text-align";
import Underline from "@tiptap/extension-underline";
import CodeBlockLowlight from "@tiptap/extension-code-block-lowlight";
import { all, createLowlight } from "lowlight";
import EditorToolbar from "./EditorToolbar.vue";
import EditorBubbleMenu from "./EditorBubbleMenu.vue";

const lowlight = createLowlight(all);

const props = defineProps({
  modelValue: {
    type: String,
    default: "",
  },
  placeholder: {
    type: String,
    default: "开始输入内容...",
  },
  maxLength: {
    type: Number,
    default: 10000,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  uploadEndpoint: {
    type: String,
    default: "/api/upload",
  },
});

const emit = defineEmits(["update:modelValue", "change", "image-upload"]);

const isUploading = ref(false);
const uploadError = ref("");
const fileInputRef = ref(null);

const editor = useEditor({
  extensions: [
    StarterKit.configure({
      codeBlock: false,
    }),
    Image.configure({
      allowBase64: false,
      inline: true,
    }),
    Link.configure({
      openOnClick: false,
      HTMLAttributes: {
        rel: "noopener noreferrer nofollow",
        target: "_blank",
      },
    }),
    Placeholder.configure({
      placeholder: props.placeholder,
    }),
    TextAlign.configure({
      types: ["heading", "paragraph"],
    }),
    Underline,
    CodeBlockLowlight.configure({
      lowlight,
    }),
  ],
  content: props.modelValue,
  editable: !props.disabled,
  onUpdate: ({ editor }) => {
    const html = editor.getHTML();
    if (html.length <= props.maxLength) {
      emit("update:modelValue", html);
      emit("change", html);
      uploadError.value = "";
    } else {
      uploadError.value = `内容超出最大长度限制 (${props.maxLength} 字符)`;
    }
  },
  editorProps: {
    attributes: {
      class: "rich-text-content",
    },
    handlePaste: (view, event) => {
      const items = event.clipboardData?.items;
      if (!items) return false;

      for (const item of items) {
        if (item.type.startsWith("image/")) {
          const file = item.getAsFile();
          if (file) {
            handleImageUpload(file);
            return true;
          }
        }
      }
      return false;
    },
    handleDrop: (view, event) => {
      const files = event.dataTransfer?.files;
      if (!files) return false;

      for (const file of files) {
        if (file.type.startsWith("image/")) {
          handleImageUpload(file);
          return true;
        }
      }
      return false;
    },
  },
});

const characterCount = computed(() => {
  return editor.value?.storage.characterCount?.characters() || 0;
});

const wordCount = computed(() => {
  return editor.value?.storage.characterCount?.words() || 0;
});

watch(
  () => props.modelValue,
  (newValue) => {
    if (editor.value && newValue !== editor.value.getHTML()) {
      editor.value.commands.setContent(newValue, false);
    }
  },
);

watch(
  () => props.disabled,
  (newValue) => {
    editor.value?.setEditable(!newValue);
  },
);

function handleToolbarAction(action, payload = null) {
  if (!editor.value) return;

  const commands = {
    bold: () => editor.value.chain().focus().toggleBold().run(),
    italic: () => editor.value.chain().focus().toggleItalic().run(),
    underline: () => editor.value.chain().focus().toggleUnderline().run(),
    strike: () => editor.value.chain().focus().toggleStrike().run(),
    code: () => editor.value.chain().focus().toggleCode().run(),
    paragraph: () => editor.value.chain().focus().setParagraph().run(),
    h1: () => editor.value.chain().focus().toggleHeading({ level: 1 }).run(),
    h2: () => editor.value.chain().focus().toggleHeading({ level: 2 }).run(),
    h3: () => editor.value.chain().focus().toggleHeading({ level: 3 }).run(),
    bulletList: () => editor.value.chain().focus().toggleBulletList().run(),
    orderedList: () => editor.value.chain().focus().toggleOrderedList().run(),
    taskList: () => editor.value.chain().focus().toggleTaskList().run(),
    blockquote: () => editor.value.chain().focus().toggleBlockquote().run(),
    codeBlock: () => editor.value.chain().focus().toggleCodeBlock().run(),
    horizontalRule: () => editor.value.chain().focus().setHorizontalRule().run(),
    hardBreak: () => editor.value.chain().focus().setHardBreak().run(),
    undo: () => editor.value.chain().focus().undo().run(),
    redo: () => editor.value.chain().focus().redo().run(),
    alignLeft: () => editor.value.chain().focus().setTextAlign("left").run(),
    alignCenter: () => editor.value.chain().focus().setTextAlign("center").run(),
    alignRight: () => editor.value.chain().focus().setTextAlign("right").run(),
    alignJustify: () => editor.value.chain().focus().setTextAlign("justify").run(),
    link: () => {
      const previousUrl = editor.value.getAttributes("link").href;
      const url = payload || window.prompt("输入链接地址", previousUrl);
      if (url === null) return;
      if (url === "") {
        editor.value.chain().focus().extendMarkRange("link").unsetLink().run();
      } else {
        editor.value.chain().focus().extendMarkRange("link").setLink({ href: url }).run();
      }
    },
    unsetLink: () => editor.value.chain().focus().unsetLink().run(),
    image: () => {
      fileInputRef.value?.click();
    },
    clear: () => editor.value.chain().focus().clearNodes().unsetAllMarks().run(),
  };

  if (commands[action]) {
    commands[action]();
  }
}

async function handleImageUpload(file) {
  if (!file.type.startsWith("image/")) {
    uploadError.value = "只能上传图片文件";
    return;
  }

  const maxSize = 10 * 1024 * 1024; // 10MB
  if (file.size > maxSize) {
    uploadError.value = "图片大小不能超过 10MB";
    return;
  }

  isUploading.value = true;
  uploadError.value = "";

  try {
    emit("image-upload", { file, status: "uploading" });

    const formData = new FormData();
    formData.append("file", file);

    const response = await fetch(props.uploadEndpoint, {
      method: "POST",
      body: formData,
      credentials: "include",
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || "上传失败");
    }

    const data = await response.json();

    if (!data.url) {
      throw new Error("上传响应缺少图片 URL");
    }

    editor.value
      .chain()
      .focus()
      .setImage({
        src: data.url,
        alt: file.name,
        title: file.name,
      })
      .run();

    emit("image-upload", { file, status: "success", url: data.url });
  } catch (err) {
    uploadError.value = err.message || "图片上传失败";
    emit("image-upload", { file, status: "error", error: err.message });
  } finally {
    isUploading.value = false;
  }
}

function handleFileInputChange(event) {
  const file = event.target.files?.[0];
  if (file) {
    handleImageUpload(file);
  }
  event.target.value = "";
}

function insertImageByUrl() {
  const url = window.prompt("输入图片 URL");
  if (url) {
    editor.value?.chain().focus().setImage({ src: url }).run();
  }
}
</script>

<template>
  <div class="rich-text-editor" :class="{ 'is-disabled': disabled, 'is-uploading': isUploading }">
    <input
      ref="fileInputRef"
      type="file"
      accept="image/*"
      style="display: none"
      @change="handleFileInputChange"
    />

    <EditorToolbar
      :editor="editor"
      @action="handleToolbarAction"
      @insert-image-url="insertImageByUrl"
    />

    <div class="editor-wrapper">
      <EditorBubbleMenu :editor="editor" @action="handleToolbarAction" />

      <EditorContent :editor="editor" class="editor-content" />

      <div v-if="isUploading" class="upload-overlay">
        <div class="upload-spinner"></div>
        <span>上传中...</span>
      </div>
    </div>

    <div class="editor-footer">
      <div v-if="uploadError" class="error-message">
        {{ uploadError }}
      </div>
      <div class="character-count">
        <span :class="{ 'is-over-limit': characterCount > maxLength }">
          {{ characterCount }} / {{ maxLength }}
        </span>
        <span class="word-count">{{ wordCount }} 词</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.rich-text-editor {
  border: 1px solid var(--n-border-color);
  border-radius: 12px;
  overflow: hidden;
  background: var(--n-card-color);
  transition: all 0.2s ease;
}

.rich-text-editor:focus-within {
  border-color: var(--n-primary-color);
  box-shadow: 0 0 0 2px var(--n-primary-color-hover);
}

.rich-text-editor.is-disabled {
  opacity: 0.6;
  pointer-events: none;
}

.rich-text-editor.is-uploading .editor-content {
  opacity: 0.7;
}

.editor-wrapper {
  position: relative;
  min-height: 200px;
  max-height: 600px;
  overflow-y: auto;
}

.editor-content {
  padding: 16px 20px;
}

.editor-content :deep(.rich-text-content) {
  outline: none;
  min-height: 160px;
  line-height: 1.7;
  color: var(--n-text-color);
}

.editor-content :deep(.rich-text-content p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  float: left;
  color: var(--n-placeholder-color);
  pointer-events: none;
  height: 0;
}

.editor-content :deep(.rich-text-content p) {
  margin: 0.75em 0;
}

.editor-content :deep(.rich-text-content h1) {
  font-size: 1.75em;
  font-weight: 600;
  margin: 1em 0 0.5em;
  line-height: 1.3;
}

.editor-content :deep(.rich-text-content h2) {
  font-size: 1.5em;
  font-weight: 600;
  margin: 1em 0 0.5em;
  line-height: 1.3;
}

.editor-content :deep(.rich-text-content h3) {
  font-size: 1.25em;
  font-weight: 600;
  margin: 1em 0 0.5em;
  line-height: 1.3;
}

.editor-content :deep(.rich-text-content ul),
.editor-content :deep(.rich-text-content ol) {
  padding-left: 1.5em;
  margin: 0.75em 0;
}

.editor-content :deep(.rich-text-content li) {
  margin: 0.25em 0;
}

.editor-content :deep(.rich-text-content blockquote) {
  border-left: 4px solid var(--n-primary-color);
  padding-left: 1em;
  margin: 1em 0;
  color: var(--n-text-color-3);
  font-style: italic;
}

.editor-content :deep(.rich-text-content pre) {
  background: var(--n-code-color);
  border-radius: 8px;
  padding: 1em;
  margin: 1em 0;
  overflow-x: auto;
}

.editor-content :deep(.rich-text-content pre code) {
  background: none;
  padding: 0;
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 0.9em;
  line-height: 1.5;
}

.editor-content :deep(.rich-text-content code) {
  background: var(--n-code-color);
  padding: 0.2em 0.4em;
  border-radius: 4px;
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 0.9em;
}

.editor-content :deep(.rich-text-content img) {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
  margin: 1em 0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.editor-content :deep(.rich-text-content img.ProseMirror-selectednode) {
  outline: 2px solid var(--n-primary-color);
}

.editor-content :deep(.rich-text-content a) {
  color: var(--n-primary-color);
  text-decoration: none;
  border-bottom: 1px solid transparent;
  transition: border-color 0.2s;
}

.editor-content :deep(.rich-text-content a:hover) {
  border-bottom-color: var(--n-primary-color);
}

.editor-content :deep(.rich-text-content hr) {
  border: none;
  border-top: 2px solid var(--n-divider-color);
  margin: 2em 0;
}

.editor-content :deep(.rich-text-content table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1em 0;
}

.editor-content :deep(.rich-text-content th),
.editor-content :deep(.rich-text-content td) {
  border: 1px solid var(--n-border-color);
  padding: 0.75em;
  text-align: left;
}

.editor-content :deep(.rich-text-content th) {
  background: var(--n-table-header-color);
  font-weight: 600;
}

.editor-content :deep(.rich-text-content .selectedCell:after) {
  background: var(--n-primary-color-hover);
}

.editor-content :deep(.rich-text-content s),
.editor-content :deep(.rich-text-content del) {
  text-decoration: line-through;
}

.editor-content :deep(.rich-text-content u) {
  text-decoration: underline;
}

.upload-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(4px);
  gap: 12px;
  color: var(--n-primary-color);
  font-weight: 500;
}

.upload-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--n-primary-color-hover);
  border-top-color: var(--n-primary-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.editor-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  border-top: 1px solid var(--n-border-color);
  background: var(--n-action-color);
}

.error-message {
  color: var(--n-error-color);
  font-size: 0.85em;
}

.character-count {
  display: flex;
  gap: 12px;
  font-size: 0.85em;
  color: var(--n-text-color-3);
}

.character-count .is-over-limit {
  color: var(--n-error-color);
  font-weight: 600;
}

.word-count {
  opacity: 0.7;
}
</style>
