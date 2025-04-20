"use client"

import { useCallback } from "react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import { useAuthStore } from "@/store/auth-store"
import { updateUserApi } from "@/services/api-auth"
import type { User, UpdateUserRequest } from "@/services/api-types"

export function useAuth() {
  const router = useRouter()
  const {
    user,
    token,
    isAuthenticated,
    login: storeLogin,
    logout: storeLogout,
    updateUser: storeUpdateUser,
  } = useAuthStore()

  const login = useCallback(
    (token: string, user: User) => {
      storeLogin(token, user)
    },
    [storeLogin],
  )

  const logout = useCallback(() => {
    storeLogout()
    toast.success("Logged out successfully")
    router.push("/login")
  }, [router, storeLogout])

  const updateProfile = useCallback(
    async (data: UpdateUserRequest) => {
      if (!token) {
        toast.error("You must be logged in to update your profile")
        return null
      }

      try {
        const updatedUser = await updateUserApi(data, token)
        storeUpdateUser(updatedUser)
        return updatedUser
      } catch (error) {
        console.error("Failed to update profile:", error)
        throw error
      }
    },
    [token, storeUpdateUser],
  )

  return {
    user,
    token,
    isAuthenticated,
    login,
    logout,
    updateProfile,
  }
}
