import { baseFetch } from './client.js'

export async function createPost(data) {
  return baseFetch('/posts', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function getPosts(params = {}) {
  const queryParams = new URLSearchParams()
  if (params.category_id) queryParams.append('category_id', params.category_id)
  if (params.page) queryParams.append('page', params.page)
  if (params.page_size) queryParams.append('page_size', params.page_size)
  if (params.sort) queryParams.append('sort', params.sort)

  const query = queryParams.toString()
  return baseFetch(`/posts${query ? `?${query}` : ''}`)
}

export async function getPost(id) {
  return baseFetch(`/posts/${id}`)
}

export async function updatePost(id, data) {
  return baseFetch(`/posts/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export async function deletePost(id, reason) {
  const body = reason ? { reason } : {}
  return baseFetch(`/posts/${id}`, {
    method: 'DELETE',
    body: JSON.stringify(body),
  })
}
