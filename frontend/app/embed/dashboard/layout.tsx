"use client";

import { refreshAdmin, useAuthStore } from "@/lib/store";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import UnreadBell from "@/components/UnreadBell";

// Embedded layout - minimal styling for iframe integration
function EmbeddedDashboardShell({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, user, logout, checkAuth, initialized } =
    useAuthStore();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [mounted, setMounted] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  // Get company_id from URL to preserve it in navigation
  const urlCompanyId = searchParams.get("company_id");

  useEffect(() => {
    setMounted(true);
    checkAuth();
    refreshAdmin().catch(() => undefined);
  }, [checkAuth]);

  const navItems = [
    ["Dashboard", "/embed/dashboard"],
    ["Jobs", "/embed/dashboard/jobs"],
    ["Applications", "/embed/dashboard/applications"],
    ["Find Candidates", "/embed/dashboard/find-candidates"],
    ["Talent Pool", "/embed/dashboard/talent-pool"],
    ["Activity", "/embed/dashboard/activity-logs"],
    ...(user?.role === "owner" ? [["Team", "/embed/dashboard/team"]] : []),
  ];

  useEffect(() => {
    if (mounted && initialized && !isAuthenticated) {
      // In embedded mode, redirect to login within the iframe with company_id
      const loginUrl = urlCompanyId
        ? `/embed/login?company_id=${urlCompanyId}`
        : "/embed/login";
      router.push(loginUrl);
    }
  }, [isAuthenticated, initialized, mounted, router, urlCompanyId]);

  const getUrlWithCompanyId = (path: string) => {
    if (urlCompanyId) {
      return `${path}${
        path.includes("?") ? "&" : "?"
      }company_id=${urlCompanyId}`;
    }
    return path;
  };

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
            <div className="flex items-center flex-1 min-w-0">
              <span className="text-base sm:text-lg font-semibold text-blue-600 truncate">
                {user?.name || "ATS Dashboard"}
              </span>
              {/* Desktop Navigation */}
              <div className="hidden md:flex md:ml-6 md:space-x-3 overflow-x-auto">
                {navItems.map(([label, path]) => (
                  <NavLink key={path} href={getUrlWithCompanyId(path)}>
                    {label}
                  </NavLink>
                ))}
              </div>
            </div>
            <div className="flex items-center gap-2">
              <UnreadBell />
            {/* Desktop Logout */}
            <div className="hidden md:block">
              <button
                onClick={logout}
                className="text-sm text-red-600 hover:text-red-800 px-3 py-1"
              >
                Logout
              </button>
            </div>
            {/* Mobile Menu Button */}
            <div className="md:hidden flex items-center">
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
        </div>
        {/* Mobile Menu */}
        {mobileMenuOpen && (
          <div className="md:hidden border-t border-gray-200">
            <div className="px-2 pt-2 pb-3 space-y-1">
              {navItems.map(([label, path]) => (
                <MobileNavLink
                  key={path}
                  href={getUrlWithCompanyId(path)}
                  onClick={() => setMobileMenuOpen(false)}
                >
                  {label}
                </MobileNavLink>
              ))}
              <div className="pt-4 border-t border-gray-200">
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
      className="text-gray-700 hover:text-blue-600 px-2 py-1 rounded text-sm font-medium transition-colors"
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

// useSearchParams needs a Suspense boundary so embedded pages can prerender.
export default function EmbeddedDashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <Suspense>
      <EmbeddedDashboardShell>{children}</EmbeddedDashboardShell>
    </Suspense>
  );
}
