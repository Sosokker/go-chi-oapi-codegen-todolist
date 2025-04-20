"use client";

import { useState } from "react";
import { toast } from "sonner";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import {
  Card,
  CardContent,
  CardTitle,
  CardDescription,
  CardHeader,
} from "@/components/ui/card";
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
  onUpdate: (todo: Partial<Todo>) => void;
  onDelete: () => void;
  isDraggable?: boolean;
}

export function TodoCard({
  todo,
  tags,
  onUpdate,
  onDelete,
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

  // Style helpers
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

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "pending":
        return <Icons.clock className="h-5 w-5 text-amber-500" />;
      case "in-progress":
        return <Icons.loader className="h-5 w-5 text-sky-500" />;
      case "completed":
        return <Icons.checkSquare className="h-5 w-5 text-emerald-500" />;
      default:
        return <Icons.circle className="h-5 w-5 text-slate-400" />;
    }
  };

  const todoTags = tags.filter((tag) => todo.tagIds.includes(tag.id));
  const hasImage = !!todo.image;
  const hasAttachments = todo.attachments && todo.attachments.length > 0;
  const hasSubtasks = todo.subtasks && todo.subtasks.length > 0;
  const completedSubtasks = todo.subtasks
    ? todo.subtasks.filter((subtask) => subtask.completed).length
    : 0;

  const handleStatusToggle = () => {
    const newStatus = todo.status === "completed" ? "pending" : "completed";
    onUpdate({ status: newStatus });
  };

  const formatDate = (dateString?: string | null) => {
    if (!dateString) return "";
    const date = new Date(dateString);
    return date.toLocaleDateString();
  };

  return (
    <>
      <Card
        ref={setNodeRef}
        style={style}
        className={cn(
          "transition-colors",
          getStatusColor(todo.status),
          "shadow-sm min-w-[220px] max-w-[420px]",
          isDraggable ? "cursor-grab active:cursor-grabbing" : "",
          isDragging ? "shadow-lg" : ""
        )}
        {...(isDraggable ? { ...attributes, ...listeners } : {})}
      >
        <CardContent className="p-4 flex gap-4">
          {/* Left icon, like notification card */}
          <div className="flex-shrink-0 mt-1">{getStatusIcon(todo.status)}</div>
          {/* Main content */}
          <div className="flex-grow">
            <CardHeader className="p-0">
              <div className="flex items-center justify-between">
                <CardTitle
                  className={cn(
                    "text-base",
                    todo.status === "completed" &&
                      "line-through text-muted-foreground"
                  )}
                >
                  {todo.title}
                </CardTitle>
                {!isDraggable && (
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7"
                        aria-label="More actions"
                      >
                        <Icons.moreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem
                        onClick={() => setIsViewDialogOpen(true)}
                      >
                        <Icons.eye className="h-4 w-4 mr-2" />
                        View details
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={() => setIsEditDialogOpen(true)}
                      >
                        <Icons.edit className="h-4 w-4 mr-2" />
                        Edit
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={onDelete}
                        className="text-red-600"
                      >
                        <Icons.trash className="h-4 w-4 mr-2" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                )}
              </div>
              {/* Actions and deadline on a new line */}
              <div className="flex items-center gap-2 mt-1">
                {todo.deadline && (
                  <span className="text-xs text-muted-foreground flex items-center gap-1">
                    <Icons.calendar className="h-3 w-3" />
                    {formatDate(todo.deadline)}
                  </span>
                )}
              </div>
            </CardHeader>
            <CardDescription
              className={cn(
                "mt-1",
                todo.status === "completed" &&
                  "line-through text-muted-foreground/70"
              )}
            >
              {todo.description}
            </CardDescription>
            {/* Tags and indicators */}
            <div className="flex flex-wrap items-center gap-1.5 mt-2">
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
                  <Icons.paperclip className="h-3 w-3" />
                  {todo.attachments.length}
                </span>
              )}
              {hasImage && (
                <span className="flex items-center gap-0.5 text-[10px] text-muted-foreground ml-2">
                  <Icons.image className="h-3 w-3" />
                </span>
              )}
            </div>
            {/* Bottom row: created date and status toggle */}
            <div className="flex justify-between items-center mt-3">
              <span className="text-[10px] text-muted-foreground/70">
                {todo.createdAt ? formatDate(todo.createdAt) : ""}
              </span>
              <Button
                variant={todo.status === "completed" ? "ghost" : "outline"}
                size="sm"
                className="h-6 text-[10px] px-2 py-0 rounded-full"
                onClick={handleStatusToggle}
              >
                {todo.status === "completed" ? (
                  <Icons.x className="h-3 w-3" />
                ) : (
                  <Icons.check className="h-3 w-3" />
                )}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader className="space-y-1">
            <DialogTitle className="text-xl">Edit Todo</DialogTitle>
            <DialogDescription className="text-sm">
              Make changes to your task and save when you&apos;re done
            </DialogDescription>
          </DialogHeader>
          <TodoForm
            todo={todo}
            tags={tags}
            onSubmit={(updatedTodo) => {
              onUpdate(updatedTodo);
              setIsEditDialogOpen(false);
              toast.success("Todo updated successfully");
            }}
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
                  "px-2.5 py-1 capitalize font-medium text-sm",
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
                    "w-2.5 h-2.5 rounded-full mr-1.5",
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
          {hasImage && (
            <div className="w-full h-48 overflow-hidden relative">
              <Image
                src={todo.image || "/placeholder.svg?height=192&width=450"}
                alt={todo.title}
                fill
                style={{ objectFit: "cover" }}
                sizes="450px"
                priority
              />
            </div>
          )}
          <div className="px-6 py-4 space-y-4">
            {todo.description && (
              <div className="text-sm text-muted-foreground">
                {todo.description}
              </div>
            )}
            {todoTags.length > 0 && (
              <div>
                <h4 className="text-xs font-medium mb-1.5">Tags</h4>
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
            {hasAttachments && (
              <div>
                <h4 className="text-xs font-medium mb-1.5">Attachments</h4>
                <div className="space-y-1">
                  {todo.attachments.map((a, i) => (
                    <div
                      key={i}
                      className="flex items-center gap-1.5 text-xs text-muted-foreground"
                    >
                      <Icons.paperclip className="h-3.5 w-3.5" />
                      <span>{a}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
            {hasSubtasks && (
              <div>
                <h4 className="text-xs font-medium mb-1.5">
                  Subtasks ({completedSubtasks}/{todo.subtasks.length})
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
          </div>
          <div className="border-t px-6 py-4 flex justify-end gap-2">
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
              Edit
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
