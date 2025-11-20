"use client";

import { useEffect, useState } from "react";
import { crmAPI, Application } from "@/lib/api";
import { toast } from "@/components/Toast";
import Link from "next/link";

export default function TalentPoolPage() {
  const [applications, setApplications] = useState<Application[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadTalentPool();
  }, []);

  const loadTalentPool = async () => {
    setLoading(true);
    try {
      const response = await crmAPI.getTalentPool();
      setApplications(response.data.applications);
    } catch (error: any) {
      console.error("Failed to load talent pool:", error);
      toast.error(error.response?.data?.error || "Failed to load talent pool");
    } finally {
      setLoading(false);
    }
  };

  const handleRemove = async (id: string, name: string) => {
    if (!window.confirm(`Remove ${name} from talent pool?`)) return;
    try {
      await crmAPI.removeFromTalentPool(id);
      toast.success("Removed from talent pool");
      loadTalentPool();
    } catch (error: any) {
      toast.error(
        error.response?.data?.error || "Failed to remove from talent pool"
      );
    }
  };

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <div>
            <h1 className="text-3xl font-bold">⭐ Talent Pool</h1>
            <p className="text-gray-600 mt-2">
              Promising candidates kept for future opportunities
            </p>
          </div>
        </div>

        {applications.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-12 text-center">
            <p className="text-gray-500 text-lg mb-4">
              Your talent pool is empty
            </p>
            <p className="text-gray-400 text-sm">
              Add candidates to the talent pool from the Applications page
            </p>
            <Link
              href="/admin/dashboard/applications"
              className="mt-4 inline-block bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700"
            >
              Go to Applications
            </Link>
          </div>
        ) : (
          <div className="bg-white rounded-lg shadow">
            <div className="p-4 border-b bg-gray-50">
              <p className="text-sm text-gray-600">
                <strong>{applications.length}</strong> candidate
                {applications.length !== 1 ? "s" : ""} in talent pool
              </p>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Candidate
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Job Applied For
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Experience
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Match Score
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Added to Pool
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {applications.map((app) => (
                    <tr key={app.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4">
                        <div>
                          <p className="font-semibold">{app.full_name}</p>
                          <p className="text-sm text-gray-500">{app.email}</p>
                          {app.phone && (
                            <p className="text-sm text-gray-500">{app.phone}</p>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <p className="text-sm">
                          {app.job?.title || "Job Deleted"}
                        </p>
                      </td>
                      <td className="px-6 py-4">
                        <p className="text-sm">
                          {app.years_of_experience} years
                        </p>
                      </td>
                      <td className="px-6 py-4">
                        {app.score > 0 ? (
                          <span
                            className={`px-2 py-1 rounded-full text-xs font-semibold ${
                              app.score >= 80
                                ? "bg-green-100 text-green-800"
                                : app.score >= 60
                                ? "bg-yellow-100 text-yellow-800"
                                : "bg-gray-100 text-gray-800"
                            }`}
                          >
                            {app.score}%
                          </span>
                        ) : (
                          <span className="text-xs text-gray-400">
                            Not analyzed
                          </span>
                        )}
                      </td>
                      <td className="px-6 py-4">
                        {app.talent_pool_added_at ? (
                          <p className="text-sm">
                            {new Date(
                              app.talent_pool_added_at
                            ).toLocaleDateString()}
                          </p>
                        ) : (
                          <span className="text-xs text-gray-400">N/A</span>
                        )}
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex gap-2">
                          <a
                            href={app.resume_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="px-3 py-1 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
                          >
                            View CV
                          </a>
                          <button
                            onClick={() => handleRemove(app.id, app.full_name)}
                            className="px-3 py-1 bg-red-600 text-white rounded text-sm hover:bg-red-700"
                          >
                            Remove
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
