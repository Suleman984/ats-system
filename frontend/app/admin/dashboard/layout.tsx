"use client";

import { useAuthStore } from "@/lib/store";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import Link from "next/link";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, user, logout, checkAuth, initialized } =
    useAuthStore();
  const router = useRouter();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    // Check auth on mount (client-side only)
    checkAuth();
  }, [checkAuth]);

  useEffect(() => {
    // Only redirect after initialization and mounting
    if (mounted && initialized && !isAuthenticated) {
      router.push("/admin/login");
    }
  }, [isAuthenticated, initialized, mounted, router]);

  // Show nothing while checking auth (prevents hydration mismatch)
  if (!mounted || !initialized || !isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-6">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-8">
              <Link
                href="/admin/dashboard"
                className="text-xl font-bold text-blue-600"
              >
                ATS Platform
              </Link>
              <div className="hidden md:flex space-x-4">
                <NavLink href="/admin/dashboard">Dashboard</NavLink>
                <NavLink href="/admin/dashboard/jobs">Jobs</NavLink>
                <NavLink href="/admin/dashboard/applications">
                  Applications
                </NavLink>
                <NavLink href="/admin/dashboard/embed">Embed Code</NavLink>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-600">{user?.name}</span>
              <button
                onClick={logout}
                className="text-sm text-red-600 hover:text-red-800"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>
      <main>{children}</main>
    </div>
  );
}

function NavLink({
  href,
  children,
}: {
  href: string;
  children: React.ReactNode;
}) {
  return (
    <Link
      href={href}
      className="text-gray-700 hover:text-blue-600 px-3 py-2 rounded-md text-sm font-medium"
    >
      {children}
    </Link>
  );
}
