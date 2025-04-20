// Tags API service

import { apiClient } from "./api-client"
import type { Tag, CreateTagRequest, UpdateTagRequest } from "./api-types"

export async function listUserTags(token?: string): Promise<Tag[]> {
  return await apiClient.get<Tag[]>("/tags", token)
}

export async function getTagById(id: string, token: string): Promise<Tag> {
  return await apiClient.get<Tag>(`/tags/${id}`, token)
}

export async function createTag(request: Partial<CreateTagRequest>, token: string): Promise<Tag> {
  return await apiClient.post<Tag>("/tags", request, token)
}

export async function updateTagById(id: string, request: Partial<UpdateTagRequest>, token: string): Promise<Tag> {
  return await apiClient.patch<Tag>(`/tags/${id}`, request, token)
}

export async function deleteTagById(id: string, token: string): Promise<void> {
  await apiClient.delete<void>(`/tags/${id}`, token)
}