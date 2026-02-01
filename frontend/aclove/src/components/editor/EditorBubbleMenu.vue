<script setup>
// 简化的气泡菜单组件 - 当前不使用 BubbleMenu 扩展
const props = defineProps({
  editor: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(["action"]);
</script>

<template>
  <!-- 空组件，气泡菜单功能暂时禁用 -->
</template>
