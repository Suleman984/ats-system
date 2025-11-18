"use client";

import { useEffect, useState } from "react";
import { activityLogAPI, ActivityLog } from "@/lib/api";

export default function ActivityLogsPage() {
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState({
    action_type: "",
    entity_type: "",
    date_from: "",
    date_to: "",
  });

  useEffect(() => {
    fetchLogs();
  }, [filters]);

  const fetchLogs = async () => {
    try {
      setLoading(true);
      const params: any = {};
      if (filters.action_type) params.action_type = filters.action_type;
      if (filters.entity_type) params.entity_type = filters.entity_type;
      if (filters.date_from) params.date_from = filters.date_from;
      if (filters.date_to) params.date_to = filters.date_to;

      const response = await activityLogAPI.getAll(params);
      setLogs(response.data.logs || []);
    } catch (error) {
      console.error("Failed to fetch activity logs:", error);
    } finally {
      setLoading(false);
    }
  };

  const getActionIcon = (actionType: string) => {
    const icons: Record<string, string> = {
      company_registered: "🏢",
      job_created: "➕",
      job_updated: "✏️",
      job_deleted: "🗑️",
      job_status_changed: "🔄",
      application_shortlisted: "✅",
      application_rejected: "❌",
      application_status_changed: "🔄",
    };
    return icons[actionType] || "📝";
  };

  const getActionColor = (actionType: string) => {
    if (actionType.includes("created") || actionType.includes("registered")) {
      return "bg-green-100 text-green-800";
    }
    if (actionType.includes("deleted") || actionType.includes("rejected")) {
      return "bg-red-100 text-red-800";
    }
    if (actionType.includes("updated") || actionType.includes("changed")) {
      return "bg-blue-100 text-blue-800";
    }
    return "bg-gray-100 text-gray-800";
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading activity logs...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">Activity Logs</h1>

        {/* Filters */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <h2 className="text-lg font-semibold mb-4">Filters</h2>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div>
              <label className="block text-sm font-medium mb-1">
                Action Type
              </label>
              <select
                className="w-full px-3 py-2 border rounded-lg"
                value={filters.action_type}
                onChange={(e) =>
                  setFilters({ ...filters, action_type: e.target.value })
                }
              >
                <option value="">All Actions</option>
                <option value="company_registered">Company Registered</option>
                <option value="job_created">Job Created</option>
                <option value="job_updated">Job Updated</option>
                <option value="job_deleted">Job Deleted</option>
                <option value="job_status_changed">Job Status Changed</option>
                <option value="application_shortlisted">
                  Application Shortlisted
                </option>
                <option value="application_rejected">
                  Application Rejected
                </option>
                <option value="application_status_changed">
                  Application Status Changed
                </option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">
                Entity Type
              </label>
              <select
                className="w-full px-3 py-2 border rounded-lg"
                value={filters.entity_type}
                onChange={(e) =>
                  setFilters({ ...filters, entity_type: e.target.value })
                }
              >
                <option value="">All Entities</option>
                <option value="company">Company</option>
                <option value="job">Job</option>
                <option value="application">Application</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">
                From Date
              </label>
              <input
                type="date"
                className="w-full px-3 py-2 border rounded-lg"
                value={filters.date_from}
                onChange={(e) =>
                  setFilters({ ...filters, date_from: e.target.value })
                }
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">To Date</label>
              <input
                type="date"
                className="w-full px-3 py-2 border rounded-lg"
                value={filters.date_to}
                onChange={(e) =>
                  setFilters({ ...filters, date_to: e.target.value })
                }
              />
            </div>
          </div>
        </div>

        {/* Logs Table */}
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Time
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Action
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Admin
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Description
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {logs.length === 0 ? (
                  <tr>
                    <td
                      colSpan={4}
                      className="px-6 py-4 text-center text-gray-500"
                    >
                      No activity logs found
                    </td>
                  </tr>
                ) : (
                  logs.map((log) => (
                    <tr key={log.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {new Date(log.created_at).toLocaleString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${getActionColor(
                            log.action_type
                          )}`}
                        >
                          {getActionIcon(log.action_type)}{" "}
                          {log.action_type.replace(/_/g, " ")}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {log.admin?.name || "System"}
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500">
                        {log.description}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
