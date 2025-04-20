"use client";

import type React from "react";
import { useState, useEffect } from "react";
import { toast } from "sonner";
import { useAuth } from "@/hooks/use-auth";
import { uploadAttachment, deleteAttachment } from "@/services/api-attachments";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { MultiSelect } from "@/components/multi-select";
import type { Todo, Tag } from "@/services/api-types";
import { Progress } from "./ui/progress";

interface TodoFormProps {
  todo?: Todo;
  tags: Tag[];
  onSubmit: (todo: Partial<Todo>) => Promise<void>;
  onAttachmentsChanged?: (attachments: string[]) => void; // Now just an array of one or zero
}

export function TodoForm({
  todo,
  tags,
  onSubmit,
  onAttachmentsChanged,
}: TodoFormProps) {
  const { token } = useAuth();
  const [formData, setFormData] = useState<Partial<Todo>>({
    title: "",
    description: "",
    status: "pending",
    deadline: undefined,
    tagIds: [],
    attachmentUrl: null,
  });
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);

  useEffect(() => {
    if (todo) {
      setFormData({
        title: todo.title,
        description: todo.description || "",
        status: todo.status,
        deadline: todo.deadline,
        tagIds: todo.tagIds || [],
        attachmentUrl: todo.attachmentUrl || null,
      });
    } else {
      setFormData({
        title: "",
        description: "",
        status: "pending",
        deadline: undefined,
        tagIds: [],
        attachmentUrl: null,
      });
    }
  }, [todo]);

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSelectChange = (name: string, value: string) => {
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleTagsChange = (selected: string[]) => {
    setFormData((prev) => ({ ...prev, tagIds: selected }));
  };

  // Upload new attachment
  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    // Only one attachment is supported, so uploading a new one replaces the old

    if (!todo || !token) {
      toast.error("Cannot attach files until the todo is saved.");
      return;
    }
    const file = e.target.files?.[0];
    if (!file) return;
    // Only allow images
    if (!file.type.startsWith("image/")) {
      toast.error("Only image files are allowed.");
      return;
    }

    setIsUploading(true);
    setUploadProgress(0);

    // Simulate progress for demo – replace if backend supports it
    const progressInterval = setInterval(() => {
      setUploadProgress((p) => Math.min(p + 10, 90));
    }, 200);

    try {
      const response = await uploadAttachment(todo.id, file, token);
      clearInterval(progressInterval);
      setUploadProgress(100);
      toast.success(`Uploaded: "${response.fileName}"`);

      setFormData((f) => ({ ...f, attachmentUrl: response.fileUrl }));
      onAttachmentsChanged?.([response.fileUrl]);

      setTimeout(() => {
        setIsUploading(false);
        setUploadProgress(0);
      }, 500);
    } catch (err) {
      clearInterval(progressInterval);
      setIsUploading(false);
      setUploadProgress(0);
      console.error(err);
      toast.error(
        `Upload failed: ${err instanceof Error ? err.message : "Unknown error"}`
      );
    } finally {
      e.target.value = "";
    }
  };

  // Remove an existing attachment
  const handleRemoveAttachment = async (attachmentId: string) => {
    // Only one attachment is supported, so attachmentId is ignored

    if (!todo || !token) {
      toast.error("Cannot remove attachments right now.");
      return;
    }
    try {
      await deleteAttachment(todo.id, attachmentId, token);
      // Only one attachment is supported now, so just clear the attachmentUrl
      setFormData((f) => ({ ...f, attachmentUrl: null }));
      onAttachmentsChanged?.([]);
      toast.success("Attachment removed.");
    } catch (err) {
      console.error(err);
      toast.error(
        `Delete failed: ${err instanceof Error ? err.message : "Unknown error"}`
      );
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await onSubmit(formData);
  };
  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* Title, Description, Status, Deadline, Tags... */}
      <div className="space-y-2">
        <Label htmlFor="title">Title</Label>
        <Input
          id="title"
          name="title"
          value={formData.title}
          onChange={handleChange}
          required
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="description">Description (optional)</Label>
        <Textarea
          id="description"
          name="description"
          value={formData.description ?? ""}
          onChange={handleChange}
          rows={3}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="status">Status</Label>
        <Select
          value={formData.status}
          onValueChange={(value) => handleSelectChange("status", value)}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="pending">Pending</SelectItem>
            <SelectItem value="in-progress">In Progress</SelectItem>
            <SelectItem value="completed">Completed</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <Label htmlFor="deadline">Deadline (optional)</Label>
        <Input
          id="deadline"
          name="deadline"
          type="datetime-local"
          value={
            formData.deadline
              ? new Date(formData.deadline).toISOString().slice(0, 16)
              : ""
          }
          onChange={handleChange}
        />
      </div>
      <div className="space-y-2">
        <Label>Tags</Label>
        <MultiSelect
          options={tags.map((tag) => ({ label: tag.name, value: tag.id }))}
          selected={formData.tagIds || []}
          onChange={handleTagsChange}
          placeholder="Select tags"
        />
      </div>

      {/* Attachment Section - Only shown when editing an existing todo */}
      {todo && (
        <div className="space-y-2">
          <Label htmlFor="attachments">Attachments</Label>
          <Input
            id="file"
            type="file"
            accept="image/*"
            multiple={false}
            onChange={handleFileChange}
            disabled={isUploading}
          />
          {isUploading && (
            <Progress value={uploadProgress} className="w-full h-2" />
          )}
        </div>
      )}
      {formData.attachmentUrl && (
        <div className="mt-2">
          <ul>
            <li className="flex items-center gap-2">
              <a
                href={formData.attachmentUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-600 underline"
              >
                View Attachment
              </a>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => handleRemoveAttachment("")}
                disabled={isUploading}
              >
                Remove
              </Button>
            </li>
          </ul>
        </div>
      )}
      <Button type="submit" className="w-full h-10" disabled={isUploading}>
        {isUploading ? "Uploading..." : todo ? "Update Todo" : "Create Todo"}
      </Button>
    </form>
  );
}
