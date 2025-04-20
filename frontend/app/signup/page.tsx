"use client";

import type React from "react";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import { useAuth } from "@/hooks/use-auth";
import { signupUserApi } from "@/services/api-auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Icons } from "@/components/icons";
import { Checkbox } from "@/components/ui/checkbox";
import Image from "next/image";

export default function SignupPage() {
  const router = useRouter();
  const { login } = useAuth();
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState({
    firstName: "",
    lastName: "",
    email: "",
    password: "",
  });
  const [showPassword, setShowPassword] = useState(false);
  const [agreedToTerms, setAgreedToTerms] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!agreedToTerms) {
      toast.error("You must agree to the Terms & Conditions");
      return;
    }

    setIsLoading(true);

    try {
      const { firstName, lastName, email, password } = formData;
      const username = `${firstName} ${lastName}`.trim();
      const user = await signupUserApi({ username, email, password });

      // In a real app, we'd get a token back from signup or do a separate login
      // For now, we'll simulate getting a token
      login("dummy-token", user);

      toast.success("Account created successfully");
      router.push("/todos");
    } catch (error) {
      console.error("Signup failed:", error);
      toast.error("Signup failed. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex bg-white">
      {/* Left side - Image */}
      <div className="hidden lg:flex lg:w-1/2 relative bg-[#FF5A5F]/10">
        <div className="absolute inset-0 bg-gradient-to-b from-[#FF5A5F]/20 to-[#FF5A5F]/5 z-10"></div>
        <Image
          src="/gradient-bg.jpg"
          width={1080}
          height={1080}
          alt="Background"
          className="absolute inset-0 w-full h-full object-cover brightness-60"
        />
        <div className="relative z-20 flex flex-col justify-between h-full p-12">
          <div>
            <Link href="/" className="text-[#FF5A5F] text-2xl font-bold">
              TODO
            </Link>
          </div>
          <div className="mb-12">
            <h2 className="text-[#FF5A5F] text-3xl font-bold mb-4">
              Capturing Moments,
              <br />
              Creating Memories
            </h2>
            <div className="flex space-x-2 mt-6">
              <div className="w-2 h-2 rounded-full bg-[#FF5A5F]/40"></div>
              <div className="w-2 h-2 rounded-full bg-[#FF5A5F]/20"></div>
              <div className="w-2 h-2 rounded-full bg-[#FF5A5F]"></div>
            </div>
          </div>
        </div>
      </div>

      {/* Right side - Form */}
      <div className="w-full lg:w-1/2 flex items-center justify-center p-8 bg-white">
        <div className="w-full max-w-md">
          <div className="flex justify-between items-center mb-8">
            <Link
              href="/"
              className="text-[#FF5A5F] text-2xl font-bold lg:hidden"
            >
              TODO
            </Link>
            <Link
              href="/"
              className="text-sm text-gray-500 hover:text-[#FF5A5F] transition-colors flex items-center"
            >
              <span>Back to website</span>
              <Icons.arrowRight className="ml-1 h-4 w-4" />
            </Link>
          </div>

          <h1 className="text-gray-900 text-3xl font-bold mb-2">
            Create an account
          </h1>
          <p className="text-gray-600 mb-8">
            Already have an account?{" "}
            <Link href="/login" className="text-[#FF5A5F] hover:underline">
              Log in
            </Link>
          </p>

          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="firstName" className="text-gray-900">
                  First name
                </Label>
                <Input
                  id="firstName"
                  name="firstName"
                  placeholder="John"
                  value={formData.firstName}
                  onChange={handleChange}
                  required
                  className="bg-white border-gray-300 text-gray-900 placeholder-gray-400"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="lastName" className="text-gray-900">
                  Last name
                </Label>
                <Input
                  id="lastName"
                  name="lastName"
                  placeholder="Doe"
                  value={formData.lastName}
                  onChange={handleChange}
                  required
                  className="bg-white border-gray-300 text-gray-900 placeholder-gray-400"
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="email" className="text-gray-900">
                Email
              </Label>
              <Input
                id="email"
                name="email"
                type="email"
                placeholder="name@example.com"
                value={formData.email}
                onChange={handleChange}
                required
                className="bg-white border-gray-300 text-gray-900 placeholder-gray-400"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password" className="text-gray-900">
                Password
              </Label>
              <div className="relative">
                <Input
                  id="password"
                  name="password"
                  type={showPassword ? "text" : "password"}
                  placeholder="Enter your password"
                  value={formData.password}
                  onChange={handleChange}
                  required
                  className="bg-white border-gray-300 text-gray-900 pr-10 placeholder-gray-400"
                />
                <button
                  type="button"
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400"
                  onClick={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? (
                    <Icons.eyeOff className="h-4 w-4" />
                  ) : (
                    <Icons.eye className="h-4 w-4" />
                  )}
                </button>
              </div>
            </div>

            <div className="flex items-center space-x-2">
              <Checkbox
                id="terms"
                checked={agreedToTerms}
                onCheckedChange={(checked) =>
                  setAgreedToTerms(checked as boolean)
                }
                className="border-gray-300 data-[state=checked]:bg-[#FF5A5F]"
              />
              <label
                htmlFor="terms"
                className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 text-gray-900"
              >
                I agree to the{" "}
                <Link href="#" className="text-[#FF5A5F] hover:underline">
                  Terms & Conditions
                </Link>
              </label>
            </div>

            <Button
              type="submit"
              className="w-full airbnb-button"
              disabled={isLoading}
            >
              {isLoading ? (
                <Icons.spinner className="mr-2 h-4 w-4 animate-spin" />
              ) : null}
              Create account
            </Button>

            <div className="relative flex items-center justify-center">
              <div className="border-t border-gray-200 w-full"></div>
              <span className="bg-white px-2 text-sm text-gray-400 absolute">
                Or register with
              </span>
            </div>

            <div className="grid grid-cols-1 gap-4">
              <Button
                type="button"
                variant="outline"
                className="w-full border-gray-300 text-gray-900 hover:bg-gray-100"
                onClick={() =>
                  toast.info("Google signup would be implemented here")
                }
              >
                <Icons.google className="mr-2 h-4 w-4" />
                Google
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
