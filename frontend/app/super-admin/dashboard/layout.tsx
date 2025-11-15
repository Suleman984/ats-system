"use client";

import { useSuperAdminStore } from "@/lib/store";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import Link from "next/link";

export default function SuperAdminDashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, user, logout, checkAuth, initialized } =
    useSuperAdminStore();
  const router = useRouter();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    checkAuth();
  }, [checkAuth]);

  useEffect(() => {
    if (mounted && initialized && !isAuthenticated) {
      router.push("/super-admin/login");
    }
  }, [isAuthenticated, initialized, mounted, router]);

  if (!mounted || !initialized || !isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-purple-600 text-white shadow-sm">
        <div className="max-w-7xl mx-auto px-6">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-8">
              <Link href="/super-admin/dashboard" className="text-xl font-bold">
                ATS Platform - Super Admin
              </Link>
              <div className="hidden md:flex space-x-4">
                <NavLink href="/super-admin/dashboard">Dashboard</NavLink>
                <NavLink href="/super-admin/dashboard/companies">
                  Companies
                </NavLink>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm">{user?.name}</span>
              <button onClick={logout} className="text-sm hover:underline">
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
      className="hover:bg-purple-700 px-3 py-2 rounded-md text-sm font-medium"
    >
      {children}
    </Link>
  );
}
