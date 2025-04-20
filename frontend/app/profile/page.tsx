"use client";

import type React from "react";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { useAuth } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Icons } from "@/components/icons";
import { Separator } from "@/components/ui/separator";

export default function ProfilePage() {
  const router = useRouter();
  const { user, updateProfile } = useAuth();
  const [isLoading, setIsLoading] = useState(false);
  const [username, setUsername] = useState(user?.username || "");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!username.trim()) {
      toast.error("Username cannot be empty");
      return;
    }

    setIsLoading(true);

    try {
      await updateProfile({ username });
      toast.success("Username updated successfully");
      router.push("/todos");
    } catch (error) {
      console.error("Failed to update profile:", error);
      toast.error("Failed to update profile. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="container max-w-2xl mx-auto px-4 py-6 space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold tracking-tight">Profile Settings</h1>
        <Button variant="outline" onClick={() => router.back()}>
          <Icons.arrowLeft className="mr-2 h-4 w-4" />
          Back
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Account Information</CardTitle>
          <CardDescription>
            Update your account information here.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4 mb-6">
            <div className="h-16 w-16 rounded-full bg-muted flex items-center justify-center">
              <Icons.user className="h-8 w-8 text-muted-foreground" />
            </div>
            <div>
              <p className="font-medium">{user?.username || "User"}</p>
              <p className="text-sm text-muted-foreground">
                {user?.email || "user@example.com"}
              </p>
            </div>
          </div>

          <Separator className="my-6" />

          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Enter your username"
              />
              <p className="text-sm text-muted-foreground">
                This is the name that will be displayed to other users.
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                value={user?.email || ""}
                disabled
                className="bg-muted"
              />
              <p className="text-sm text-muted-foreground">
                Your email address cannot be changed.
              </p>
            </div>
          </form>
        </CardContent>
        <CardFooter className="flex justify-between">
          <Button variant="outline" onClick={() => router.back()}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={
              isLoading || username === user?.username || !username.trim()
            }
            className="bg-[#FF5A5F] hover:bg-[#FF5A5F]/90"
          >
            {isLoading ? (
              <Icons.spinner className="mr-2 h-4 w-4 animate-spin" />
            ) : null}
            Save Changes
          </Button>
        </CardFooter>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Account Security</CardTitle>
          <CardDescription>
            Manage your account security settings.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex justify-between items-center">
            <div>
              <h3 className="font-medium">Password</h3>
              <p className="text-sm text-muted-foreground">
                Change your password
              </p>
            </div>
            <Button variant="outline">Change Password</Button>
          </div>
          <Separator />
          <div className="flex justify-between items-center">
            <div>
              <h3 className="font-medium">Two-Factor Authentication</h3>
              <p className="text-sm text-muted-foreground">
                Add an extra layer of security to your account
              </p>
            </div>
            <Button variant="outline">Enable</Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Danger Zone</CardTitle>
          <CardDescription>
            Irreversible and destructive actions
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex justify-between items-center">
            <div>
              <h3 className="font-medium text-destructive">Delete Account</h3>
              <p className="text-sm text-muted-foreground">
                Permanently delete your account and all of your data
              </p>
            </div>
            <Button variant="destructive">Delete Account</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
