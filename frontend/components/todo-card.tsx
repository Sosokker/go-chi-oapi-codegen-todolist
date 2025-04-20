"use client";

import { useState } from "react";

import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { Card, CardContent, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from "@/components/ui/dropdown-menu";

import { TodoForm } from "@/components/todo-form";
import { Icons } from "@/components/icons";
import { cn } from "@/lib/utils";
import type { Todo, Tag } from "@/services/api-types";
import Image from "next/image";

interface TodoCardProps {
  todo: Todo;
  tags: Tag[];
  onUpdate: (todo: Partial<Todo>) => Promise<void>;
  onDelete: () => void;
  onAttachmentsChanged?: (attachments: string[]) => void;
  isDraggable?: boolean;
}

export function TodoCard({
  todo,
  tags,
  onUpdate,
  onDelete,
  onAttachmentsChanged,
  isDraggable = false,
}: TodoCardProps) {
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [isViewDialogOpen, setIsViewDialogOpen] = useState(false);

  // Drag and drop logic
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

  // --- Helper Functions ---
  const getStatusColor = (status: string) => {
    switch (status) {
      case "pending":
        return "border-l-4 border-l-amber-500";
      case "in-progress":
        return "border-l-4 border-l-sky-500";
      case "completed":
        return "border-l-4 border-l-emerald-500";
      default:
        return "border-l-4 border-l-slate-400";
    }
  };
  // const getStatusIcon = (status: string) => {
  //   switch (status) {
  //     case "pending":
  //       return <Icons.clock className="h-5 w-5 text-amber-500" />;
  //     case "in-progress":
  //       return <Icons.loader className="h-5 w-5 text-sky-500" />;
  //     case "completed":
  //       return <Icons.checkSquare className="h-5 w-5 text-emerald-500" />;
  //     default:
  //       return <Icons.circle className="h-5 w-5 text-slate-400" />;
  //   }
  // };
  const formatDate = (dateString?: string | null) => {
    if (!dateString) return "";
    const date = new Date(dateString);
    return date.toLocaleDateString();
  };

  const todoTags = tags.filter((tag) => todo.tagIds?.includes(tag.id));
  const hasAttachments = !!todo.attachmentUrl;
  const hasSubtasks = todo.subtasks && todo.subtasks.length > 0;
  const completedSubtasks =
    todo.subtasks?.filter((s) => s.completed).length ?? 0;

  // --- Event Handlers ---
  const handleStatusToggle = () => {
    const newStatus = todo.status === "completed" ? "pending" : "completed";
    onUpdate({ status: newStatus });
  };

  // --- Rendering ---
  const coverImage =
    todo.attachmentUrl &&
    todo.attachmentUrl.match(/\.(jpg|jpeg|png|gif|webp)$/i)
      ? todo.attachmentUrl
      : null;

  return (
    <>
      <Card
        ref={setNodeRef}
        style={style}
        className={cn(
          "transition-shadow duration-150 ease-out",
          getStatusColor(todo.status),
          "shadow-sm hover:shadow-md min-w-[220px] max-w-[420px] bg-card",
          isDraggable ? "cursor-grab active:cursor-grabbing" : "",
          isDragging
            ? "shadow-lg ring-2 ring-primary ring-opacity-50 scale-105 z-10"
            : ""
        )}
        {...(isDraggable ? { ...attributes, ...listeners } : {})}
        onClick={() => !isDraggable && setIsViewDialogOpen(true)}
      >
        <CardContent className="p-3">
          {/* Optional Cover Image */}
          {coverImage && (
            <div className="relative h-24 w-full mb-2 rounded overflow-hidden">
              <Image
                src={coverImage || "/placeholder.svg"}
                alt="Todo Attachment"
                fill
                style={{ objectFit: "cover" }}
                sizes="(max-width: 640px) 100vw, 300px"
                className="bg-muted"
                priority={false}
              />
            </div>
          )}
          {/* Title and Menu */}
          <div className="flex items-start justify-between mb-1">
            <CardTitle
              className={cn(
                "text-sm font-medium leading-snug pr-2",
                todo.status === "completed" &&
                  "line-through text-muted-foreground"
              )}
            >
              {todo.title}
            </CardTitle>
            {!isDraggable && (
              <DropdownMenu
                onOpenChange={(open) => open && setIsViewDialogOpen(false)}
              >
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6 flex-shrink-0 -mt-1 -mr-1"
                  >
                    <Icons.moreVertical className="h-3.5 w-3.5" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="end"
                  onClick={(e) => e.stopPropagation()}
                >
                  <DropdownMenuItem onClick={() => setIsViewDialogOpen(true)}>
                    <Icons.eye className="h-3.5 w-3.5 mr-2" /> View
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => setIsEditDialogOpen(true)}>
                    <Icons.edit className="h-3.5 w-3.5 mr-2" /> Edit
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={onDelete} className="text-red-600">
                    <Icons.trash className="h-3.5 w-3.5 mr-2" /> Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            )}
          </div>
          {/* Description (optional, truncated) */}
          {todo.description && (
            <p className="text-xs text-muted-foreground mb-1.5 line-clamp-2">
              {todo.description}
            </p>
          )}
          {/* Badges and Indicators */}
          <div className="flex flex-wrap items-center gap-1.5 text-[10px] mb-1.5">
            {todoTags.map((tag) => (
              <Badge
                key={tag.id}
                variant="outline"
                className="px-1.5 py-0 h-4 text-[10px] font-normal border-0"
                style={{
                  backgroundColor: `${tag.color}20` || "#FF5A5F20",
                  color: tag.color || "#FF5A5F",
                }}
              >
                {tag.name}
              </Badge>
            ))}
            {hasSubtasks && (
              <span className="flex items-center gap-0.5 text-[10px] text-muted-foreground ml-2">
                <Icons.checkSquare className="h-3 w-3" />
                {completedSubtasks}/{todo.subtasks.length}
              </span>
            )}
            {hasAttachments && (
              <span className="flex items-center gap-0.5 text-[10px] text-muted-foreground ml-2">
                <Icons.paperclip className="h-3 w-3" />1
              </span>
            )}
            {todo.deadline && (
              <span className="flex items-center gap-0.5 text-[10px] text-muted-foreground ml-2">
                <Icons.calendar className="h-3 w-3" />
                {formatDate(todo.deadline)}
              </span>
            )}
          </div>
          {/* Bottom Row: Created Date & Status Toggle */}
          <div className="flex justify-between items-center mt-1">
            <span className="text-[10px] text-muted-foreground/70">
              {todo.createdAt ? formatDate(todo.createdAt) : ""}
            </span>
            <Button
              variant={todo.status === "completed" ? "ghost" : "outline"}
              size="sm"
              className="h-5 px-1.5 py-0 rounded-full"
              onClick={(e) => {
                e.stopPropagation();
                handleStatusToggle();
              }}
            >
              {todo.status === "completed" ? (
                <Icons.x className="h-2.5 w-2.5" />
              ) : (
                <Icons.check className="h-2.5 w-2.5" />
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader className="space-y-1">
            <DialogTitle className="text-xl">Edit Todo</DialogTitle>
            <DialogDescription className="text-sm">
              Make changes to your task, add attachments, and save.
            </DialogDescription>
          </DialogHeader>
          <TodoForm
            todo={todo}
            tags={tags}
            onSubmit={async (updatedTodoData) => {
              await onUpdate(updatedTodoData);
              setIsEditDialogOpen(false);
            }}
            onAttachmentsChanged={onAttachmentsChanged}
          />
        </DialogContent>
      </Dialog>

      {/* View Dialog */}
      <Dialog open={isViewDialogOpen} onOpenChange={setIsViewDialogOpen}>
        <DialogContent className="sm:max-w-[450px] p-0 overflow-hidden">
          <DialogHeader className="px-6 pt-6 pb-2">
            <div className="flex items-center gap-2 mb-2">
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
                    getStatusColor(todo.status).replace("border-l-4 ", "bg-")
                  )}
                />
                {todo.status.replace("-", " ")}
              </Badge>
              {todo.deadline && (
                <Badge
                  variant="outline"
                  className="text-xs font-normal px-2 py-0 ml-auto"
                >
                  <Icons.calendar className="h-3 w-3 mr-1" />
                  {formatDate(todo.deadline)}
                </Badge>
              )}
            </div>
            <DialogTitle className="text-xl font-semibold">
              {todo.title}
            </DialogTitle>
            <DialogDescription className="text-xs mt-1">
              Created {formatDate(todo.createdAt)}
            </DialogDescription>
          </DialogHeader>
          {/* Cover Image */}
          {coverImage && (
            <div className="w-full h-48 overflow-hidden relative bg-muted">
              <Image
                src={coverImage || "/placeholder.svg"}
                alt={todo.title}
                fill
                style={{ objectFit: "cover" }}
                sizes="450px"
                priority
              />
            </div>
          )}
          {/* Content Section */}
          <div className="px-6 py-4 space-y-4 max-h-[50vh] overflow-y-auto">
            {/* Description */}
            {todo.description && (
              <div className="text-sm text-muted-foreground">
                {todo.description}
              </div>
            )}
            {/* Tags */}
            {todoTags.length > 0 && (
              <div>
                <h4 className="text-xs font-medium text-muted-foreground mb-1.5">
                  TAGS
                </h4>
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
                </div>
              </div>
            )}
            {/* Subtasks */}
            {hasSubtasks && (
              <div>
                <h4 className="text-xs font-medium text-muted-foreground mb-1.5">
                  SUBTASKS ({completedSubtasks}/{todo.subtasks.length})
                </h4>
                <ul className="space-y-1.5 text-sm">
                  {todo.subtasks.map((subtask) => (
                    <li key={subtask.id} className="flex items-center gap-2">
                      <div
                        className={cn(
                          "w-4 h-4 rounded-full flex items-center justify-center border",
                          subtask.completed
                            ? "border-emerald-500 bg-emerald-50"
                            : "border-gray-200"
                        )}
                      >
                        {subtask.completed && (
                          <Icons.check className="h-2.5 w-2.5 text-emerald-500" />
                        )}
                      </div>
                      <span
                        className={cn(
                          "text-xs",
                          subtask.completed
                            ? "line-through text-muted-foreground"
                            : ""
                        )}
                      >
                        {subtask.description}
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
            {/* Attachments */}
            {hasAttachments && (
              <div>
                <h4 className="text-xs font-medium text-muted-foreground mb-1.5">
                  ATTACHMENT
                </h4>
                <div className="space-y-2">
                  {todo.attachmentUrl && (
                    <div className="flex items-center justify-between p-2 rounded-md border bg-muted/50">
                      <div className="flex items-center gap-2 flex-1 min-w-0">
                        <div className="w-8 h-8 rounded overflow-hidden flex-shrink-0 bg-background">
                          <Image
                            src={todo.attachmentUrl}
                            alt={todo.title}
                            width={32}
                            height={32}
                            className="object-cover"
                          />
                        </div>
                        <a
                          href={todo.attachmentUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="truncate text-xs text-blue-600 underline"
                        >
                          View Image
                        </a>
                      </div>
                    </div>
                  )}
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
              Close
            </Button>
            <Button
              size="sm"
              onClick={() => {
                setIsViewDialogOpen(false);
                setIsEditDialogOpen(true);
              }}
            >
              <Icons.edit className="h-3.5 w-3.5 mr-1.5" /> Edit Todo
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
