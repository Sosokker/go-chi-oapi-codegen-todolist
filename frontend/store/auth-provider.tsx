"use client"

import type React from "react"

import { useEffect } from "react"
import { useAuthStore } from "./auth-store"

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { initialize } = useAuthStore()

  useEffect(() => {
    // Initialize auth state from localStorage (if available)
    initialize()
  }, [initialize])

  return <>{children}</>
}
