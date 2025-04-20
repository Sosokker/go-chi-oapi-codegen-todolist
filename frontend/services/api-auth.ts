import { apiClient } from "./api-client"
import type { User, SignupRequest, LoginRequest, LoginResponse, UpdateUserRequest } from "./api-types"

export async function signupUserApi(request: SignupRequest): Promise<User> {
  return await apiClient.post<User>("/auth/signup", request)
}

export async function loginUserApi(request: LoginRequest): Promise<LoginResponse> {
  return await apiClient.post<LoginResponse>("/auth/login", request)
}

export async function getCurrentUser(token: string): Promise<User> {
  return await apiClient.get<User>("/users/me", token)
}

export async function updateUserApi(request: UpdateUserRequest, token: string): Promise<User> {
  return await apiClient.patch<User>("/users/me", request, token)
}

export async function validateToken(token: string): Promise<boolean> {
  return await apiClient.get<boolean>("/users/me", token).then(() => true).catch(() => false)
}

export async function logoutUser(token: string): Promise<void> {
  await apiClient.post<void>("/auth/logout", {}, token)
}