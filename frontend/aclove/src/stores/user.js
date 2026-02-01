import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { fetchCurrentUser } from "@/api/user.js";

/**
 * 用户状态枚举
 */
export const UserStatus = {
  ACTIVE: "active",
  BANNED: "banned",
  COOLDOWN: "cooldown",
};

/**
 * 用户 Store - 管理匿名用户状态和 Session
 */
export const useUserStore = defineStore("user", () => {
  // State
  const user = ref(null);
  const isLoading = ref(false);
  const error = ref(null);

  // Getters
  const isLoggedIn = computed(() => !!user.value?.id);
  const isBanned = computed(() => user.value?.status === UserStatus.BANNED);
  const isCooldown = computed(() => user.value?.status === UserStatus.COOLDOWN);
  const isActive = computed(() => user.value?.status === UserStatus.ACTIVE);
  const userId = computed(() => user.value?.id);

  /**
   * 初始化用户 Session（页面加载时调用）
   * 后端会自动处理 cookie 和 session 验证
   */
  async function initSession() {
    isLoading.value = true;
    error.value = null;

    try {
      const userInfo = await fetchCurrentUser();
      user.value = userInfo;
      return userInfo;
    } catch (err) {
      error.value = err.message || "获取用户信息失败";
      throw err;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * 清除用户状态
   */
  function clearUser() {
    user.value = null;
    error.value = null;
  }

  return {
    // State
    user,
    isLoading,
    error,
    // Getters
    isLoggedIn,
    isBanned,
    isCooldown,
    isActive,
    userId,
    // Actions
    initSession,
    clearUser,
  };
});
