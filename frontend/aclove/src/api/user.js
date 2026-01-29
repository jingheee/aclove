import { baseFetch } from "./client.js";

/**
 * 获取当前用户信息（验证 session 和 cookie）
 * @returns {Promise<UserInfo>}
 */
export async function fetchCurrentUser() {
  return baseFetch("/users/me", {
    credentials: "include",
  });
}

/**
 * @typedef {Object} UserInfo
 * @property {number} id - 用户ID
 * @property {string} status - 用户状态 (active/banned/cooldown)
 * @property {string} [status_reason] - 状态原因
 * @property {string} [banned_until] - 封禁截止时间
 * @property {string} [cooldown_until] - 冷却截止时间
 */
