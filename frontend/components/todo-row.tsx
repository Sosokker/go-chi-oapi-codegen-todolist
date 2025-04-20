"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import Image from "next/image"; // Import Image
import { useAuth } from "@/hooks/use-auth";
import { deleteAttachment } from "@/services/api-attachments";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"; // Import AlertDialog
import { TodoForm } from "@/components/todo-form";
import { Icons } from "@/components/icons";
import { cn } from "@/lib/utils";
import type { Todo, Tag } from "@/services/api-types";

interface TodoRowProps {
  todo: Todo;
  tags: Tag[];
  onUpdate: (todo: Partial<Todo>) => Promise<void>; // Make async
  onDelete: () => void;
  onAttachmentsChanged?: (attachments: AttachmentInfo[]) => void; // Add callback
  isDraggable?: boolean;
}

export function TodoRow({
  todo,
  tags,
  onUpdate,
  onDelete,
  onAttachmentsChanged,
  isDraggable = false,
}: TodoRowProps) {
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [isViewDialogOpen, setIsViewDialogOpen] = useState(false);
  const { token } = useAuth();

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: todo.id,
    disabled: !isDraggable,
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const todoTags = tags.filter((tag) => todo.tagIds?.includes(tag.id));
  const hasAttachments = todo.attachments && todo.attachments.length > 0;

  const handleStatusToggle = () => {
    const newStatus = todo.status === "completed" ? "pending" : "completed";
    onUpdate({ status: newStatus });
  };

  const formatDate = (dateString?: string | null) => {
    if (!dateString) return null;
    const date = new Date(dateString);
    return date.toLocaleDateString();
  };

  const handleDeleteAttachment = async (attachmentId: string) => {
    if (!token) {
      toast.error("Authentication required.");
      return;
    }
    try {
      await deleteAttachment(todo.id, attachmentId, token);
      toast.success("Attachment deleted.");
      const updatedAttachments = todo.attachments.filter(
        (att) => att.fileId !== attachmentId
      );
      onAttachmentsChanged?.(updatedAttachments);
      // Also update the local state for the dialog if it's open
      setFormData((prev) => ({ ...prev, attachments: updatedAttachments }));
    } catch (error) {
      console.error("Failed to delete attachment:", error);
      toast.error("Failed to delete attachment.");
    }
  };

  // State to manage form data within the row component context if needed, e.g., for the view dialog
  const [formData, setFormData] = useState(todo);
  useEffect(() => {
    setFormData(todo);
  }, [todo]);

  return (
    <>
      <div
        ref={setNodeRef}
        style={style}
        className={cn(
          "group flex items-center gap-3 px-4 py-2 hover:bg-muted/50 rounded-md transition-colors",
          todo.status === "completed" ? "text-muted-foreground" : "",
          isDraggable ? "cursor-grab active:cursor-grabbing" : "",
          isDragging ? "shadow-lg bg-muted" : ""
        )}
        {...(isDraggable ? { ...attributes, ...listeners } : {})}
      >
        <Checkbox
          checked={todo.status === "completed"}
          onCheckedChange={() => handleStatusToggle()}
          className="h-5 w-5 rounded-full flex-shrink-0" // Added flex-shrink-0
        />

        <div
          className="flex-1 min-w-0 cursor-pointer"
          onClick={() => setIsViewDialogOpen(true)}
        >
          {" "}
          {/* Make main area clickable */}
          <div className="flex items-center gap-2">
            <span
              className={cn(
                "text-sm font-medium truncate",
                todo.status === "completed" ? "line-through" : ""
              )}
            >
              {todo.title}
            </span>
            {/* Indicators */}
            <div className="flex items-center gap-1">
              {hasAttachments && (
                <Icons.paperclip className="h-3 w-3 text-muted-foreground" />
              )}
              {/* Add other indicators like subtasks if needed */}
            </div>
          </div>
          {todo.description && (
            <p className="text-xs text-muted-foreground truncate mt-0.5">
              {todo.description}
            </p>
          )}
        </div>

        {/* Right-aligned section */}
        <div className="flex items-center gap-2 ml-auto flex-shrink-0">
          {" "}
          {/* Added ml-auto and flex-shrink-0 */}
          {/* Tags */}
          {todoTags.length > 0 && (
            <div className="hidden sm:flex gap-1">
              {" "}
              {/* Hide tags on very small screens */}
              {todoTags.slice(0, 2).map((tag) => (
                <div
                  key={tag.id}
                  className="w-2 h-2 rounded-full"
                  style={{ backgroundColor: tag.color || "#FF5A5F" }}
                  title={tag.name}
                />
              ))}
              {todoTags.length > 2 && (
                <span className="text-[10px] text-muted-foreground">
                  +{todoTags.length - 2}
                </span>
              )}
            </div>
          )}
          {/* Deadline */}
          {todo.deadline && (
            <span className="text-xs text-muted-foreground whitespace-nowrap hidden md:inline">
              {formatDate(todo.deadline)}
            </span>
          )}
          {/* Action Buttons (Appear on Hover) */}
          <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity duration-150">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setIsViewDialogOpen(true)}
              className="h-7 w-7"
            >
              <Icons.eye className="h-3.5 w-3.5" />{" "}
              <span className="sr-only">View</span>
            </Button>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setIsEditDialogOpen(true)}
              className="h-7 w-7"
            >
              <Icons.edit className="h-3.5 w-3.5" />{" "}
              <span className="sr-only">Edit</span>
            </Button>
            {/* Delete Confirmation */}
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 text-muted-foreground hover:text-destructive"
                >
                  <Icons.trash className="h-3.5 w-3.5" />{" "}
                  <span className="sr-only">Delete</span>
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete Todo?</AlertDialogTitle>
                  <AlertDialogDescription>
                    Are you sure you want to delete &quot;{todo.title}&quot;?
                    This action cannot be undone.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    onClick={onDelete}
                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  >
                    Delete
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </div>

      {/* Edit Dialog */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Edit Todo</DialogTitle>
            <DialogDescription>
              Update your todo details and attachments.
            </DialogDescription>
          </DialogHeader>
          <TodoForm
            todo={todo}
            tags={tags}
            onSubmit={async (updatedTodoData) => {
              await onUpdate(updatedTodoData);
              setIsEditDialogOpen(false);
            }}
            onAttachmentsChanged={(newAttachments) => {
              // If the parent needs immediate notification of attachment changes
              onAttachmentsChanged?.(newAttachments);
              // Optionally update local state if needed within this component context
              setFormData((prev) => ({ ...prev, attachments: newAttachments }));
            }}
          />
        </DialogContent>
      </Dialog>

      {/* View Dialog */}
      <Dialog open={isViewDialogOpen} onOpenChange={setIsViewDialogOpen}>
        {/* Re-use the View Dialog structure from TodoCard, adapting as needed */}
        <DialogContent className="sm:max-w-[450px] p-0 overflow-hidden">
          <DialogHeader className="px-6 pt-6 pb-2">
            {/* ... Header content (status badge, title, etc.) ... */}
            <div className="flex items-center gap-2 mb-2">
              {/* Status Badge */}
              <Badge
                className={cn(
                  "px-2.5 py-1 capitalize font-medium text-xs",
                  todo.status === "pending"
                    ? "bg-amber-50 text-amber-700"
                    : todo.status === "in-progress"
                    ? "bg-sky-50 text-sky-700"
                    : todo.status === "completed"
                    ? "bg-emerald-50 text-emerald-700"
                    : "bg-slate-50 text-slate-700"
                )}
              >
                <div
                  className={cn(
                    "w-2 h-2 rounded-full mr-1.5",
                    todo.status === "pending"
                      ? "bg-amber-500"
                      : todo.status === "in-progress"
                      ? "bg-sky-500"
                      : todo.status === "completed"
                      ? "bg-emerald-500"
                      : "bg-slate-400"
                  )}
                />
                {todo.status.replace("-", " ")}
              </Badge>
              {/* Deadline Badge */}
              {todo.deadline && (
                <Badge
                  variant="outline"
                  className="text-xs font-normal px-2 py-0 ml-auto"
                >
                  {" "}
                  <Icons.calendar className="h-3 w-3 mr-1" />{" "}
                  {formatDate(todo.deadline)}{" "}
                </Badge>
              )}
            </div>
            <DialogTitle className="text-xl font-semibold">
              {formData.title}
            </DialogTitle>
            <DialogDescription className="text-xs mt-1">
              Created {formatDate(formData.createdAt)}
            </DialogDescription>
          </DialogHeader>

          {/* Cover Image (find first image attachment) */}
          {formData.attachments?.find((att) =>
            att.contentType.startsWith("image/")
          ) && (
            <div className="w-full h-48 overflow-hidden relative bg-muted">
              <Image
                src={
                  formData.attachments.find((att) =>
                    att.contentType.startsWith("image/")
                  )?.fileUrl || "/placeholder.svg"
                }
                alt={formData.title}
                fill
                style={{ objectFit: "cover" }}
                sizes="450px"
              />
            </div>
          )}

          <div className="px-6 py-4 space-y-4 max-h-[50vh] overflow-y-auto">
            {/* Description */}
            {formData.description && (
              <div className="text-sm text-muted-foreground">
                {formData.description}
              </div>
            )}
            {/* Tags */}
            {todoTags.length > 0 && (
              <div>
                {" "}
                <h4 className="text-xs font-medium text-muted-foreground mb-1.5">
                  TAGS
                </h4>{" "}
                <div className="flex flex-wrap gap-1.5">
                  {todoTags.map((tag) => (
                    <Badge
                      key={tag.id}
                      variant="outline"
                      className="px-2 py-0.5 text-xs font-normal border-0"
                      style={{
                        backgroundColor: `${tag.color}20` || "#FF5A5F20",
                        color: tag.color || "#FF5A5F",
                      }}
                    >
                      {tag.name}
                    </Badge>
                  ))}
                </div>{" "}
              </div>
            )}
            {/* Subtasks */}
            {formData.subtasks && formData.subtasks.length > 0 && (
              <div>
                <h4 className="text-xs font-medium text-muted-foreground mb-1.5">
                  SUBTASKS (
                  {formData.subtasks.filter((s) => s.completed).length}/
                  {formData.subtasks.length})
                </h4>
                {/* ... Subtask list rendering ... */}
              </div>
            )}
            {/* Attachments */}
            {(formData.attachments?.length ?? 0) > 0 && (
              <div>
                <h4 className="text-xs font-medium text-muted-foreground mb-1.5">
                  ATTACHMENTS
                </h4>
                <div className="space-y-2">
                  {formData.attachments.map((att) => (
                    <div
                      key={att.fileId}
                      className="flex items-center justify-between p-2 rounded-md border bg-muted/50"
                    >
                      <div className="flex items-center gap-2 flex-1 min-w-0">
                        {/* Icon or Thumbnail */}
                        {att.contentType.startsWith("image/") ? (
                          <div className="w-8 h-8 rounded overflow-hidden flex-shrink-0 bg-background">
                            <Image
                              src={att.fileUrl}
                              alt={att.fileName}
                              width={32}
                              height={32}
                              className="object-cover"
                            />
                          </div>
                        ) : (
                          <div className="w-8 h-8 rounded bg-background flex items-center justify-center flex-shrink-0">
                            <Icons.file className="w-4 h-4 text-muted-foreground" />
                          </div>
                        )}
                        {/* File Info */}
                        <div className="flex-1 min-w-0">
                          <a
                            href={att.fileUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-xs font-medium truncate block hover:underline"
                            title={att.fileName}
                          >
                            {" "}
                            {att.fileName}{" "}
                          </a>
                          <span className="text-[10px] text-muted-foreground">
                            {" "}
                            {(att.size / 1024).toFixed(1)} KB -{" "}
                            {att.contentType}{" "}
                          </span>
                        </div>
                      </div>
                      {/* Delete Button */}
                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7 text-muted-foreground hover:text-destructive flex-shrink-0"
                          >
                            {" "}
                            <Icons.trash className="h-3.5 w-3.5" />{" "}
                          </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            {" "}
                            <AlertDialogTitle>
                              Delete Attachment?
                            </AlertDialogTitle>{" "}
                            <AlertDialogDescription>
                              {" "}
                              Are you sure you want to delete &quot;
                              {att.fileName}&quot;? This action cannot be
                              undone.{" "}
                            </AlertDialogDescription>{" "}
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            {" "}
                            <AlertDialogCancel>Cancel</AlertDialogCancel>{" "}
                            <AlertDialogAction
                              onClick={() => handleDeleteAttachment(att.fileId)}
                              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                            >
                              {" "}
                              Delete{" "}
                            </AlertDialogAction>{" "}
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
          {/* Footer Actions */}
          <div className="border-t px-6 py-3 bg-muted/30 flex justify-end gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setIsViewDialogOpen(false)}
            >
              {" "}
              Close{" "}
            </Button>
            <Button
              size="sm"
              onClick={() => {
                setIsViewDialogOpen(false);
                setIsEditDialogOpen(true);
              }}
            >
              {" "}
              <Icons.edit className="h-3.5 w-3.5 mr-1.5" /> Edit Todo{" "}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
