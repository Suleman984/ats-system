"use client";

import { useEffect, useState } from "react";
import { jobAPI, Job } from "@/lib/api";
import Link from "next/link";
import { toast } from "@/components/Toast";

export default function JobsPage() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);
  const [filterStatus, setFilterStatus] = useState("");

  useEffect(() => {
    fetchJobs();
  }, [filterStatus]);

  const fetchJobs = async () => {
    try {
      const response = await jobAPI.getAll(
        filterStatus ? { status: filterStatus } : undefined
      );
      setJobs(response.data.jobs);
    } catch (error) {
      console.error("Failed to fetch jobs:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm("Are you sure you want to delete this job?")) return;
    try {
      await jobAPI.delete(id);
      toast.success("Job deleted successfully");
      fetchJobs();
    } catch (error) {
      toast.error("Failed to delete job");
    }
  };

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold">Jobs</h1>
          <Link
            href="/embed/dashboard/jobs/create"
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Post New Job
          </Link>
        </div>

        {/* Filter */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <label className="block text-sm font-medium mb-2">
            Filter by Status
          </label>
          <select
            className="w-full md:w-64 px-4 py-2 border rounded-lg"
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
          >
            <option value="">All Jobs</option>
            <option value="open">Open</option>
            <option value="closed">Closed</option>
            <option value="archived">Archived</option>
          </select>
        </div>

        {/* Jobs List */}
        <div className="bg-white rounded-lg shadow overflow-hidden">
          {jobs.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              <p>No jobs found. Create your first job posting!</p>
            </div>
          ) : (
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Title
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Location
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Posted Date
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Deadline
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {jobs.map((job) => (
                  <tr key={job.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div>
                        <p className="font-semibold">{job.title}</p>
                        <p className="text-sm text-gray-500">{job.job_type}</p>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm">
                      {job.location || "N/A"}
                    </td>
                    <td className="px-6 py-4 text-sm">
                      {job.created_at
                        ? new Date(job.created_at).toLocaleDateString()
                        : "N/A"}
                    </td>
                    <td className="px-6 py-4 text-sm">
                      {new Date(job.deadline).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`px-2 py-1 rounded-full text-xs ${
                          job.status === "open"
                            ? "bg-green-100 text-green-800"
                            : job.status === "closed"
                            ? "bg-red-100 text-red-800"
                            : "bg-gray-100 text-gray-800"
                        }`}
                      >
                        {job.status}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex gap-2">
                        <Link
                          href={`/embed/dashboard/jobs/${job.id}/edit`}
                          className="text-blue-600 hover:underline text-sm"
                        >
                          Edit
                        </Link>
                        <button
                          onClick={() => handleDelete(job.id)}
                          className="text-red-600 hover:underline text-sm"
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  );
}
