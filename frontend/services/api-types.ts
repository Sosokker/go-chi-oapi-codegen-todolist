// Generated types based on the OpenAPI spec

export interface User {
  id: string
  username: string
  email: string
  emailVerified: boolean
  createdAt: string
  updatedAt: string
}

export interface SignupRequest {
  username: string
  email: string
  password: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  accessToken: string
  tokenType: string
}

export interface UpdateUserRequest {
  username?: string
}

export interface Tag {
  id: string
  userId: string
  name: string
  color?: string | null
  icon?: string | null
  createdAt: string
  updatedAt: string
}

export interface CreateTagRequest {
  name: string
  color?: string | null
  icon?: string | null
}

export interface UpdateTagRequest {
  name?: string
  color?: string | null
  icon?: string | null
}

export interface Todo {
  id: string
  userId: string
  title: string
  description?: string | null
  status: "pending" | "in-progress" | "completed"
  deadline?: string | null
  tagIds: string[]
  attachments: string[]
  image?: string | null
  subtasks: Subtask[]
  createdAt: string
  updatedAt: string
}

export interface CreateTodoRequest {
  title: string
  description?: string | null
  status?: "pending" | "in-progress" | "completed"
  deadline?: string | null
  tagIds?: string[]
  image?: string | null
}

export interface UpdateTodoRequest {
  title?: string
  description?: string | null
  status?: "pending" | "in-progress" | "completed"
  deadline?: string | null
  tagIds?: string[]
  attachments?: string[]
  image?: string | null
}

export interface Subtask {
  id: string
  todoId: string
  description: string
  completed: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateSubtaskRequest {
  description: string
}

export interface UpdateSubtaskRequest {
  description?: string
  completed?: boolean
}

export interface FileUploadResponse {
  fileId: string
  fileName: string
  fileUrl: string
  contentType: string
  size: number
}

export interface Error {
  code: number
  message: string
}
