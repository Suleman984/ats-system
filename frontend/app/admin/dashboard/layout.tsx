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
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

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

  const appMode = process.env.NEXT_PUBLIC_APP_MODE || "development";
  const isProduction = appMode === "production";

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Mode Indicator in Dashboard */}
      <div
        className={`fixed top-0 left-0 right-0 text-center py-1 text-xs font-semibold z-40 ${
          isProduction ? "bg-green-600 text-white" : "bg-yellow-500 text-black"
        }`}
      >
        {isProduction ? "🚀 PRODUCTION MODE" : "🔧 DEVELOPMENT MODE"}
      </div>
      <nav className="bg-white shadow-sm mt-7">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <Link
                href="/admin/dashboard"
                className="text-lg sm:text-xl font-bold text-blue-600"
              >
                ATS Platform
              </Link>
              {/* Desktop Navigation */}
              <div className="hidden lg:flex lg:ml-8 lg:space-x-4">
                <NavLink href="/admin/dashboard">Dashboard</NavLink>
                <NavLink href="/admin/dashboard/jobs">Jobs</NavLink>
                <NavLink href="/admin/dashboard/applications">
                  Applications
                </NavLink>
                <NavLink href="/admin/dashboard/talent-pool">
                  ⭐ Talent Pool
                </NavLink>
                <NavLink href="/admin/dashboard/find-candidates">
                  Find Candidates
                </NavLink>
                <NavLink href="/admin/dashboard/embed">Embed Code</NavLink>
                <NavLink href="/admin/dashboard/activity-logs">
                  Activity Logs
                </NavLink>
              </div>
            </div>
            {/* Desktop User Info */}
            <div className="hidden md:flex items-center space-x-4">
              <span className="text-sm text-gray-600">{user?.name}</span>
              <button
                onClick={logout}
                className="text-sm text-red-600 hover:text-red-800"
              >
                Logout
              </button>
            </div>
            {/* Mobile Menu Button */}
            <div className="flex items-center space-x-2 md:hidden">
              <span className="text-xs text-gray-600 truncate max-w-[100px]">
                {user?.name}
              </span>
              <button
                onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                className="p-2 rounded-md text-gray-600 hover:text-gray-900 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
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
          <div className="md:hidden border-t border-gray-200">
            <div className="px-2 pt-2 pb-3 space-y-1">
              <MobileNavLink
                href="/admin/dashboard"
                onClick={() => setMobileMenuOpen(false)}
              >
                Dashboard
              </MobileNavLink>
              <MobileNavLink
                href="/admin/dashboard/jobs"
                onClick={() => setMobileMenuOpen(false)}
              >
                Jobs
              </MobileNavLink>
              <MobileNavLink
                href="/admin/dashboard/applications"
                onClick={() => setMobileMenuOpen(false)}
              >
                Applications
              </MobileNavLink>
              <MobileNavLink
                href="/admin/dashboard/talent-pool"
                onClick={() => setMobileMenuOpen(false)}
              >
                ⭐ Talent Pool
              </MobileNavLink>
              <MobileNavLink
                href="/admin/dashboard/find-candidates"
                onClick={() => setMobileMenuOpen(false)}
              >
                Find Candidates
              </MobileNavLink>
              <MobileNavLink
                href="/admin/dashboard/embed"
                onClick={() => setMobileMenuOpen(false)}
              >
                Embed Code
              </MobileNavLink>
              <MobileNavLink
                href="/admin/dashboard/activity-logs"
                onClick={() => setMobileMenuOpen(false)}
              >
                Activity Logs
              </MobileNavLink>
              <div className="pt-4 border-t border-gray-200">
                <div className="px-3 py-2 text-sm text-gray-600">
                  {user?.name}
                </div>
                <button
                  onClick={() => {
                    setMobileMenuOpen(false);
                    logout();
                  }}
                  className="w-full text-left px-3 py-2 text-sm text-red-600 hover:text-red-800 hover:bg-gray-50 rounded-md"
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
      className="text-gray-700 hover:text-blue-600 px-3 py-2 rounded-md text-sm font-medium transition-colors"
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
      className="block px-3 py-2 text-base font-medium text-gray-700 hover:text-blue-600 hover:bg-gray-50 rounded-md transition-colors"
    >
      {children}
    </Link>
  );
}
