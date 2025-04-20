"use client";

import type React from "react";
import Image from "next/image";

import { useState, useEffect } from "react";
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
import { Icons } from "@/components/icons";
import type { Todo, Tag } from "@/services/api-types";

interface TodoFormProps {
  todo?: Todo;
  tags: Tag[];
  onSubmit: (todo: Partial<Todo>) => void;
}

export function TodoForm({ todo, tags, onSubmit }: TodoFormProps) {
  const [formData, setFormData] = useState<Partial<Todo>>({
    title: "",
    description: "",
    status: "pending",
    deadline: undefined,
    tagIds: [],
    image: null,
  });
  const [imagePreview, setImagePreview] = useState<string | null>(null);

  useEffect(() => {
    if (todo) {
      setFormData({
        title: todo.title,
        description: todo.description || "",
        status: todo.status,
        deadline: todo.deadline,
        tagIds: todo.tagIds || [],
        image: todo.image || null,
      });

      if (todo.image) {
        setImagePreview(todo.image);
      }
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

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // In a real app, we would upload the file to a server and get a URL back
    // For now, we'll create a local object URL
    const imageUrl = URL.createObjectURL(file);
    setImagePreview(imageUrl);
    setFormData((prev) => ({ ...prev, image: imageUrl }));
  };

  const handleRemoveImage = () => {
    setImagePreview(null);
    setFormData((prev) => ({ ...prev, image: null }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
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
      <div className="space-y-2">
        <Label htmlFor="image">Image (optional)</Label>
        <div className="flex items-center gap-2">
          <Input
            id="image"
            name="image"
            type="file"
            accept="image/*"
            onChange={handleImageChange}
            className="flex-1"
          />
          {imagePreview && (
            <Button
              type="button"
              variant="outline"
              size="icon"
              onClick={handleRemoveImage}
            >
              <Icons.trash className="h-4 w-4" />
              <span className="sr-only">Remove image</span>
            </Button>
          )}
        </div>
        {imagePreview && (
          <div className="mt-2 relative w-full h-32 rounded-md overflow-hidden border">
            <Image
              src={imagePreview || "/placeholder.svg"}
              alt="Preview"
              width={400}
              height={128}
              className="w-full h-full object-cover rounded-md"
            />
          </div>
        )}
      </div>
      <Button
        type="submit"
        className="w-full bg-[#FF5A5F] hover:bg-[#FF5A5F]/90"
      >
        {todo ? "Update" : "Create"} Todo
      </Button>
    </form>
  );
}
