"use client";

import { useDroppable } from "@dnd-kit/core";
import { motion, AnimatePresence } from "framer-motion";
import { TodoCard } from "@/components/todo-card";
import type { Todo, Tag } from "@/services/api-types";

interface KanbanColumnProps {
  id: string;
  title: string;
  todos: Todo[];
  tags: Tag[];
  isLoading: boolean;
}

export function KanbanColumn({
  id,
  title,
  todos,
  tags,
  isLoading,
}: KanbanColumnProps) {
  const { setNodeRef, isOver } = useDroppable({
    id,
  });

  const getColumnColor = () => {
    switch (id) {
      case "pending":
        return "border-yellow-200 bg-yellow-50 dark:bg-yellow-950/20 dark:border-yellow-900/50";
      case "in-progress":
        return "border-blue-200 bg-blue-50 dark:bg-blue-950/20 dark:border-blue-900/50";
      case "completed":
        return "border-green-200 bg-green-50 dark:bg-green-950/20 dark:border-green-900/50";
      default:
        return "border-gray-200 bg-gray-50 dark:bg-gray-800/20 dark:border-gray-700";
    }
  };

  return (
    <div
      ref={setNodeRef}
      className={`flex flex-col h-[calc(100vh-200px)] rounded-lg border p-2 transition-colors ${getColumnColor()} ${
        isOver ? "ring-2 ring-primary ring-opacity-50" : ""
      }`}
    >
      <div className="flex items-center justify-between mb-2 px-1">
        <h3 className="font-medium text-sm uppercase tracking-wider">
          {title}
        </h3>
        <div className="bg-background text-muted-foreground rounded-full px-2 py-0.5 text-xs font-medium">
          {todos.length}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto space-y-2 pr-1">
        {isLoading ? (
          Array(3)
            .fill(0)
            .map((_, i) => (
              <div
                key={i}
                className="h-[100px] rounded-lg bg-muted animate-pulse"
              />
            ))
        ) : todos.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-24 text-center py-4 px-2 border border-dashed rounded-md bg-background/50">
            <p className="text-muted-foreground text-xs">
              No todos in this column
            </p>
            <p className="text-xs text-muted-foreground/70 mt-1">
              Drag and drop tasks here
            </p>
          </div>
        ) : (
          <AnimatePresence>
            {todos.map((todo) => (
              <motion.div
                key={todo.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.9 }}
                transition={{ duration: 0.2 }}
              >
                <TodoCard
                  todo={todo}
                  tags={tags}
                  onUpdate={() => {}}
                  onDelete={() => {}}
                  isDraggable={true}
                />
              </motion.div>
            ))}
          </AnimatePresence>
        )}
      </div>
    </div>
  );
}
