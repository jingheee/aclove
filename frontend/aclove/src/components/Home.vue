<script setup>
import { h, computed } from "vue";
import {
  NCard,
  NSpace,
  NText,
  NDivider,
  NGrid,
  NGridItem,
  NButton,
  NIcon,
  NSpin,
  NEmpty,
  NTag,
} from "naive-ui";
import {
  HomeOutline,
  AnalyticsOutline,
  BookOutline,
  PeopleOutline,
  LayersOutline,
  CreateOutline,
  TrashOutline,
  RefreshOutline,
  AddCircleOutline,
} from "@vicons/ionicons5";

const props = defineProps({
  categories: {
    type: Array,
    default: () => [],
  },
  loading: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["refresh-categories"]);

const statCards = computed(() => [
  {
    title: "分类总数",
    value: props.categories.length,
    icon: LayersOutline,
    color: "#18a058",
  },
  {
    title: "用户数量",
    value: 0,
    icon: PeopleOutline,
    color: "#2080f0",
  },
  {
    title: "数据条目",
    value: 0,
    icon: BookOutline,
    color: "#f0a020",
  },
  {
    title: "访问量",
    value: 0,
    icon: AnalyticsOutline,
    color: "#d03050",
  },
]);

const quickActions = [
  { label: "新建分类", icon: AddCircleOutline, type: "primary" },
  { label: "导入数据", icon: CreateOutline, type: "default" },
  { label: "导出报表", icon: AnalyticsOutline, type: "default" },
];

function renderIcon(icon) {
  return () => h(NIcon, null, { default: () => h(icon) });
}
</script>

<template>
  <div class="home-container">

    <NDivider />
    <NDivider />
  </div>
</template>

<style scoped>
.home-container {
  max-width: 1400px;
  margin: 0 auto;
}

.welcome-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
}

.dark .welcome-card {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.2);
}

.welcome-content {
  padding: 8px 0;
}

.quick-actions {
  margin-top: 16px;
}
</style>
