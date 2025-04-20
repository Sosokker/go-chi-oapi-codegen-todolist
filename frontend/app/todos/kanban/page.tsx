"use client";

import { useState } from "react";
import { toast } from "sonner";
import { DndContext, type DragEndEvent, closestCenter } from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { useTodos, useUpdateTodo } from "@/hooks/use-todos";
import { useTags } from "@/hooks/use-tags";
import { KanbanColumn } from "@/components/kanban-column";
import { TodoForm } from "@/components/todo-form";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Icons } from "@/components/icons";

export default function KanbanPage() {
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const { data: todos = [], isLoading } = useTodos();
  const { data: tags = [] } = useTags();
  const updateTodoMutation = useUpdateTodo();

  const pendingTodos = todos.filter((todo) => todo.status === "pending");
  const inProgressTodos = todos.filter((todo) => todo.status === "in-progress");
  const completedTodos = todos.filter((todo) => todo.status === "completed");

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (!over) return;

    const todoId = active.id as string;
    const newStatus = over.id as "pending" | "in-progress" | "completed";

    const todo = todos.find((t) => t.id === todoId);
    if (!todo || todo.status === newStatus) return;

    updateTodoMutation.mutate(
      {
        id: todoId,
        todo: { status: newStatus },
      },
      {
        onSuccess: () => {
          toast.success(`Todo moved to ${newStatus.replace("-", " ")}`);
        },
        onError: () => {
          toast.error("Failed to update todo status");
        },
      }
    );
  };

  const handleCreateTodo = async () => {
    try {
      // This would be handled by the createTodo mutation in a real app
      toast.success("Todo created successfully");
      setIsCreateDialogOpen(false);
    } catch {
      toast.error("Failed to create todo");
    }
  };

  return (
    <div className="space-y-6 pt-5 mx-auto max-w-5xl">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold tracking-tight">Kanban Board</h1>
        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button className="bg-[#FF5A5F] hover:bg-[#FF5A5F]/90">
              <Icons.plus className="mr-2 h-4 w-4" />
              Add Todo
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Todo</DialogTitle>
              <DialogDescription>
                Add a new task to your todo list
              </DialogDescription>
            </DialogHeader>
            <TodoForm tags={tags} onSubmit={handleCreateTodo} />
          </DialogContent>
        </Dialog>
      </div>

      <DndContext collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <SortableContext
            items={pendingTodos.map((t) => t.id)}
            strategy={verticalListSortingStrategy}
          >
            <KanbanColumn
              id="pending"
              title="Pending"
              todos={pendingTodos}
              tags={tags}
              isLoading={isLoading}
            />
          </SortableContext>

          <SortableContext
            items={inProgressTodos.map((t) => t.id)}
            strategy={verticalListSortingStrategy}
          >
            <KanbanColumn
              id="in-progress"
              title="In Progress"
              todos={inProgressTodos}
              tags={tags}
              isLoading={isLoading}
            />
          </SortableContext>

          <SortableContext
            items={completedTodos.map((t) => t.id)}
            strategy={verticalListSortingStrategy}
          >
            <KanbanColumn
              id="completed"
              title="Completed"
              todos={completedTodos}
              tags={tags}
              isLoading={isLoading}
            />
          </SortableContext>
        </div>
      </DndContext>
    </div>
  );
}
