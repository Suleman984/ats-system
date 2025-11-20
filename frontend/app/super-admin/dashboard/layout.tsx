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
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

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
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center flex-1 min-w-0">
              <Link
                href="/super-admin/dashboard"
                className="text-lg sm:text-xl font-bold truncate"
              >
                ATS Platform - Super Admin
              </Link>
              {/* Desktop Navigation */}
              <div className="hidden lg:flex lg:ml-8 lg:space-x-4">
                <NavLink href="/super-admin/dashboard">Dashboard</NavLink>
                <NavLink href="/super-admin/dashboard/companies">
                  Companies
                </NavLink>
                <NavLink href="/super-admin/dashboard/activity-logs">
                  Activity Logs
                </NavLink>
              </div>
            </div>
            {/* Desktop User Info */}
            <div className="hidden md:flex items-center space-x-4">
              <span className="text-sm">{user?.name}</span>
              <button onClick={logout} className="text-sm hover:underline">
                Logout
              </button>
            </div>
            {/* Mobile Menu Button */}
            <div className="flex items-center space-x-2 md:hidden">
              <span className="text-xs truncate max-w-[100px]">
                {user?.name}
              </span>
              <button
                onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                className="p-2 rounded-md text-white hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white"
                aria-label="Toggle menu"
              >
                {mobileMenuOpen ? (
                  <svg
                    className="h-6 w-6"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                ) : (
                  <svg
                    className="h-6 w-6"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M4 6h16M4 12h16M4 18h16"
                    />
                  </svg>
                )}
              </button>
            </div>
          </div>
        </div>
        {/* Mobile Menu */}
        {mobileMenuOpen && (
          <div className="md:hidden border-t border-purple-500">
            <div className="px-2 pt-2 pb-3 space-y-1">
              <MobileNavLink
                href="/super-admin/dashboard"
                onClick={() => setMobileMenuOpen(false)}
              >
                Dashboard
              </MobileNavLink>
              <MobileNavLink
                href="/super-admin/dashboard/companies"
                onClick={() => setMobileMenuOpen(false)}
              >
                Companies
              </MobileNavLink>
              <MobileNavLink
                href="/super-admin/dashboard/activity-logs"
                onClick={() => setMobileMenuOpen(false)}
              >
                Activity Logs
              </MobileNavLink>
              <div className="pt-4 border-t border-purple-500">
                <div className="px-3 py-2 text-sm">{user?.name}</div>
                <button
                  onClick={() => {
                    setMobileMenuOpen(false);
                    logout();
                  }}
                  className="w-full text-left px-3 py-2 text-sm hover:bg-purple-700 rounded-md"
                >
                  Logout
                </button>
              </div>
            </div>
          </div>
        )}
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
      className="hover:bg-purple-700 px-3 py-2 rounded-md text-sm font-medium transition-colors"
    >
      {children}
    </Link>
  );
}

function MobileNavLink({
  href,
  children,
  onClick,
}: {
  href: string;
  children: React.ReactNode;
  onClick: () => void;
}) {
  return (
    <Link
      href={href}
      onClick={onClick}
      className="block px-3 py-2 text-base font-medium text-white hover:bg-purple-700 rounded-md transition-colors"
    >
      {children}
    </Link>
  );
}
