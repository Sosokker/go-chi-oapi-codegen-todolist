"use client"

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { listTodos, createTodo, updateTodoById, deleteTodoById } from "@/services/api-todos"
import { useAuth } from "@/hooks/use-auth"
import type { Todo } from "@/services/api-types"

export function useTodos(params?: { status?: string; tagId?: string }) {
  const { token } = useAuth()

  return useQuery({
    queryKey: ["todos", params],
    queryFn: () => listTodos(params, token),
    enabled: !!token,
  })
}

export function useCreateTodo() {
  const { token } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (todo: Partial<Todo>) => createTodo(todo, token!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["todos"] })
    },
    onError: (error) => {
      console.error("Failed to create todo", error)
    },
  })
}

export function useUpdateTodo() {
  const { token } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, todo }: { id: string; todo: Partial<Todo> }) => updateTodoById(id, todo, token!),
    onSuccess: (updatedTodo) => {
      queryClient.invalidateQueries({ queryKey: ["todos"] })

      // Optimistic update for the specific todo
      queryClient.setQueryData(["todos", { id: updatedTodo.id }], updatedTodo)
    },
    onError: (error) => {
      console.error("Failed to update todo", error)
    },
  })
}

export function useDeleteTodo() {
  const { token } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteTodoById(id, token!),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ["todos"] })

      // Remove the deleted todo from any cached queries
      queryClient.setQueriesData({ queryKey: ["todos"] }, (old: Todo[] | undefined) => {
        if (!old) return []
        return old.filter((todo) => todo.id !== id)
      })
    },
    onError: (error) => {
      console.error("Failed to delete todo", error)
    },
  })
}
