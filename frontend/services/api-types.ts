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
  attachmentUrl?: string | null
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
}

export interface UpdateTodoRequest {
  title?: string
  description?: string | null
  status?: "pending" | "in-progress" | "completed"
  deadline?: string | null
  tagIds?: string[]
  // attachments are managed via separate endpoints, removed from here
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

export type FileUploadResponse = {
	fileId: string // Identifier used for deletion (e.g., GCS object path)
	fileName: string // Original filename
	fileUrl: string // Publicly accessible URL (e.g., signed URL)
	contentType: string // MIME type
	size: number // Size in bytes
}

export interface Error {
	code: number
	message: string
}
