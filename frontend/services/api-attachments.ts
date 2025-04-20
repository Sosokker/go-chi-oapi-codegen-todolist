import { apiClient } from "./api-client";
import type { FileUploadResponse } from "./api-types";

/**
 * Uploads a file attachment to a specific todo item.
 * @param todoId - The ID of the todo item.
 * @param file - The File object to upload.
 * @param token - The authentication token.
 * @returns A promise resolving to the FileUploadResponse.
 */
export async function uploadAttachment(
  todoId: string,
  file: File,
  token: string
): Promise<FileUploadResponse> {
  const formData = new FormData();
  formData.append("file", file);

  return await apiClient.upload<FileUploadResponse>(
    `/todos/${todoId}/attachments`,
    formData,
    token
  );
}

/**
 * Deletes an attachment from a specific todo item.
 * @param todoId - The ID of the todo item.
 * @param attachmentPath - The storage path/ID of the attachment to delete (URL-encoded).
 * @param token - The authentication token.
 * @returns A promise resolving when the deletion is complete.
 */
export async function deleteAttachment(
  todoId: string,
  attachmentPath: string, // This is the FileID from AttachmentInfo
  token: string
): Promise<void> {
  // Ensure the path is URL encoded for the request path
  const encodedPath = encodeURIComponent(attachmentPath);
  await apiClient.delete<void>(
    `/todos/${todoId}/attachments/${encodedPath}`,
    token
  );
}
