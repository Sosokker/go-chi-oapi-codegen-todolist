"use client";

import { useState } from "react";
import { toast } from "sonner";
import {
  useTags,
  useCreateTag,
  useUpdateTag,
  useDeleteTag,
} from "@/hooks/use-tags";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Icons } from "@/components/icons";
import type { Tag } from "@/services/api-types";

export default function TagsPage() {
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [currentTag, setCurrentTag] = useState<Tag | null>(null);
  const [formData, setFormData] = useState({
    name: "",
    color: "#FF5A5F",
    icon: "",
  });

  const { data: tags = [], isLoading } = useTags();
  const createTagMutation = useCreateTag();
  const updateTagMutation = useUpdateTag();
  const deleteTagMutation = useDeleteTag();

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleCreateTag = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await createTagMutation.mutateAsync(formData);
      setIsCreateDialogOpen(false);
      setFormData({ name: "", color: "#FF5A5F", icon: "" });
      toast.success("Tag created successfully");
    } catch (error) {
      toast.error("Failed to create tag");
    }
  };

  const handleEditTag = (tag: Tag) => {
    setCurrentTag(tag);
    setFormData({
      name: tag.name,
      color: tag.color || "#FF5A5F",
      icon: tag.icon || "",
    });
    setIsEditDialogOpen(true);
  };

  const handleUpdateTag = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentTag) return;

    try {
      await updateTagMutation.mutateAsync({
        id: currentTag.id,
        tag: formData,
      });
      setIsEditDialogOpen(false);
      toast.success("Tag updated successfully");
    } catch (error) {
      toast.error("Failed to update tag");
    }
  };

  const handleDeleteTag = async (id: string) => {
    try {
      await deleteTagMutation.mutateAsync(id);
      toast.success("Tag deleted successfully");
    } catch (error) {
      toast.error("Failed to delete tag");
    }
  };

  return (
    <div className="container max-w-3xl mx-auto px-4 py-6 space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold tracking-tight">Labels</h1>
        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button variant="ghost" size="icon" className="rounded-full">
              <Icons.plus className="h-5 w-5" />
              <span className="sr-only">Add Label</span>
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Label</DialogTitle>
              <DialogDescription>
                Add a new label to organize your todos
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleCreateTag} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  name="name"
                  value={formData.name}
                  onChange={handleChange}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="color">Color</Label>
                <div className="flex gap-2">
                  <Input
                    id="color"
                    name="color"
                    type="color"
                    value={formData.color}
                    onChange={handleChange}
                    className="w-12 h-10 p-1"
                  />
                  <Input
                    name="color"
                    value={formData.color}
                    onChange={handleChange}
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="icon">Icon (optional)</Label>
                <Input
                  id="icon"
                  name="icon"
                  value={formData.icon}
                  onChange={handleChange}
                  placeholder="e.g., home, work, star"
                />
              </div>
              <Button
                type="submit"
                className="w-full bg-[#FF5A5F] hover:bg-[#FF5A5F]/90"
              >
                Create Label
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <div className="rounded-lg border bg-card">
        <div className="flex items-center justify-between p-3 border-b">
          <h2 className="text-sm font-medium">Labels</h2>
        </div>
        {isLoading ? (
          <div className="p-4 space-y-3">
            {Array(6)
              .fill(0)
              .map((_, i) => (
                <div
                  key={i}
                  className="h-8 animate-pulse bg-muted rounded-md"
                />
              ))}
          </div>
        ) : tags.length === 0 ? (
          <div className="p-8 text-center">
            <Icons.tag className="h-12 w-12 mx-auto text-muted-foreground/50 mb-4" />
            <h3 className="text-lg font-medium text-muted-foreground mb-2">
              No labels found
            </h3>
            <p className="text-sm text-muted-foreground mb-4">
              Create a label to organize your todos
            </p>
            <Button
              onClick={() => setIsCreateDialogOpen(true)}
              className="bg-[#FF5A5F] hover:bg-[#FF5A5F]/90"
            >
              <Icons.plus className="mr-2 h-4 w-4" />
              Create your first label
            </Button>
          </div>
        ) : (
          <div className="divide-y">
            {tags.map((tag) => (
              <div
                key={tag.id}
                className="flex items-center justify-between py-3 px-4 hover:bg-muted/50 transition-colors cursor-pointer"
                onClick={() => handleEditTag(tag)}
              >
                <div className="flex items-center gap-3">
                  <div
                    className="w-5 h-5 flex items-center justify-center"
                    style={{ color: tag.color || "#FF5A5F" }}
                  >
                    <Icons.tag className="h-5 w-5" />
                  </div>
                  <span className="text-sm font-medium">{tag.name}</span>
                </div>
                <div className="flex items-center">
                  <span className="text-xs text-muted-foreground mr-2">
                    {Math.floor(Math.random() * 10)}
                  </span>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 opacity-0 group-hover:opacity-100"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleDeleteTag(tag.id);
                    }}
                  >
                    <Icons.trash className="h-4 w-4" />
                    <span className="sr-only">Delete</span>
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Label</DialogTitle>
            <DialogDescription>Update your label details</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleUpdateTag} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="edit-name">Name</Label>
              <Input
                id="edit-name"
                name="name"
                value={formData.name}
                onChange={handleChange}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-color">Color</Label>
              <div className="flex gap-2">
                <Input
                  id="edit-color"
                  name="color"
                  type="color"
                  value={formData.color}
                  onChange={handleChange}
                  className="w-12 h-10 p-1"
                />
                <Input
                  name="color"
                  value={formData.color}
                  onChange={handleChange}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-icon">Icon (optional)</Label>
              <Input
                id="edit-icon"
                name="icon"
                value={formData.icon}
                onChange={handleChange}
                placeholder="e.g., home, work, star"
              />
            </div>
            <div className="flex gap-2">
              <Button
                type="submit"
                className="flex-1 bg-[#FF5A5F] hover:bg-[#FF5A5F]/90"
              >
                Update Label
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => handleDeleteTag(currentTag?.id || "")}
                className="flex-1"
              >
                Delete Label
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
