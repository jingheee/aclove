const API_BASE_URL = "/api";

class ApiError extends Error {
  constructor(message, status, code, data) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.data = data;
  }
}

/**
 * 通用响应结构
 * @typedef {Object} ApiResponse
 * @property {number} code - 业务状态码: 0-成功, 非0-失败
 * @property {string} message - 提示信息
 * @property {any} data - 响应数据
 */

/**
 * 解析响应数据，提取 data 字段
 * @param {ApiResponse} response
 * @returns {any}
 */
function parseResponse(response) {
  if (response === null || response === undefined) {
    return null;
  }

  // 如果响应已经是数组或没有 code 字段，直接返回（兼容旧格式）
  if (Array.isArray(response) || response.code === undefined) {
    return response;
  }

  // 业务错误
  if (response.code !== 0) {
    throw new ApiError(
      response.message || "请求失败",
      200,
      response.code,
      response.data,
    );
  }

  return response.data;
}

async function baseFetch(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`;

  const config = {
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    credentials: "include",
    ...options,
  };

  const response = await fetch(url, config);

  const data = await response.json().catch(() => null);

  if (!response.ok) {
    const message = data?.message || `HTTP Error: ${response.status}`;
    const code = data?.code ?? -1;
    throw new ApiError(message, response.status, code, data);
  }

  return parseResponse(data);
}

export { baseFetch, API_BASE_URL, ApiError, parseResponse };
