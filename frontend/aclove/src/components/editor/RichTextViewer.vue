<script setup>
import { computed } from "vue";
import DOMPurify from "dompurify";

const props = defineProps({
  content: {
    type: String,
    default: "",
  },
  maxLines: {
    type: Number,
    default: 0, // 0 means no limit
  },
});

const sanitizedContent = computed(() => {
  if (!props.content) return "";

  const config = {
    ALLOWED_TAGS: [
      "p", "br", "strong", "b", "em", "i", "u", "s", "del", "strike",
      "h1", "h2", "h3", "h4", "h5", "h6",
      "ul", "ol", "li",
      "blockquote", "code", "pre",
      "a", "img",
      "table", "thead", "tbody", "tr", "th", "td",
      "hr",
    ],
    ALLOWED_ATTR: [
      "href", "title", "target", "rel",
      "src", "alt", "title", "width", "height",
      "class", "style",
    ],
    ALLOW_DATA_ATTR: false,
  };

  return DOMPurify.sanitize(props.content, config);
});

const containerClasses = computed(() => {
  const classes = ["rich-text-viewer"];
  if (props.maxLines > 0) {
    classes.push("is-truncated");
  }
  return classes.join(" ");
});

const containerStyle = computed(() => {
  if (props.maxLines > 0) {
    return {
      "-webkit-line-clamp": props.maxLines,
    };
  }
  return {};
});
</script>

<template>
  <div
    :class="containerClasses"
    :style="containerStyle"
    v-html="sanitizedContent"
  />
</template>

<style scoped>
.rich-text-viewer {
  line-height: 1.7;
  color: var(--n-text-color);
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.rich-text-viewer.is-truncated {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.rich-text-viewer :deep(p) {
  margin: 0.75em 0;
}

.rich-text-viewer :deep(p:first-child) {
  margin-top: 0;
}

.rich-text-viewer :deep(p:last-child) {
  margin-bottom: 0;
}

.rich-text-viewer :deep(h1) {
  font-size: 1.5em;
  font-weight: 600;
  margin: 1em 0 0.5em;
  line-height: 1.3;
}

.rich-text-viewer :deep(h2) {
  font-size: 1.3em;
  font-weight: 600;
  margin: 1em 0 0.5em;
  line-height: 1.3;
}

.rich-text-viewer :deep(h3) {
  font-size: 1.15em;
  font-weight: 600;
  margin: 1em 0 0.5em;
  line-height: 1.3;
}

.rich-text-viewer :deep(ul),
.rich-text-viewer :deep(ol) {
  padding-left: 1.5em;
  margin: 0.75em 0;
}

.rich-text-viewer :deep(li) {
  margin: 0.25em 0;
}

.rich-text-viewer :deep(blockquote) {
  border-left: 4px solid var(--n-primary-color);
  padding-left: 1em;
  margin: 1em 0;
  color: var(--n-text-color-3);
  font-style: italic;
}

.rich-text-viewer :deep(pre) {
  background: var(--n-code-color);
  border-radius: 8px;
  padding: 1em;
  margin: 1em 0;
  overflow-x: auto;
}

.rich-text-viewer :deep(pre code) {
  background: none;
  padding: 0;
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 0.9em;
  line-height: 1.5;
}

.rich-text-viewer :deep(code) {
  background: var(--n-code-color);
  padding: 0.2em 0.4em;
  border-radius: 4px;
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 0.9em;
}

.rich-text-viewer :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
  margin: 1em 0;
}

.rich-text-viewer :deep(a) {
  color: var(--n-primary-color);
  text-decoration: none;
  border-bottom: 1px solid transparent;
  transition: border-color 0.2s;
}

.rich-text-viewer :deep(a:hover) {
  border-bottom-color: var(--n-primary-color);
}

.rich-text-viewer :deep(hr) {
  border: none;
  border-top: 2px solid var(--n-divider-color);
  margin: 2em 0;
}

.rich-text-viewer :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1em 0;
  font-size: 0.95em;
}

.rich-text-viewer :deep(th),
.rich-text-viewer :deep(td) {
  border: 1px solid var(--n-border-color);
  padding: 0.6em 0.8em;
  text-align: left;
}

.rich-text-viewer :deep(th) {
  background: var(--n-table-header-color);
  font-weight: 600;
}

.rich-text-viewer :deep(tr:nth-child(even)) {
  background: var(--n-action-color);
}
</style>
