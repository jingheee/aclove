import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { baseFetch } from './client.js'
import { queryKeys } from './queryKeys.js'

const DEFAULT_PAGE_SIZE = 20

const SORT_OPTIONS = {
  NEWEST: 'newest',
  HOT: 'hot',
}

async function createPost(data) {
  console.log("Create Post Data:", data);
  return baseFetch('/posts', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

async function fetchPosts(params = {}) {
  console.log("Fetch Posts Params:", params);
  const queryParams = new URLSearchParams()
  if (params.category_id) queryParams.append('category_id', String(params.category_id))
  if (params.page) queryParams.append('page', String(params.page))
  if (params.page_size) queryParams.append('page_size', String(params.page_size))
  if (params.sort) queryParams.append('sort', params.sort)

  const query = queryParams.toString()
  return baseFetch(`/posts${query ? `?${query}` : ''}`)
}

async function fetchPost(id) {
  return baseFetch(`/posts/${id}`)
}

async function updatePost({ id, data }) {
  return baseFetch(`/posts/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

async function deletePost({ id, reason }) {
  const body = reason ? { reason } : {}
  return baseFetch(`/posts/${id}`, {
    method: 'DELETE',
    body: JSON.stringify(body),
  })
}

function usePostsQuery(filters = {}, options = {}) {
  const queryKey = queryKeys.posts.list(filters)

  return useQuery({
    queryKey,
    queryFn: () => fetchPosts(filters),
    staleTime: 1000 * 60 * 2,
    ...options,
  })
}

function usePostQuery(id, options = {}) {
  return useQuery({
    queryKey: queryKeys.posts.detail(id),
    queryFn: () => fetchPost(id),
    enabled: !!id,
    staleTime: 1000 * 60 * 5,
    ...options,
  })
}

function useCreatePostMutation(options = {}) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createPost,
    onSuccess: (data) => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.posts.lists(),
      })
      options.onSuccess?.(data)
    },
    onError: options.onError,
  })
}

function useUpdatePostMutation(options = {}) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: updatePost,
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.posts.detail(variables.id),
      })
      queryClient.invalidateQueries({
        queryKey: queryKeys.posts.lists(),
      })
      options.onSuccess?.(data, variables)
    },
    onError: options.onError,
  })
}

function useDeletePostMutation(options = {}) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: deletePost,
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.posts.lists(),
      })
      queryClient.removeQueries({
        queryKey: queryKeys.posts.detail(variables.id),
      })
      options.onSuccess?.(data, variables)
    },
    onError: options.onError,
  })
}

function usePrefetchPost() {
  const queryClient = useQueryClient()

  return (id) => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.posts.detail(id),
      queryFn: () => fetchPost(id),
      staleTime: 1000 * 60 * 5,
    })
  }
}

export {
  SORT_OPTIONS,
  DEFAULT_PAGE_SIZE,
  createPost,
  fetchPosts,
  fetchPost,
  updatePost,
  deletePost,
  usePostsQuery,
  usePostQuery,
  useCreatePostMutation,
  useUpdatePostMutation,
  useDeletePostMutation,
  usePrefetchPost,
}
