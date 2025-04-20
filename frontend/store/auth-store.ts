import { create } from "zustand"
import { persist } from "zustand/middleware"
import { validateToken } from "@/services/api-auth"
import type { User } from "@/services/api-types"

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  hydrated: boolean
  initialize: () => void
  login: (token: string, user: User) => void
  logout: () => void
  updateUser: (user: User) => void
  setHydrated: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      hydrated: false,
      initialize: () => {
        const token = localStorage.getItem("token")
        if (token) {
          validateToken(token).then((isValid) => {
            if (isValid) {
              set({ token, isAuthenticated: true })
            } else {
              set({ token: null, isAuthenticated: false })
            }
          })
        }
      },
      login: (token, user) => set({ token, user, isAuthenticated: true }),
      logout: () => set({ token: null, user: null, isAuthenticated: false }),
      updateUser: (user) => set({ user }),
      setHydrated: () => set({ hydrated: true }),
    }),
    {
      name: "auth-storage",
      onRehydrateStorage: () => {
        return async (state) => {
          const token = state?.token
          if (token) {
            const isValid = await validateToken(token)
            if (!isValid) {
              state?.logout?.()
            }
          }
          state?.setHydrated?.()
        }
      },
    }
  )
)