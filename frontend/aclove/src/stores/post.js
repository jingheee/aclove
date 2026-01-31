import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as postApi from '@/api/post.js'

export const usePostStore = defineStore('post', () => {
  const posts = ref([])
  const currentPost = ref(null)
  const loading = ref(false)
  const error = ref(null)
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(20)

  const hasMore = computed(() => {
    return posts.value.length < total.value
  })

  async function fetchPosts(params = {}) {
    loading.value = true
    error.value = null
    try {
      const response = await postApi.getPosts({
        page: currentPage.value,
        page_size: pageSize.value,
        ...params,
      })
      posts.value = response.items || []
      total.value = response.total || 0
      return response
    } catch (err) {
      error.value = err.message || '获取帖子列表失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchMorePosts(params = {}) {
    if (loading.value || !hasMore.value) return
    loading.value = true
    try {
      currentPage.value++
      const response = await postApi.getPosts({
        page: currentPage.value,
        page_size: pageSize.value,
        ...params,
      })
      posts.value.push(...(response.items || []))
      return response
    } catch (err) {
      error.value = err.message || '加载更多帖子失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchPostDetail(id) {
    loading.value = true
    error.value = null
    try {
      const response = await postApi.getPost(id)
      currentPost.value = response
      return response
    } catch (err) {
      error.value = err.message || '获取帖子详情失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createPost(data) {
    loading.value = true
    error.value = null
    try {
      const response = await postApi.createPost(data)
      posts.value.unshift(response)
      total.value++
      return response
    } catch (err) {
      error.value = err.message || '创建帖子失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updatePost(id, data) {
    loading.value = true
    error.value = null
    try {
      const response = await postApi.updatePost(id, data)
      const index = posts.value.findIndex(p => p.id === id)
      if (index !== -1) {
        posts.value[index] = { ...posts.value[index], ...response }
      }
      if (currentPost.value?.id === id) {
        currentPost.value = { ...currentPost.value, ...response }
      }
      return response
    } catch (err) {
      error.value = err.message || '更新帖子失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deletePost(id, reason) {
    loading.value = true
    error.value = null
    try {
      await postApi.deletePost(id, reason)
      posts.value = posts.value.filter(p => p.id !== id)
      total.value--
      if (currentPost.value?.id === id) {
        currentPost.value = null
      }
    } catch (err) {
      error.value = err.message || '删除帖子失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  function resetPosts() {
    posts.value = []
    currentPost.value = null
    total.value = 0
    currentPage.value = 1
    error.value = null
  }

  function setPage(page) {
    currentPage.value = page
  }

  return {
    posts,
    currentPost,
    loading,
    error,
    total,
    currentPage,
    pageSize,
    hasMore,
    fetchPosts,
    fetchMorePosts,
    fetchPostDetail,
    createPost,
    updatePost,
    deletePost,
    resetPosts,
    setPage,
  }
})
