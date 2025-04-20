"use client"

import { useRouter } from "next/navigation"
import { useEffect } from "react"

export default function TodosPage() {
  const router = useRouter()

  useEffect(() => {
    router.push("/todos/list")
  }, [router])

  return null
}
