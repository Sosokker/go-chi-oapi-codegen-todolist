// Base API client for making requests to the backend

// Helper for simulating network delay in development
const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

// Get the API base URL from environment variable or use default
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080/api/v1"

// Request options with authentication token if available
const getRequestOptions = (token?: string | null) => {
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  }

  if (token) {
    headers["Authorization"] = `Bearer ${token}`
  }

  return { headers }
}

// Generic fetch function with error handling
export async function apiFetch<T>(
  endpoint: string,
  options: RequestInit = {},
  token?: string | null,
  simulateDelay = process.env.NODE_ENV === "development" ? 300 : 0, // Only simulate delay in development
): Promise<T> {
  if (simulateDelay > 0) {
    await delay(simulateDelay)
  }

  const url = `${API_BASE_URL}${endpoint.startsWith("/") ? endpoint : `/${endpoint}`}`
  const requestOptions = {
    ...options,
    headers: {
      ...getRequestOptions(token).headers,
      ...options.headers,
    },
  }

  const response = await fetch(url, requestOptions)

  // Handle non-2xx responses
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new Error(errorData.message || `API error: ${response.status}`)
  }

  // Handle empty responses (like for DELETE operations)
  const contentType = response.headers.get("content-type")
  if (contentType?.includes("application/json")) {
    return (await response.json()) as T
  }

  return {} as T
}

// Convenience methods for common HTTP methods
export const apiClient = {
  get: <T>(endpoint: string, token?: string | null) => 
    apiFetch<T>(endpoint, { method: 'GET' }, token),
  
  post: <T>(endpoint: string, data: unknown, token?: string | null) => 
    apiFetch<T>(endpoint, { 
      method: 'POST', 
      body: JSON.stringify(data) 
    }, token),
  
  put: <T>(endpoint: string, data: unknown, token?: string | null) => 
    apiFetch<T>(endpoint, { 
      method: 'PUT', 
      body: JSON.stringify(data) 
    }, token),
  
  patch: <T>(endpoint: string, data: unknown, token?: string | null) => 
    apiFetch<T>(endpoint, { 
      method: 'PATCH', 
      body: JSON.stringify(data) 
    }, token),
  
  delete: <T>(endpoint: string, token?: string | null) => 
    apiFetch<T>(endpoint, { method: 'DELETE' }, token),
}
