"use client";

import { useState } from "react";
import { toast } from "sonner";
import { motion } from "framer-motion";
import {
  useTodos,
  useCreateTodo,
  useUpdateTodo,
  useDeleteTodo,
} from "@/hooks/use-todos";
import { useTags } from "@/hooks/use-tags";
import { TodoCard } from "@/components/todo-card";
import { TodoRow } from "@/components/todo-row";
import { TodoForm } from "@/components/todo-form";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Icons } from "@/components/icons";
import type { Tag, Todo } from "@/services/api-types";

export default function TodoListPage() {
  const [status, setStatus] = useState<string | undefined>(undefined);
  const [tagFilter, setTagFilter] = useState<string | undefined>(undefined);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");

  const {
    data: todos = [],
    isLoading,
    isError,
  } = useTodos({ status, tagId: tagFilter });
  const { data: tags = [] } = useTags();
  const createTodoMutation = useCreateTodo();
  const updateTodoMutation = useUpdateTodo();
  const deleteTodoMutation = useDeleteTodo();

  const handleCreateTodo = async (todo: Partial<Todo>) => {
    try {
      await createTodoMutation.mutateAsync(todo);
      setIsCreateDialogOpen(false);
      toast.success("Todo created successfully");
    } catch {
      toast.error("Failed to create todo");
    }
  };

  const handleUpdateTodo = async (id: string, todo: Partial<Todo>) => {
    try {
      await updateTodoMutation.mutateAsync({ id, todo });
      toast.success("Todo updated successfully");
    } catch {
      toast.error("Failed to update todo");
    }
  };

  const handleDeleteTodo = async (id: string) => {
    try {
      await deleteTodoMutation.mutateAsync(id);
      toast.success("Todo deleted successfully");
    } catch {
      toast.error("Failed to delete todo");
    }
  };

  const filteredTodos = todos.filter(
    (todo) =>
      todo.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (todo.description &&
        todo.description.toLowerCase().includes(searchQuery.toLowerCase()))
  );

  if (isError) {
    return (
      <div className="container max-w-5xl mx-auto px-4 py-6">
        <div className="text-center py-10 bg-muted/20 rounded-lg border border-dashed">
          <Icons.warning className="h-12 w-12 mx-auto text-muted-foreground/50 mb-4" />
          <h3 className="text-lg font-medium text-muted-foreground mb-2">
            Error loading todos
          </h3>
          <p className="text-sm text-muted-foreground mb-4">
            Please try again later
          </p>
          <Button onClick={() => window.location.reload()} variant="outline">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="container max-w-5xl mx-auto px-4 py-6 space-y-6">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <h1 className="text-3xl font-bold tracking-tight">Todos</h1>
        <div className="flex items-center gap-2">
          <div className="flex border rounded-md overflow-hidden">
            <Button
              variant={viewMode === "grid" ? "default" : "ghost"}
              size="sm"
              className="rounded-none h-9 px-3"
              onClick={() => setViewMode("grid")}
            >
              <Icons.layoutGrid className="h-4 w-4" />
              <span className="sr-only">Grid View</span>
            </Button>
            <Button
              variant={viewMode === "list" ? "default" : "ghost"}
              size="sm"
              className="rounded-none h-9 px-3"
              onClick={() => setViewMode("list")}
            >
              <Icons.list className="h-4 w-4" />
              <span className="sr-only">List View</span>
            </Button>
          </div>
          <Dialog
            open={isCreateDialogOpen}
            onOpenChange={setIsCreateDialogOpen}
          >
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
      </div>

      <div className="flex flex-col sm:flex-row gap-4">
        <div className="flex-1">
          <Input
            placeholder="Search todos..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="max-w-md"
          />
        </div>
        <div className="flex gap-2">
          <Select
            value={tagFilter ?? "all"}
            onValueChange={(val) =>
              setTagFilter(val === "all" ? undefined : val)
            }
          >
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Filter by tag" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All tags</SelectItem>
              {tags.map((tag) => (
                <SelectItem key={tag.id} value={tag.id}>
                  {tag.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <Tabs
        defaultValue="all"
        className="w-full"
        onValueChange={(value) =>
          setStatus(value === "all" ? undefined : value)
        }
      >
        <TabsList className="grid w-full max-w-md grid-cols-3">
          <TabsTrigger value="all">All</TabsTrigger>
          <TabsTrigger value="pending">Pending</TabsTrigger>
          <TabsTrigger value="in-progress">In Progress</TabsTrigger>
        </TabsList>
        <TabsContent value="all" className="mt-6">
          <TodoList
            todos={filteredTodos}
            tags={tags}
            isLoading={isLoading}
            onUpdate={handleUpdateTodo}
            onDelete={handleDeleteTodo}
            viewMode={viewMode}
          />
        </TabsContent>
        <TabsContent value="pending" className="mt-6">
          <TodoList
            todos={filteredTodos.filter((todo) => todo.status === "pending")}
            tags={tags}
            isLoading={isLoading}
            onUpdate={handleUpdateTodo}
            onDelete={handleDeleteTodo}
            viewMode={viewMode}
          />
        </TabsContent>
        <TabsContent value="in-progress" className="mt-6">
          <TodoList
            todos={filteredTodos.filter(
              (todo) => todo.status === "in-progress"
            )}
            tags={tags}
            isLoading={isLoading}
            onUpdate={handleUpdateTodo}
            onDelete={handleDeleteTodo}
            viewMode={viewMode}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function TodoList({
  todos,
  tags,
  isLoading,
  onUpdate,
  onDelete,
  viewMode,
}: {
  todos: Todo[];
  tags: Tag[];
  isLoading: boolean;
  onUpdate: (id: string, todo: Partial<Todo>) => void;
  onDelete: (id: string) => void;
  viewMode: "grid" | "list";
}) {
  const container = {
    hidden: { opacity: 0 },
    show: {
      opacity: 1,
      transition: {
        staggerChildren: 0.05,
      },
    },
  };

  const item = {
    hidden: { opacity: 0, y: 20 },
    show: { opacity: 1, y: 0 },
  };

  if (isLoading) {
    if (viewMode === "grid") {
      return (
        <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div
              key={i}
              className="h-[150px] rounded-lg bg-muted animate-pulse"
            />
          ))}
        </div>
      );
    } else {
      return (
        <div className="space-y-2">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="h-12 rounded-md bg-muted animate-pulse" />
          ))}
        </div>
      );
    }
  }

  if (todos.length === 0) {
    return (
      <div className="text-center py-10 bg-muted/20 rounded-lg border border-dashed">
        <Icons.inbox className="h-12 w-12 mx-auto text-muted-foreground/50 mb-4" />
        <h3 className="text-lg font-medium text-muted-foreground mb-2">
          No todos found
        </h3>
        <p className="text-sm text-muted-foreground mb-4">
          Create one to get started!
        </p>
      </div>
    );
  }

  if (viewMode === "grid") {
    return (
      <motion.div
        className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3"
        variants={container}
        initial="hidden"
        animate="show"
      >
        {todos.map((todo) => (
          <motion.div key={todo.id} variants={item}>
            <TodoCard
              todo={todo}
              tags={tags}
              onUpdate={(updatedTodo) => onUpdate(todo.id, updatedTodo)}
              onDelete={() => onDelete(todo.id)}
            />
          </motion.div>
        ))}
      </motion.div>
    );
  } else {
    return (
      <motion.div
        className="space-y-1 border rounded-md overflow-hidden bg-card"
        variants={container}
        initial="hidden"
        animate="show"
      >
        {todos.map((todo) => (
          <motion.div key={todo.id} variants={item}>
            <TodoRow
              todo={todo}
              tags={tags}
              onUpdate={(updatedTodo) => onUpdate(todo.id, updatedTodo)}
              onDelete={() => onDelete(todo.id)}
            />
          </motion.div>
        ))}
      </motion.div>
    );
  }
}
