"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/hooks/use-auth";
import { getCurrentUser } from "@/services/api-auth";
import { toast } from "sonner";

export default function OAuthCallbackPage() {
  const router = useRouter();
  const { login } = useAuth();

  useEffect(() => {
    let token: string | null = null;
    if (typeof window !== "undefined") {
      const urlParams = new URLSearchParams(window.location.search);
      token = urlParams.get("access_token");

      // If not in query, check hash fragment
      if (!token && window.location.hash) {
        const hashParams = new URLSearchParams(
          window.location.hash.substring(1)
        );
        token = hashParams.get("access_token");
      }
    }

    if (token) {
      localStorage.setItem("access_token", token);

      async function fetchUser() {
        try {
          const user = await getCurrentUser(token!);
          login(token!, user);
          toast.success("Logged in with Google!");
          router.replace("/todos");
        } catch (err) {
          console.error(err);
          toast.error("Google login failed");
          router.replace("/login");
        }
      }
      fetchUser();
    } else {
      toast.error("No access token found");
      router.replace("/login");
    }
  }, [login, router]);

  return (
    <div className="flex items-center justify-center min-h-screen">
      <span>Logging you in...</span>
    </div>
  );
}
