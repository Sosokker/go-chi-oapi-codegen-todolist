"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Icons } from "@/components/icons";
import { Badge } from "@/components/ui/badge";

// Mock notification data - in a real app, this would come from an API
const mockNotifications = [
  {
    id: 1,
    title: "New task assigned",
    description: "You have been assigned a new task 'Update documentation'",
    time: "5 minutes ago",
    read: false,
    type: "task",
  },
  {
    id: 2,
    title: "Task completed",
    description: "Your task 'Update documentation' has been completed",
    time: "1 hour ago",
    read: false,
    type: "task",
  },
  {
    id: 3,
    title: "Meeting reminder",
    description: "Team meeting starts in 30 minutes",
    time: "2 hours ago",
    read: false,
    type: "reminder",
  },
  {
    id: 4,
    title: "New comment",
    description: "John commented on your task 'Design homepage'",
    time: "1 day ago",
    read: true,
    type: "comment",
  },
  {
    id: 5,
    title: "Task deadline approaching",
    description: "Task 'Finalize project proposal' is due tomorrow",
    time: "1 day ago",
    read: true,
    type: "reminder",
  },
  {
    id: 6,
    title: "System update",
    description: "The system will be updated tonight at 2 AM",
    time: "2 days ago",
    read: true,
    type: "system",
  },
];

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState(mockNotifications);
  const [activeTab, setActiveTab] = useState("all");
  const [isLoading, setIsLoading] = useState(false);

  const unreadCount = notifications.filter((n) => !n.read).length;

  const filteredNotifications = notifications.filter((notification) => {
    if (activeTab === "all") return true;
    if (activeTab === "unread") return !notification.read;
    if (activeTab === "read") return notification.read;
    return true;
  });

  const markAsRead = (id: number) => {
    setNotifications(
      notifications.map((notification) =>
        notification.id === id ? { ...notification, read: true } : notification
      )
    );
    toast.success("Notification marked as read");
  };

  const markAllAsRead = () => {
    setIsLoading(true);
    // Simulate API call
    setTimeout(() => {
      setNotifications(
        notifications.map((notification) => ({ ...notification, read: true }))
      );
      setIsLoading(false);
      toast.success("All notifications marked as read");
    }, 500);
  };

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case "task":
        return <Icons.list className="h-5 w-5 text-blue-500" />;
      case "reminder":
        return <Icons.calendar className="h-5 w-5 text-yellow-500" />;
      case "comment":
        return <Icons.messageCircle className="h-5 w-5 text-green-500" />;
      case "system":
        return <Icons.settings className="h-5 w-5 text-purple-500" />;
      default:
        return <Icons.bell className="h-5 w-5 text-gray-500" />;
    }
  };

  return (
    <div className="container max-w-4xl mx-auto px-4 py-6 space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Notifications</h1>
          <p className="text-muted-foreground mt-1">
            You have {unreadCount} unread notification{unreadCount !== 1 && "s"}
          </p>
        </div>
        {unreadCount > 0 && (
          <Button
            onClick={markAllAsRead}
            variant="outline"
            disabled={isLoading}
            className="flex items-center gap-2"
          >
            {isLoading ? (
              <Icons.spinner className="h-4 w-4 animate-spin" />
            ) : (
              <Icons.check className="h-4 w-4" />
            )}
            Mark all as read
          </Button>
        )}
      </div>

      <Tabs
        defaultValue="all"
        value={activeTab}
        onValueChange={setActiveTab}
        className="w-full"
      >
        <TabsList className="grid w-full max-w-md grid-cols-3">
          <TabsTrigger value="all">
            All
            <Badge className="ml-2 bg-gray-500">{notifications.length}</Badge>
          </TabsTrigger>
          <TabsTrigger value="unread">
            Unread
            <Badge className="ml-2 bg-blue-500">{unreadCount}</Badge>
          </TabsTrigger>
          <TabsTrigger value="read">
            Read
            <Badge className="ml-2 bg-green-500">
              {notifications.length - unreadCount}
            </Badge>
          </TabsTrigger>
        </TabsList>

        <TabsContent value={activeTab} className="mt-6">
          {filteredNotifications.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-10">
                <Icons.inbox className="h-12 w-12 text-muted-foreground/50 mb-4" />
                <h3 className="text-lg font-medium text-muted-foreground mb-2">
                  No notifications
                </h3>
                <p className="text-sm text-muted-foreground mb-4">
                  {activeTab === "unread"
                    ? "You have no unread notifications"
                    : activeTab === "read"
                    ? "You have no read notifications"
                    : "You have no notifications"}
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-4">
              {filteredNotifications.map((notification) => (
                <Card
                  key={notification.id}
                  className={`transition-colors ${
                    !notification.read ? "border-l-4 border-l-blue-500" : ""
                  }`}
                >
                  <CardContent className="p-4 flex gap-4">
                    <div className="flex-shrink-0 mt-1">
                      {getNotificationIcon(notification.type)}
                    </div>
                    <div className="flex-grow">
                      <div className="flex justify-between items-start">
                        <CardTitle className="text-base">
                          {notification.title}
                        </CardTitle>
                        <div className="flex items-center gap-2">
                          <span className="text-xs text-muted-foreground">
                            {notification.time}
                          </span>
                          {!notification.read && (
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-7 w-7"
                              onClick={() => markAsRead(notification.id)}
                            >
                              <Icons.check className="h-4 w-4" />
                              <span className="sr-only">Mark as read</span>
                            </Button>
                          )}
                        </div>
                      </div>
                      <CardDescription className="mt-1">
                        {notification.description}
                      </CardDescription>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
