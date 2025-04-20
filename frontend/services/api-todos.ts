// Todo API service

import { apiClient } from "./api-client"
import type { Todo, CreateTodoRequest, UpdateTodoRequest } from "./api-types"

export async function listTodos(
  params?: { status?: string; tagId?: string },
  token?: string
): Promise<Todo[]> {
  // Build query string for params
  const queryParams = new URLSearchParams()
  if (params?.status) queryParams.append("status", params.status)
  if (params?.tagId) queryParams.append("tagId", params.tagId)

  const queryString = queryParams.toString() ? `?${queryParams.toString()}` : ""

  return await apiClient.get<Todo[]>(`/todos${queryString}`, token)
}

export async function getTodoById(id: string, token: string): Promise<Todo> {
  return await apiClient.get<Todo>(`/todos/${id}`, token)
}

export async function createTodo(
  request: Partial<CreateTodoRequest>,
  token: string
): Promise<Todo> {
  return await apiClient.post<Todo>("/todos", request, token)
}

export async function updateTodoById(
  id: string,
  request: Partial<UpdateTodoRequest>,
  token: string
): Promise<Todo> {
  return await apiClient.patch<Todo>(`/todos/${id}`, request, token)
}

export async function deleteTodoById(id: string, token: string): Promise<void> {
  await apiClient.delete<void>(`/todos/${id}`, token)
}