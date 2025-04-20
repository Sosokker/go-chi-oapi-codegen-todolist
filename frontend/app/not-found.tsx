import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Icons } from "@/components/icons"

export default function NotFound() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen px-6 py-12 bg-background">
      <div className="flex flex-col items-center max-w-md mx-auto text-center">
        <div className="flex items-center justify-center w-16 h-16 rounded-full bg-muted mb-6">
          <Icons.warning className="h-8 w-8 text-muted-foreground" />
        </div>
        <h1 className="text-4xl font-bold tracking-tight mb-2">404</h1>
        <h2 className="text-2xl font-semibold mb-4">Page not found</h2>
        <p className="text-muted-foreground mb-8">
          Sorry, we couldn&apos;t find the page you&apos;re looking for. It might have been moved, deleted, or never existed.
        </p>
        <div className="flex flex-col sm:flex-row gap-4">
          <Button asChild>
            <Link href="/">
              <Icons.home className="mr-2 h-4 w-4" />
              Go to Home
            </Link>
          </Button>
          <Button variant="outline" asChild>
            <Link href="/todos">
              <Icons.list className="mr-2 h-4 w-4" />
              View Todos
            </Link>
          </Button>
        </div>
      </div>
    </div>
  )
}
