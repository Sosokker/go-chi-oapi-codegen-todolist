import type React from "react";
import { Inter } from "next/font/google";
import { Toaster } from "sonner";
import { QueryClientProvider } from "@/components/query-client-provider";
import { ThemeProvider } from "@/components/theme-provider";
import { AuthProvider } from "@/store/auth-provider";
import "@/app/globals.css";

const inter = Inter({ subsets: ["latin"], variable: "--font-sans" });

export const metadata = {
  title: "Todo App",
  description: "A beautiful todo app with Airbnb-inspired design",
  generator: "v0.dev",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`font-sans ${inter.variable}`}>
        <ThemeProvider attribute="class" defaultTheme="light">
          <QueryClientProvider>
            <AuthProvider>
              {children}
              <Toaster position="top-right" />
            </AuthProvider>
          </QueryClientProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
