"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "@/hooks/use-auth";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Icons } from "@/components/icons";
import { Badge } from "@/components/ui/badge";

export function Navbar() {
  const pathname = usePathname();
  const { user, logout } = useAuth();
  const [notificationCount, setNotificationCount] = useState(3);

  const notifications = [
    {
      id: 1,
      title: "New task assigned",
      description: "You have been assigned a new task",
      time: "5 minutes ago",
      read: false,
    },
    {
      id: 2,
      title: "Task completed",
      description: "Your task 'Update documentation' has been completed",
      time: "1 hour ago",
      read: false,
    },
    {
      id: 3,
      title: "Meeting reminder",
      description: "Team meeting starts in 30 minutes",
      time: "2 hours ago",
      read: false,
    },
  ];

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const markAsRead = (id: number) => {
    setNotificationCount(Math.max(0, notificationCount - 1));
  };

  const markAllAsRead = () => {
    setNotificationCount(0);
  };

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center max-w-5xl mx-auto">
        <div className="mr-4 hidden md:flex">
          <Link href="/" className="mr-6 flex items-center space-x-2">
            <Icons.logo className="h-6 w-6 text-[#FF5A5F]" />
            <span className="hidden font-bold sm:inline-block">Todo App</span>
          </Link>
          <nav className="flex items-center space-x-6 text-sm font-medium">
            <Link
              href="/todos/list"
              className={cn(
                "transition-colors hover:text-foreground/80",
                pathname?.startsWith("/todos/list")
                  ? "text-foreground"
                  : "text-foreground/60"
              )}
            >
              <div className="flex items-center gap-1">
                <Icons.list className="h-4 w-4" />
                <span>List</span>
              </div>
            </Link>
            <Link
              href="/todos/kanban"
              className={cn(
                "transition-colors hover:text-foreground/80",
                pathname?.startsWith("/todos/kanban")
                  ? "text-foreground"
                  : "text-foreground/60"
              )}
            >
              <div className="flex items-center gap-1">
                <Icons.kanban className="h-4 w-4" />
                <span>Kanban</span>
              </div>
            </Link>
            <Link
              href="/todos/tags"
              className={cn(
                "transition-colors hover:text-foreground/80",
                pathname?.startsWith("/todos/tags")
                  ? "text-foreground"
                  : "text-foreground/60"
              )}
            >
              <div className="flex items-center gap-1">
                <Icons.tag className="h-4 w-4" />
                <span>Tags</span>
              </div>
            </Link>
          </nav>
        </div>
        <div className="flex flex-1 items-center justify-between space-x-2 md:justify-end">
          {/* Notifications Dropdown */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                className="relative h-8 w-8 rounded-full"
              >
                <Icons.bell className="h-4 w-4" />
                {notificationCount > 0 && (
                  <Badge className="absolute -top-1 -right-1 h-4 w-4 p-0 flex items-center justify-center bg-red-500 text-white text-[10px]">
                    {notificationCount}
                  </Badge>
                )}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-80">
              <DropdownMenuLabel className="flex items-center justify-between">
                <span>Notifications</span>
                {notificationCount > 0 && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={markAllAsRead}
                    className="h-auto p-0 text-xs text-muted-foreground"
                  >
                    Mark all as read
                  </Button>
                )}
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              {notifications.length === 0 ? (
                <div className="py-4 text-center text-sm text-muted-foreground">
                  No new notifications
                </div>
              ) : (
                notifications.map((notification) => (
                  <DropdownMenuItem
                    key={notification.id}
                    className="flex flex-col items-start p-4 cursor-pointer"
                    onClick={() => markAsRead(notification.id)}
                  >
                    <div className="flex w-full justify-between">
                      <span className="font-medium">{notification.title}</span>
                      <span className="text-xs text-muted-foreground">
                        {notification.time}
                      </span>
                    </div>
                    <p className="text-sm text-muted-foreground mt-1">
                      {notification.description}
                    </p>
                  </DropdownMenuItem>
                ))
              )}
              <DropdownMenuSeparator />
              <DropdownMenuItem asChild className="justify-center">
                <Link
                  href="/todos/notifications"
                  className="w-full text-center text-sm"
                >
                  View all notifications
                </Link>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                className="relative h-8 w-8 rounded-full"
              >
                <Icons.user className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>{user?.username || "User"}</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem asChild>
                <Link href="/profile">
                  <Icons.user className="mr-2 h-4 w-4" />
                  <span>Profile</span>
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild>
                <Link href="/todos/list">
                  <Icons.list className="mr-2 h-4 w-4" />
                  <span>List View</span>
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild>
                <Link href="/todos/kanban">
                  <Icons.kanban className="mr-2 h-4 w-4" />
                  <span>Kanban View</span>
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild>
                <Link href="/tags">
                  <Icons.tag className="mr-2 h-4 w-4" />
                  <span>Manage Tags</span>
                </Link>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={logout}>
                <Icons.logout className="mr-2 h-4 w-4" />
                <span>Log out</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  );
}
