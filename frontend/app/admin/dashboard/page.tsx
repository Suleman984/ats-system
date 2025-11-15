"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/lib/store";
import { jobAPI, applicationAPI, Job, Application } from "@/lib/api";
import Link from "next/link";

export default function DashboardPage() {
  const router = useRouter();
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

  useEffect(() => {
    if (!isAuthenticated) {
      router.push("/admin/login");
      return;
    }
    if (isAuthenticated) {
      fetchDashboardData();
    }
  }, [isAuthenticated, router]);

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
      <div className="flex items-center justify-center h-screen">
        Loading...
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="p-6">
        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
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
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <ActionCard
            title="Post New Job"
            description="Create a new job opening"
            link="/admin/dashboard/jobs/create"
            icon="➕"
          />
          <ActionCard
            title="View Applications"
            description="Review candidate applications"
            link="/admin/dashboard/applications"
            icon="📋"
          />
          <ActionCard
            title="Manage Jobs"
            description="Edit or close job postings"
            link="/admin/dashboard/jobs"
            icon="💼"
          />
        </div>

        {/* Recent Applications */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-bold mb-4">Recent Applications</h2>
          <div className="space-y-3">
            {recentApplications.length === 0 ? (
              <p className="text-gray-500 text-center py-4">
                No applications yet
              </p>
            ) : (
              recentApplications.map((app) => (
                <div
                  key={app.id}
                  className="flex justify-between items-center p-4 border rounded-lg"
                >
                  <div>
                    <p className="font-semibold">{app.full_name}</p>
                    <p className="text-sm text-gray-600">{app.email}</p>
                    <p className="text-xs text-gray-500">
                      Applied: {new Date(app.applied_at).toLocaleDateString()}
                    </p>
                  </div>
                  <span
                    className={`px-3 py-1 rounded-full text-sm ${
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
  const colors: Record<string, string> = {
    blue: "bg-blue-500",
    green: "bg-green-500",
    purple: "bg-purple-500",
    orange: "bg-orange-500",
  };
  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className={`w-12 h-12 ${colors[color]} rounded-lg mb-4`}></div>
      <p className="text-gray-600 text-sm">{title}</p>
      <p className="text-3xl font-bold">{value}</p>
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
    <Link
      href={link}
      className="bg-white rounded-lg shadow p-6 hover:shadow-lg transition"
    >
      <div className="text-4xl mb-3">{icon}</div>
      <h3 className="font-bold text-lg mb-2">{title}</h3>
      <p className="text-gray-600 text-sm">{description}</p>
    </Link>
  );
}
