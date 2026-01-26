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
    <div class="welcome-section">
      <NCard bordered="false" class="welcome-card">
        <template #header>
          <NSpace align="center" justify="space-between">
            <NSpace align="center">
              <NIcon size="24" color="#18a058">
                <HomeOutline />
              </NIcon>
              <NText style="font-size: 18px; font-weight: 600;">
                欢迎使用 ACLOVE
              </NText>
            </NSpace>
            <NTag type="success" size="small" round>系统运行正常</NTag>
          </NSpace>
        </template>
        <div class="welcome-content">
          <NText depth="1" style="font-size: 15px; line-height: 1.8;">
            欢迎来到 ACLOVE 管理平台。这是一个现代化的内容管理系统，提供了完整的分类管理、数据统计和用户管理功能。
          </NText>
          <NDivider />
          <div class="quick-actions">
            <NText
              strong
              depth="1"
              style="margin-bottom: 12px; display: block"
            >
              快捷操作
            </NText>
            <NSpace :size="12">
              <NButton
                v-for="action in quickActions"
                :key="action.label"
                :type="action.type"
                size="small"
              >
                <template #icon>
                  <NIcon>
                    <Component :is="renderIcon(action.icon)" />
                  </NIcon>
                </template>
                {{ action.label }}
              </NButton>
            </NSpace>
          </div>
        </div>
      </NCard>
    </div>

    <NDivider />

    <div class="stats-section">
      <NGrid :x-gap="16" :y-gap="16" :cols="4" responsive="screen" :item-responsive="true">
        <NGridItem v-for="(stat, index) in statCards" :key="index" :span="4 / 4">
          <NCard bordered="false" class="stat-card">
            <NSpace align="center" justify="space-between">
              <div>
                <NText depth="2" style="font-size: 13px">{{ stat.title }}</NText>
                <div style="margin-top: 8px">
                  <NText strong style="font-size: 28px; font-weight: 600">
                    {{ stat.value }}
                  </NText>
                </div>
              </div>
              <div class="stat-icon">
                <NIcon size="32" :color="stat.color">
                  <Component :is="renderIcon(stat.icon)" />
                </NIcon>
              </div>
            </NSpace>
          </NCard>
        </NGridItem>
      </NGrid>
    </div>

    <NDivider />

    <div class="categories-section">
      <NCard bordered="false" title="分类概览">
        <template #header-extra>
          <NButton text type="primary" @click="emit('refresh-categories')">
            <template #icon>
              <NIcon><RefreshOutline /></NIcon>
            </template>
            刷新
          </NButton>
        </template>
        <NSpin :show="loading">
          <div v-if="categories.length > 0" class="category-list">
            <div
              v-for="(category, index) in categories"
              :key="category.id || index"
              class="category-item"
            >
              <NSpace align="center">
                <NIcon size="20" color="#18a058">
                  <LayersOutline />
                </NIcon>
                <NText strong>{{ category.name || "未命名分类" }}</NText>
              </NSpace>
              <NSpace :size="8">
                <NButton size="small" quaternary>
                  <template #icon>
                    <NIcon><CreateOutline /></NIcon>
                  </template>
                </NButton>
                <NButton size="small" quaternary type="error">
                  <template #icon>
                    <NIcon><TrashOutline /></NIcon>
                  </template>
                </NButton>
              </NSpace>
            </div>
          </div>
          <NEmpty v-else description="暂无分类数据" style="padding: 40px 0">
            <template #extra>
              <NText depth="3" style="font-size: 13px"
                >点击"创建分类"添加您的第一个分类</NText
              >
            </template>
          </NEmpty>
        </NSpin>
      </NCard>
    </div>
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
