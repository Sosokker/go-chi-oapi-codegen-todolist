import { useAuthStore } from "@/store/auth-store";

const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080/api/v1";

const getRequestOptions = (token?: string | null, isFormData = false) => {
  const headers: HeadersInit = {};

  if (!isFormData) {
    headers["Content-Type"] = "application/json";
  }

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return { headers };
};

async function apiFetch<T>(
  endpoint: string,
  options: RequestInit = {},
  token?: string | null,
  isFormData = false,
  simulateDelay = process.env.NODE_ENV === "development" ? 300 : 0
): Promise<T> {
  if (simulateDelay > 0) {
    await delay(simulateDelay);
  }

  const url = `${API_BASE_URL}${
    endpoint.startsWith("/") ? endpoint : `/${endpoint}`
  }`;

  // Merge headers carefully
  const baseOptions = getRequestOptions(token, isFormData);
  const requestOptions = {
    ...options,
    headers: {
      ...baseOptions.headers, // Use headers from getRequestOptions
      ...options.headers, // Allow overriding/adding headers from options
    },
  };

  const response = await fetch(url, requestOptions);

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(
      errorData.message || `API error: ${response.status} ${response.statusText}`
    );
  }

  // Handle empty responses (like 204 No Content)
  if (response.status === 204) {
    return {} as T;
  }

  const contentType = response.headers.get("content-type");
  if (contentType?.includes("application/json")) {
    return (await response.json()) as T;
  }

  // Handle other content types if necessary, or return raw response?
  // For now, assume JSON or empty for successful non-error responses
  return {} as T;
}

// Add a function specifically for FormData uploads
async function apiUpload<T>(
  endpoint: string,
  formData: FormData,
  token?: string | null
): Promise<T> {
  // Get token from store if not provided explicitly
  const authToken = token ?? useAuthStore.getState().token;

  const url = `${API_BASE_URL}${
    endpoint.startsWith("/") ? endpoint : `/${endpoint}`
  }`;

  const headers: HeadersInit = {};
  if (authToken) {
    headers["Authorization"] = `Bearer ${authToken}`;
  }
  // DO NOT set Content-Type header for FormData

  const response = await fetch(url, {
    method: "POST",
    body: formData,
    headers: headers,
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(
      errorData.message || `API error: ${response.status} ${response.statusText}`
    );
  }

  if (response.status === 204) {
    return {} as T;
  }

  const contentType = response.headers.get("content-type");
  if (contentType?.includes("application/json")) {
    return (await response.json()) as T;
  }

  return {} as T;
}

export const apiClient = {
  get: <T>(endpoint: string, token?: string | null) =>
    apiFetch<T>(endpoint, { method: "GET" }, token),

  post: <T>(endpoint: string, data: unknown, token?: string | null) =>
    apiFetch<T>(
      endpoint,
      {
        method: "POST",
        body: JSON.stringify(data),
      },
      token
    ),

  put: <T>(endpoint: string, data: unknown, token?: string | null) =>
    apiFetch<T>(
      endpoint,
      {
        method: "PUT",
        body: JSON.stringify(data),
      },
      token
    ),

  patch: <T>(endpoint: string, data: unknown, token?: string | null) =>
    apiFetch<T>(
      endpoint,
      {
        method: "PATCH",
        body: JSON.stringify(data),
      },
      token
    ),

  delete: <T>(endpoint: string, token?: string | null) =>
    apiFetch<T>(endpoint, { method: "DELETE" }, token),

  // Expose the upload function
  upload: <T>(endpoint: string, formData: FormData, token?: string | null) =>
    apiUpload<T>(endpoint, formData, token),
};
