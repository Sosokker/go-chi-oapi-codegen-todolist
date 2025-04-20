"use client"

import { useState } from "react"
import { toast } from "sonner"
import { useSortable } from "@dnd-kit/sortable"
import { CSS } from "@dnd-kit/utilities"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { TodoForm } from "@/components/todo-form"
import { Icons } from "@/components/icons"
import { cn } from "@/lib/utils"
import type { Todo, Tag } from "@/services/api-types"

interface TodoRowProps {
  todo: Todo
  tags: Tag[]
  onUpdate: (todo: Partial<Todo>) => void
  onDelete: () => void
  isDraggable?: boolean
}

export function TodoRow({ todo, tags, onUpdate, onDelete, isDraggable = false }: TodoRowProps) {
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false)
  const [isViewDialogOpen, setIsViewDialogOpen] = useState(false)

  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: todo.id,
    disabled: !isDraggable,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const todoTags = tags.filter((tag) => todo.tagIds.includes(tag.id))

  const handleStatusToggle = () => {
    const newStatus = todo.status === "completed" ? "pending" : "completed"
    onUpdate({ status: newStatus })
  }

  const formatDate = (dateString?: string | null) => {
    if (!dateString) return null
    const date = new Date(dateString)
    return date.toLocaleDateString()
  }

  // Check if todo has image or attachments
  const hasAttachments = todo.attachments && todo.attachments.length > 0

  return (
    <>
      <div
        ref={setNodeRef}
        style={style}
        className={cn(
          "group flex items-center gap-3 px-4 py-2 hover:bg-muted/50 rounded-md transition-colors",
          todo.status === "completed" ? "text-muted-foreground" : "",
          isDraggable ? "cursor-grab active:cursor-grabbing" : "",
          isDragging ? "shadow-lg bg-muted" : "",
        )}
        {...(isDraggable ? { ...attributes, ...listeners } : {})}
      >
        <Checkbox
          checked={todo.status === "completed"}
          onCheckedChange={() => handleStatusToggle()}
          className="h-5 w-5 rounded-full"
        />

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className={cn("text-sm font-medium truncate", todo.status === "completed" ? "line-through" : "")}>
              {todo.title}
            </span>
            {todo.image && <Icons.image className="h-3.5 w-3.5 text-muted-foreground" />}
            {hasAttachments && <Icons.paperclip className="h-3.5 w-3.5 text-muted-foreground" />}
          </div>

          {todo.description && <p className="text-xs text-muted-foreground truncate">{todo.description}</p>}
        </div>

        <div className="flex items-center gap-2">
          {todoTags.length > 0 && (
            <div className="flex gap-1">
              {todoTags.slice(0, 2).map((tag) => (
                <div
                  key={tag.id}
                  className="w-2 h-2 rounded-full"
                  style={{ backgroundColor: tag.color || "#FF5A5F" }}
                />
              ))}
              {todoTags.length > 2 && <span className="text-xs text-muted-foreground">+{todoTags.length - 2}</span>}
            </div>
          )}

          {todo.deadline && (
            <span className="text-xs text-muted-foreground whitespace-nowrap">{formatDate(todo.deadline)}</span>
          )}

          <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <Button variant="ghost" size="icon" onClick={() => setIsViewDialogOpen(true)} className="h-7 w-7">
              <Icons.eye className="h-3.5 w-3.5" />
              <span className="sr-only">View</span>
            </Button>
            <Button variant="ghost" size="icon" onClick={() => setIsEditDialogOpen(true)} className="h-7 w-7">
              <Icons.edit className="h-3.5 w-3.5" />
              <span className="sr-only">Edit</span>
            </Button>
            <Button variant="ghost" size="icon" onClick={onDelete} className="h-7 w-7">
              <Icons.trash className="h-3.5 w-3.5" />
              <span className="sr-only">Delete</span>
            </Button>
          </div>
        </div>
      </div>

      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Todo</DialogTitle>
            <DialogDescription>Update your todo details</DialogDescription>
          </DialogHeader>
          <TodoForm
            todo={todo}
            tags={tags}
            onSubmit={(updatedTodo) => {
              onUpdate(updatedTodo)
              setIsEditDialogOpen(false)
              toast.success("Todo updated successfully")
            }}
          />
        </DialogContent>
      </Dialog>

      <Dialog open={isViewDialogOpen} onOpenChange={setIsViewDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>{todo.title}</DialogTitle>
            <DialogDescription>Created on {new Date(todo.createdAt).toLocaleDateString()}</DialogDescription>
          </DialogHeader>
          {todo.image && (
            <div className="w-full h-48 overflow-hidden rounded-md mb-4">
              <img
                src={todo.image || "/placeholder.svg?height=192&width=448"}
                alt={todo.title}
                className="w-full h-full object-cover"
              />
            </div>
          )}
          <div className="space-y-4">
            <div>
              <h4 className="text-sm font-medium mb-1">Status</h4>
              <Badge
                className={cn(
                  "text-white",
                  todo.status === "pending"
                    ? "bg-yellow-500"
                    : todo.status === "in-progress"
                      ? "bg-blue-500"
                      : "bg-green-500",
                )}
              >
                {todo.status.replace("-", " ")}
              </Badge>
            </div>
            {todo.deadline && (
              <div>
                <h4 className="text-sm font-medium mb-1">Deadline</h4>
                <p className="text-sm flex items-center gap-1">
                  <Icons.calendar className="h-4 w-4" />
                  {formatDate(todo.deadline)}
                </p>
              </div>
            )}
            {todo.description && (
              <div>
                <h4 className="text-sm font-medium mb-1">Description</h4>
                <p className="text-sm">{todo.description}</p>
              </div>
            )}
            {todoTags.length > 0 && (
              <div>
                <h4 className="text-sm font-medium mb-1">Tags</h4>
                <div className="flex flex-wrap gap-2">
                  {todoTags.map((tag) => (
                    <Badge key={tag.id} style={{ backgroundColor: tag.color || "#FF5A5F" }} className="text-white">
                      {tag.name}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
            {hasAttachments && (
              <div>
                <h4 className="text-sm font-medium mb-1">Attachments</h4>
                <ul className="space-y-2">
                  {todo.attachments.map((attachment, index) => (
                    <li key={index} className="flex items-center gap-2 text-sm">
                      <Icons.paperclip className="h-4 w-4" />
                      <span>{attachment}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
            {todo.subtasks && todo.subtasks.length > 0 && (
              <div>
                <h4 className="text-sm font-medium mb-1">Subtasks</h4>
                <ul className="space-y-2">
                  {todo.subtasks.map((subtask) => (
                    <li key={subtask.id} className="flex items-center gap-2 text-sm">
                      <Icons.check className={`h-4 w-4 ${subtask.completed ? "text-green-500" : "text-gray-300"}`} />
                      <span className={subtask.completed ? "line-through text-muted-foreground" : ""}>
                        {subtask.description}
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}
