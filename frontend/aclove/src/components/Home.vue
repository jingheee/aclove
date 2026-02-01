<script setup>
import { h } from "vue";
import { useRouter } from "vue-router";
import {
  NCard,
  NSpace,
  NButton,
  NIcon,
  NGradientText,
  NDivider,
  NText,
} from "naive-ui";
import { HeartOutline, SparklesOutline } from "@vicons/ionicons5";

const router = useRouter();

const emit = defineEmits(["create-post"]);

function handleBrowsePosts() {
  router.push({ name: "category", params: { categoryId: "all" } });
}

function handleCreatePost() {
  emit("create-post");
}
</script>

<template>
  <div class="home-container">
    <NCard class="welcome-card" :bordered="false">
      <div class="welcome-content">
        <NSpace vertical align="center" :size="24">
          <div class="icon-wrapper">
            <NIcon size="80" :depth="1">
              <HeartOutline />
            </NIcon>
          </div>

          <NGradientText
            :size="48"
            :font-size="48"
            :weight="800"
            type="linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
          >
            欢迎来到aclove匿名版
          </NGradientText>

          <NDivider />

          <NText :depth="2" class="subtitle">
            <NSpace vertical align="center" :size="12">
              <span>你所热爱的就是你的生活</span>
              <NIcon :depth="3" size="24">
                <SparklesOutline />
              </NIcon>
            </NSpace>
          </NText>

          <NSpace :size="16" style="margin-top: 24px">
            <NButton type="primary" size="large" @click="handleBrowsePosts">
              浏览帖子
            </NButton>
            <NButton size="large" @click="handleCreatePost"> 发布帖子 </NButton>
          </NSpace>
        </NSpace>
      </div>
    </NCard>
  </div>
</template>

<style scoped>
.home-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 40px 24px;
}

.welcome-card {
  max-width: 800px;
  width: 100%;
  border-radius: 24px;
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.1);
  transition:
    transform 0.4s ease,
    box-shadow 0.4s ease;
}

.welcome-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 20px 64px rgba(0, 0, 0, 0.16);
}

.dark .welcome-card {
  background: rgba(30, 30, 46, 0.85);
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.4);
}

.welcome-content {
  padding: 64px 48px;
  text-align: center;
}

.icon-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 128px;
  height: 128px;
  border-radius: 50%;
  background: linear-gradient(
    135deg,
    rgba(102, 126, 234, 0.12) 0%,
    rgba(118, 75, 162, 0.12) 100%
  );
  margin: 0 auto;
}

.dark .icon-wrapper {
  background: linear-gradient(
    135deg,
    rgba(102, 126, 234, 0.2) 0%,
    rgba(118, 75, 162, 0.2) 100%
  );
}

.subtitle {
  font-size: 28px;
  line-height: 1.8;
  letter-spacing: 3px;
}
</style>
