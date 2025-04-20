"use client"

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { listUserTags, createTag, updateTagById, deleteTagById } from "@/services/api-tags"
import { useAuth } from "@/hooks/use-auth"
import type { Tag } from "@/services/api-types"

export function useTags() {
  const { token } = useAuth()

  return useQuery({
    queryKey: ["tags"],
    queryFn: () => listUserTags(token),
    enabled: !!token,
  })
}

export function useCreateTag() {
  const { token } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (tag: Partial<Tag>) => createTag(tag, token!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tags"] })
    },
    onError: (error) => {
      console.error("Failed to create tag", error)
    },
  })
}

export function useUpdateTag() {
  const { token } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, tag }: { id: string; tag: Partial<Tag> }) => updateTagById(id, tag, token!),
    onSuccess: (updatedTag) => {
      queryClient.invalidateQueries({ queryKey: ["tags"] })

      // Optimistic update for the specific tag
      queryClient.setQueryData(["tags", { id: updatedTag.id }], updatedTag)
    },
    onError: (error) => {
      console.error("Failed to update tag", error)
    },
  })
}

export function useDeleteTag() {
  const { token } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteTagById(id, token!),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ["tags"] })
      queryClient.invalidateQueries({ queryKey: ["todos"] }) // Invalidate todos as they may have this tag

      // Remove the deleted tag from any cached queries
      queryClient.setQueriesData({ queryKey: ["tags"] }, (old: Tag[] | undefined) => {
        if (!old) return []
        return old.filter((tag) => tag.id !== id)
      })
    },
    onError: (error) => {
      console.error("Failed to delete tag", error)
    },
  })
}
