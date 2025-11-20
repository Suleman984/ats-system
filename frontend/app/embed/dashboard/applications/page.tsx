"use client";

import { useEffect, useState, useCallback } from "react";
import { applicationAPI, jobAPI, Application, Job } from "@/lib/api";
import { toast } from "@/components/Toast";

export default function ApplicationsPage() {
  const [applications, setApplications] = useState<Application[]>([]);
  const [jobs, setJobs] = useState<Job[]>([]);
  const [selectedJob, setSelectedJob] = useState("");
  const [selectedStatus, setSelectedStatus] = useState("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [loading, setLoading] = useState(true);
  const [isFiltering, setIsFiltering] = useState(false);
  const [openDropdown, setOpenDropdown] = useState<string | null>(null);

  const fetchData = useCallback(
    async (skipFilters = false) => {
      try {
        const params: {
          job_id?: string;
          status?: string;
          date_from?: string;
          date_to?: string;
        } = {};

        if (!skipFilters) {
          if (selectedJob) params.job_id = selectedJob;
          if (selectedStatus) params.status = selectedStatus;
          if (dateFrom) params.date_from = dateFrom;
          if (dateTo) params.date_to = dateTo;
        }

        console.log("Fetching applications with params:", params);

        const [appsRes, jobsRes] = await Promise.all([
          applicationAPI.getAll(params),
          jobAPI.getAll(),
        ]);
        setApplications(appsRes.data.applications);
        setJobs(jobsRes.data.jobs);
        console.log("Fetched applications:", appsRes.data.applications.length);
      } catch (error) {
        console.error("Failed to fetch data:", error);
      } finally {
        setLoading(false);
        setIsFiltering(false);
      }
    },
    [selectedJob, selectedStatus, dateFrom, dateTo]
  );

  // Initial load
  useEffect(() => {
    fetchData(true);
  }, []);

  // Filter changes (debounced to avoid too many requests)
  useEffect(() => {
    // Skip if still loading initial data
    if (loading) return;

    setIsFiltering(true);
    const timeoutId = setTimeout(() => {
      fetchData(false);
    }, 300); // Debounce for 300ms

    return () => clearTimeout(timeoutId);
  }, [selectedJob, selectedStatus, dateFrom, dateTo, loading, fetchData]);

  const handleShortlist = async (id: string) => {
    if (!window.confirm("Shortlist this candidate?")) return;
    try {
      await applicationAPI.shortlist(id);
      toast.success("Candidate shortlisted! Email sent.");
      fetchData();
    } catch (error) {
      toast.error("Failed to shortlist candidate");
    }
  };

  const handleReject = async (id: string) => {
    if (!window.confirm("Reject this candidate?")) return;
    try {
      await applicationAPI.reject(id);
      toast.success("Candidate rejected. Email sent.");
      fetchData();
    } catch (error) {
      toast.error("Failed to reject candidate");
    }
  };

  const clearDateFilters = () => {
    setDateFrom("");
    setDateTo("");
  };

  const getScoreColor = (score: number) => {
    if (score >= 80) return "text-green-600 bg-green-100";
    if (score >= 60) return "text-yellow-600 bg-yellow-100";
    return "text-red-600 bg-red-100";
  };

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold">Applications</h1>
          {isFiltering && (
            <div className="text-sm text-gray-500">Filtering...</div>
          )}
        </div>

        {/* Filters */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
            <div>
              <label className="block text-sm font-medium mb-2">
                Filter by Job
              </label>
              <select
                className="w-full px-4 py-2 border rounded-lg"
                value={selectedJob}
                onChange={(e) => setSelectedJob(e.target.value)}
              >
                <option value="">All Jobs</option>
                {jobs.map((job) => (
                  <option key={job.id} value={job.id}>
                    {job.title}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">
                Filter by Status
              </label>
              <select
                className="w-full px-4 py-2 border rounded-lg"
                value={selectedStatus}
                onChange={(e) => setSelectedStatus(e.target.value)}
              >
                <option value="">All Status</option>
                <option value="pending">Pending</option>
                <option value="shortlisted">Shortlisted</option>
                <option value="rejected">Rejected</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">
                Date From
              </label>
              <input
                type="date"
                className="w-full px-4 py-2 border rounded-lg"
                value={dateFrom}
                onChange={(e) => setDateFrom(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">Date To</label>
              <input
                type="date"
                className="w-full px-4 py-2 border rounded-lg"
                value={dateTo}
                onChange={(e) => setDateTo(e.target.value)}
              />
            </div>
          </div>
          {(dateFrom || dateTo) && (
            <div className="flex justify-end">
              <button
                onClick={clearDateFilters}
                className="text-sm text-blue-600 hover:underline"
              >
                Clear Date Filters
              </button>
            </div>
          )}
        </div>

        {/* Applications List */}
        <div className="bg-white rounded-lg shadow overflow-visible">
          {applications.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              No applications found
              {(selectedJob || selectedStatus || dateFrom || dateTo) && (
                <p className="text-sm mt-2">Try adjusting your filters</p>
              )}
            </div>
          ) : (
            <div className="p-4 border-b bg-gray-50">
              <p className="text-sm text-gray-600">
                Showing <strong>{applications.length}</strong> application
                {applications.length !== 1 ? "s" : ""}
                {selectedJob && ` for selected job`}
                {(dateFrom || dateTo) && ` within date range`}
              </p>
            </div>
          )}
          {applications.length > 0 && (
            <div className="overflow-visible">
              <table className="w-full relative">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Candidate
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Job
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Experience
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Match Score
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Applied
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
                  {applications.map((app) => (
                    <tr key={app.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4">
                        <div>
                          <p className="font-semibold">{app.full_name}</p>
                          <p className="text-sm text-gray-500">{app.email}</p>
                          <p className="text-sm text-gray-500">{app.phone}</p>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <p className="text-sm">{app.job?.title || "N/A"}</p>
                      </td>
                      <td className="px-6 py-4">
                        <p className="text-sm">
                          {app.years_of_experience} years
                        </p>
                      </td>
                      <td className="px-6 py-4">
                        {app.score > 0 ? (
                          <div className="flex items-center gap-2">
                            <span
                              className={`px-2 py-1 rounded-full text-xs font-semibold ${getScoreColor(
                                app.score
                              )}`}
                              title={
                                typeof app.analysis_result === "object" &&
                                app.analysis_result
                                  ? `Match Reason: ${app.analysis_result.match_reason}`
                                  : "AI Match Score"
                              }
                            >
                              {app.score}%
                            </span>
                            {app.analysis_result && (
                              <button
                                onClick={() => {
                                  const analysis =
                                    typeof app.analysis_result === "string"
                                      ? JSON.parse(app.analysis_result)
                                      : app.analysis_result;
                                  const analysisText = `AI Analysis:\n\nMatch Score: ${
                                    analysis.match_score
                                  }%\n\nSkills: ${
                                    analysis.skills?.join(", ") || "N/A"
                                  }\nExperience: ${
                                    analysis.experience
                                  } years\nLanguages: ${
                                    analysis.languages?.join(", ") || "N/A"
                                  }\n\nStrengths:\n${
                                    analysis.strengths?.join("\n") || "N/A"
                                  }\n\nMissing Skills:\n${
                                    analysis.missing_skills?.join("\n") ||
                                    "None"
                                  }\n\nReason: ${
                                    analysis.match_reason || "N/A"
                                  }`;
                                  toast.info(analysisText.split("\n\n")[0]);
                                }}
                                className="text-xs text-blue-600 cursor-pointer hover:underline"
                                title="View AI Analysis Details"
                              >
                                📊
                              </button>
                            )}
                          </div>
                        ) : (
                          <span className="text-xs text-gray-400">
                            Not analyzed
                          </span>
                        )}
                      </td>
                      <td className="px-6 py-4">
                        <p className="text-sm">
                          {new Date(app.applied_at).toLocaleDateString()}
                        </p>
                      </td>
                      <td className="px-6 py-4">
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
                      </td>
                      <td className="px-6 py-4 relative">
                        <div className="relative inline-block">
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              setOpenDropdown(
                                openDropdown === app.id ? null : app.id
                              );
                            }}
                            className="text-gray-500 hover:text-gray-700 text-lg font-bold focus:outline-none px-2 py-1"
                          >
                            ⋮
                          </button>
                          {openDropdown === app.id && (
                            <>
                              {/* Backdrop to close on outside click */}
                              <div
                                className="fixed inset-0 z-[100]"
                                onClick={() => setOpenDropdown(null)}
                              ></div>
                              {/* Dropdown menu */}
                              <div className="absolute right-0 top-full mt-1 w-36 bg-white rounded-md shadow-2xl z-[200] border border-gray-300 py-1.5">
                                <a
                                  href={app.resume_url}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  onClick={async () => {
                                    setOpenDropdown(null);
                                    // Track CV view
                                    try {
                                      await applicationAPI.trackCVView(app.id);
                                      // Refresh applications to update status
                                      fetchData();
                                    } catch (error) {
                                      console.error(
                                        "Failed to track CV view:",
                                        error
                                      );
                                    }
                                  }}
                                  className="block px-3 py-1.5 text-xs text-gray-700 hover:bg-gray-50 transition-colors"
                                >
                                  📄 View CV
                                </a>
                                {app.portfolio_url && (
                                  <a
                                    href={app.portfolio_url}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    onClick={() => setOpenDropdown(null)}
                                    className="block px-3 py-1.5 text-xs text-gray-700 hover:bg-gray-50 transition-colors"
                                  >
                                    🎨 Portfolio
                                  </a>
                                )}
                                {app.linkedin_url && (
                                  <a
                                    href={app.linkedin_url}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    onClick={() => setOpenDropdown(null)}
                                    className="block px-3 py-1.5 text-xs text-gray-700 hover:bg-gray-50 transition-colors"
                                  >
                                    💼 LinkedIn
                                  </a>
                                )}
                                {app.status === "pending" && (
                                  <>
                                    <div className="border-t border-gray-200 my-1"></div>
                                    <button
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        handleShortlist(app.id);
                                        setOpenDropdown(null);
                                      }}
                                      className="block w-full text-left px-3 py-1.5 text-xs text-green-600 hover:bg-gray-50 transition-colors"
                                    >
                                      ✓ Shortlist
                                    </button>
                                    <button
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        handleReject(app.id);
                                        setOpenDropdown(null);
                                      }}
                                      className="block w-full text-left px-3 py-1.5 text-xs text-red-600 hover:bg-gray-50 transition-colors"
                                    >
                                      ✗ Reject
                                    </button>
                                  </>
                                )}
                              </div>
                            </>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
