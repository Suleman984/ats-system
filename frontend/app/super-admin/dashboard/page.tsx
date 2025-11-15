"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useSuperAdminStore } from "@/lib/store";
import { superAdminAPI, SuperAdminStats } from "@/lib/api";
import Link from "next/link";
export default function SuperAdminDashboardPage() {
  const router = useRouter();
  const { isAuthenticated } = useSuperAdminStore();
  const [stats, setStats] = useState<SuperAdminStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!isAuthenticated) {
      router.push("/super-admin/login");
      return;
    }
    fetchStats();
  }, [isAuthenticated, router]);

  const fetchStats = async () => {
    try {
      const response = await superAdminAPI.getStats();
      setStats(response.data.stats);
    } catch (error) {
      console.error("Failed to fetch stats:", error);
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
        <h1 className="text-3xl font-bold mb-6">Platform Overview</h1>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <StatsCard
            title="Total Companies"
            value={stats?.total_companies || 0}
            color="purple"
            icon="🏢"
          />
          <StatsCard
            title="Active Companies"
            value={stats?.active_companies || 0}
            color="green"
            icon="✅"
          />
          <StatsCard
            title="Total Jobs"
            value={stats?.total_jobs || 0}
            color="blue"
            icon="💼"
          />
          <StatsCard
            title="Open Jobs"
            value={stats?.open_jobs || 0}
            color="orange"
            icon="📋"
          />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <StatsCard
            title="Total Applications"
            value={stats?.total_applications || 0}
            color="indigo"
            icon="📝"
          />
          <StatsCard
            title="Pending"
            value={stats?.pending_applications || 0}
            color="yellow"
            icon="⏳"
          />
          <StatsCard
            title="Shortlisted"
            value={stats?.shortlisted_applications || 0}
            color="green"
            icon="⭐"
          />
          <StatsCard
            title="Total Admins"
            value={stats?.total_admins || 0}
            color="pink"
            icon="👥"
          />
        </div>

        {/* Quick Actions */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-bold mb-4">Quick Actions</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Link
              href="/super-admin/dashboard/companies"
              className="p-4 border rounded-lg hover:bg-gray-50"
            >
              <h3 className="font-semibold mb-2">View All Companies</h3>
              <p className="text-sm text-gray-600">
                See all registered companies and their statistics
              </p>
            </Link>
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
  icon,
}: {
  title: string;
  value: number;
  color: string;
  icon: string;
}) {
  const colors: Record<string, string> = {
    purple: "bg-purple-500",
    green: "bg-green-500",
    blue: "bg-blue-500",
    orange: "bg-orange-500",
    indigo: "bg-indigo-500",
    yellow: "bg-yellow-500",
    pink: "bg-pink-500",
  };
  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="flex items-center justify-between mb-4">
        <div
          className={`w-12 h-12 ${colors[color]} rounded-lg flex items-center justify-center text-2xl`}
        >
          {icon}
        </div>
      </div>
      <p className="text-gray-600 text-sm">{title}</p>
      <p className="text-3xl font-bold">{value}</p>
    </div>
  );
}
