"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuthStore } from "@/lib/store";
import { jobAPI, applicationAPI, Application } from "@/lib/api";

export default function EmbeddedDashboardPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { isAuthenticated, user } = useAuthStore();
  const [stats, setStats] = useState({
    totalJobs: 0,
    openJobs: 0,
    totalApplications: 0,
    shortlisted: 0,
  });
  const [recentApplications, setRecentApplications] = useState<Application[]>(
    []
  );
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Get company_id from URL
  const urlCompanyId = searchParams.get("company_id");

  useEffect(() => {
    // Validate company_id is present in URL
    if (!urlCompanyId) {
      setError(
        "Invalid embed code: Company ID is missing. Please use the embed code from your dashboard."
      );
      setLoading(false);
      return;
    }

    // If not authenticated, redirect to login with company_id
    if (!isAuthenticated) {
      router.push(`/embed/login?company_id=${urlCompanyId}`);
      return;
    }

    // Validate that logged-in user's company_id matches URL company_id
    if (user?.company_id !== urlCompanyId) {
      setError(
        "Security Error: The embed code does not match your account. Please log out and use the correct embed code from your dashboard."
      );
      setLoading(false);
      return;
    }

    // All validations passed, fetch dashboard data
    if (isAuthenticated && user?.company_id === urlCompanyId) {
      fetchDashboardData();
    }
  }, [isAuthenticated, user, urlCompanyId, router]);

  const fetchDashboardData = async () => {
    try {
      const [jobsRes, appsRes] = await Promise.all([
        jobAPI.getAll(),
        applicationAPI.getAll(),
      ]);
      const jobs = jobsRes.data.jobs;
      const applications = appsRes.data.applications;

      setStats({
        totalJobs: jobs.length,
        openJobs: jobs.filter((j) => j.status === "open").length,
        totalApplications: applications.length,
        shortlisted: applications.filter((a) => a.status === "shortlisted")
          .length,
      });
      setRecentApplications(applications.slice(0, 5));
    } catch (error) {
      console.error("Failed to fetch dashboard data:", error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64 p-4">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6 max-w-md">
          <h3 className="text-red-800 font-bold mb-2">⚠️ Security Error</h3>
          <p className="text-red-700 text-sm mb-4">{error}</p>
          <button
            onClick={() => {
              useAuthStore.getState().logout();
              router.push(`/embed/login?company_id=${urlCompanyId || ""}`);
            }}
            className="bg-red-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-red-700"
          >
            Go to Login
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="p-4">
      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <StatsCard title="Total Jobs" value={stats.totalJobs} color="blue" />
        <StatsCard title="Open Jobs" value={stats.openJobs} color="green" />
        <StatsCard
          title="Applications"
          value={stats.totalApplications}
          color="purple"
        />
        <StatsCard
          title="Shortlisted"
          value={stats.shortlisted}
          color="orange"
        />
      </div>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <ActionCard
          title="Post New Job"
          description="Create a new job opening"
          link="/embed/dashboard/jobs/create"
          icon="➕"
        />
        <ActionCard
          title="View Applications"
          description="Review candidate applications"
          link="/embed/dashboard/applications"
          icon="📋"
        />
        <ActionCard
          title="Manage Jobs"
          description="Edit or close job postings"
          link="/embed/dashboard/jobs"
          icon="💼"
        />
      </div>

      {/* Recent Applications */}
      <div className="bg-white rounded-lg shadow p-4">
        <h2 className="text-lg font-bold mb-3">Recent Applications</h2>
        <div className="space-y-2">
          {recentApplications.length === 0 ? (
            <p className="text-gray-500 text-center py-4 text-sm">
              No applications yet
            </p>
          ) : (
            recentApplications.map((app) => (
              <div
                key={app.id}
                className="flex justify-between items-center p-3 border rounded-lg"
              >
                <div>
                  <p className="font-semibold text-sm">{app.full_name}</p>
                  <p className="text-xs text-gray-600">{app.email}</p>
                  <p className="text-xs text-gray-500">
                    Applied: {new Date(app.applied_at).toLocaleDateString()}
                  </p>
                </div>
                <span
                  className={`px-2 py-1 rounded-full text-xs ${
                    app.status === "shortlisted"
                      ? "bg-green-100 text-green-800"
                      : app.status === "rejected"
                      ? "bg-red-100 text-red-800"
                      : "bg-yellow-100 text-yellow-800"
                  }`}
                >
                  {app.status}
                </span>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}

function StatsCard({
  title,
  value,
  color,
}: {
  title: string;
  value: number;
  color: string;
}) {
  const colors: Record<string, { bg: string; icon: string }> = {
    blue: { bg: "bg-blue-500", icon: "💼" },
    green: { bg: "bg-green-500", icon: "✅" },
    purple: { bg: "bg-purple-500", icon: "📋" },
    orange: { bg: "bg-orange-500", icon: "⭐" },
  };

  const colorConfig = colors[color] || colors.blue;

  return (
    <div className="bg-white rounded-lg shadow p-4">
      <div
        className={`w-10 h-10 ${colorConfig.bg} rounded-lg mb-3 flex items-center justify-center text-xl`}
      >
        {colorConfig.icon}
      </div>
      <p className="text-gray-600 text-xs">{title}</p>
      <p className="text-2xl font-bold">{value}</p>
    </div>
  );
}

function ActionCard({
  title,
  description,
  link,
  icon,
}: {
  title: string;
  description: string;
  link: string;
  icon: string;
}) {
  return (
    <a
      href={link}
      className="bg-white rounded-lg shadow p-4 hover:shadow-lg transition block"
    >
      <div className="text-3xl mb-2">{icon}</div>
      <h3 className="font-bold text-base mb-1">{title}</h3>
      <p className="text-gray-600 text-xs">{description}</p>
    </a>
  );
}
