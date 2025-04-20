import { Skeleton } from "@/components/ui/skeleton"

export default function Loading() {
  return (
    <div className="container max-w-5xl mx-auto px-4 py-6 space-y-6">
      <div className="flex justify-between items-center">
        <Skeleton className="h-10 w-40" />
        <Skeleton className="h-10 w-32" />
      </div>

      <div className="flex flex-col sm:flex-row gap-4">
        <Skeleton className="h-10 flex-1 max-w-md" />
        <Skeleton className="h-10 w-40" />
      </div>

      <Skeleton className="h-12 w-full max-w-md" />

      <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
        {Array(6)
          .fill(0)
          .map((_, i) => (
            <Skeleton key={i} className="h-[180px] rounded-lg" />
          ))}
      </div>
    </div>
  )
}
