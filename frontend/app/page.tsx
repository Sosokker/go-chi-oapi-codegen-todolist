import { redirect } from "next/navigation"

export default function Home() {
  // In a real app, we'd check auth status server-side
  // For now, just redirect to todos
  redirect("/todos")
}
