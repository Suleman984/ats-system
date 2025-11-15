"use client";

import { useAuthStore } from "@/lib/store";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import Link from "next/link";

// Embedded layout - minimal styling for iframe integration
export default function EmbeddedDashboardLayout({
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
    checkAuth();
  }, [checkAuth]);

  useEffect(() => {
    if (mounted && initialized && !isAuthenticated) {
      // In embedded mode, redirect to login within the iframe
      router.push("/embed/login");
    }
  }, [isAuthenticated, initialized, mounted, router]);

  if (!mounted || !initialized || !isAuthenticated) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Minimal navigation for embedded mode */}
      <nav className="bg-white shadow-sm border-b">
        <div className="px-4">
          <div className="flex justify-between items-center h-14">
            <div className="flex items-center space-x-6">
              <span className="text-lg font-semibold text-blue-600">
                {user?.name || "ATS Dashboard"}
              </span>
              <div className="hidden md:flex space-x-3">
                <NavLink href="/embed/dashboard">Dashboard</NavLink>
                <NavLink href="/embed/dashboard/jobs">Jobs</NavLink>
                <NavLink href="/embed/dashboard/applications">
                  Applications
                </NavLink>
              </div>
            </div>
            <button
              onClick={logout}
              className="text-sm text-red-600 hover:text-red-800 px-3 py-1"
            >
              Logout
            </button>
          </div>
        </div>
      </nav>
      <main className="p-4">{children}</main>
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
      className="text-gray-700 hover:text-blue-600 px-2 py-1 rounded text-sm font-medium"
    >
      {children}
    </Link>
  );
}
