"use client"

import { useEffect } from "react"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Icons } from "@/components/icons"

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    // Log the error to an error reporting service
    console.error(error)
  }, [error])

  return (
    <div className="flex flex-col items-center justify-center min-h-screen px-6 py-12 bg-background">
      <div className="flex flex-col items-center max-w-md mx-auto text-center">
        <div className="flex items-center justify-center w-16 h-16 rounded-full bg-red-100 dark:bg-red-900/20 mb-6">
          <Icons.warning className="h-8 w-8 text-red-600 dark:text-red-400" />
        </div>
        <h1 className="text-2xl font-bold tracking-tight mb-4">Something went wrong!</h1>
        <p className="text-muted-foreground mb-8">
          We&apos;re sorry, but we encountered an unexpected error. Our team has been notified and is working to fix the
          issue.
        </p>
        <div className="flex flex-col sm:flex-row gap-4">
          <Button onClick={reset}>
            <Icons.refresh className="mr-2 h-4 w-4" />
            Try again
          </Button>
          <Button variant="outline" asChild>
            <Link href="/">
              <Icons.home className="mr-2 h-4 w-4" />
              Go to Home
            </Link>
          </Button>
        </div>
      </div>
    </div>
  )
}
