"use client";

import { useEffect, useState, useCallback } from "react";
import {
  applicationAPI,
  jobAPI,
  aiShortlistAPI,
  crmAPI,
  Application,
  Job,
  CandidateNote,
  RelationshipTimelineItem,
} from "@/lib/api";
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
  const [analyzingApps, setAnalyzingApps] = useState<Set<string>>(new Set());
  const [selectedAppForDetails, setSelectedAppForDetails] = useState<
    string | null
  >(null);
  const [appNotes, setAppNotes] = useState<Record<string, CandidateNote[]>>({});
  const [appTimeline, setAppTimeline] = useState<
    Record<string, RelationshipTimelineItem[]>
  >({});
  const [showNotesModal, setShowNotesModal] = useState(false);
  const [showTimelineModal, setShowTimelineModal] = useState(false);
  const [showReferralModal, setShowReferralModal] = useState(false);
  const [newNote, setNewNote] = useState("");
  const [isPrivateNote, setIsPrivateNote] = useState(false);

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

  const handleDelete = async (id: string, applicantName: string) => {
    if (
      !window.confirm(
        `Are you sure you want to delete the application from ${applicantName}? This action cannot be undone.`
      )
    )
      return;
    try {
      await applicationAPI.delete(id);
      toast.success("Application deleted successfully");
      fetchData();
    } catch (error: any) {
      toast.error(
        error.response?.data?.error || "Failed to delete application"
      );
    }
  };

  const handleBulkDelete = async (status: string) => {
    const statusLabels: { [key: string]: string } = {
      pending: "pending",
      shortlisted: "shortlisted",
      rejected: "rejected",
    };
    const label = statusLabels[status] || status;

    if (
      !window.confirm(
        `Are you sure you want to delete ALL ${label} applications? This action cannot be undone.`
      )
    )
      return;

    try {
      const response = await applicationAPI.bulkDelete(status);
      toast.success(
        response.data.message ||
          `Deleted ${response.data.deleted_count} ${label} application(s)`
      );
      fetchData();
    } catch (error: any) {
      toast.error(
        error.response?.data?.error || "Failed to delete applications"
      );
    }
  };

  const handleAnalyzeCV = async (app: Application) => {
    // Add to analyzing set
    setAnalyzingApps((prev) => new Set(prev).add(app.id));
    setOpenDropdown(null);

    try {
      // Find the job for this application to get job details
      const job = jobs.find((j) => j.id === app.job_id);
      if (!job) {
        toast.error("Job not found for this application");
        return;
      }

      // Call AI analysis endpoint with empty criteria (will use job description/requirements)
      const response = await aiShortlistAPI.analyze({
        application_id: app.id,
        required_skills: [],
        min_experience: 0,
        required_languages: [],
        match_job_description: true,
      });

      // Update the application in the local state
      setApplications((prevApps) =>
        prevApps.map((a) =>
          a.id === app.id
            ? {
                ...a,
                score: response.data.analysis.match_score,
                analysis_result: response.data.analysis,
              }
            : a
        )
      );

      toast.success(
        `CV analyzed successfully! Match Score: ${response.data.analysis.match_score}%`
      );
    } catch (error: any) {
      console.error("Failed to analyze CV:", error);
      toast.error(
        error.response?.data?.error ||
          error.response?.data?.details ||
          "Failed to analyze CV. Please try again."
      );
    } finally {
      // Remove from analyzing set
      setAnalyzingApps((prev) => {
        const newSet = new Set(prev);
        newSet.delete(app.id);
        return newSet;
      });
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

  const loadAppNotes = async (applicationId: string) => {
    try {
      const response = await crmAPI.getNotes(applicationId);
      setAppNotes((prev) => ({
        ...prev,
        [applicationId]: response.data.notes,
      }));
    } catch (error) {
      console.error("Failed to load notes:", error);
    }
  };

  const loadAppTimeline = async (applicationId: string) => {
    try {
      const response = await crmAPI.getTimeline(applicationId);
      setAppTimeline((prev) => ({
        ...prev,
        [applicationId]: response.data.timeline,
      }));
    } catch (error) {
      console.error("Failed to load timeline:", error);
    }
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

          {/* Bulk Delete Actions */}
          <div className="mt-4 pt-4 border-t">
            <label className="block text-sm font-medium mb-2 text-gray-700">
              Bulk Delete by Status
            </label>
            <div className="flex gap-2 flex-wrap">
              <button
                onClick={() => handleBulkDelete("pending")}
                className="px-3 py-1.5 text-xs bg-yellow-100 text-yellow-800 rounded hover:bg-yellow-200 transition-colors"
              >
                Delete All Pending
              </button>
              <button
                onClick={() => handleBulkDelete("shortlisted")}
                className="px-3 py-1.5 text-xs bg-green-100 text-green-800 rounded hover:bg-green-200 transition-colors"
              >
                Delete All Shortlisted
              </button>
              <button
                onClick={() => handleBulkDelete("rejected")}
                className="px-3 py-1.5 text-xs bg-red-100 text-red-800 rounded hover:bg-red-200 transition-colors"
              >
                Delete All Rejected
              </button>
            </div>
            <p className="text-xs text-gray-500 mt-2">
              ⚠️ Warning: Bulk delete actions cannot be undone
            </p>
          </div>
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
                        <p className="text-sm">
                          {app.job?.title || "Job Deleted"}
                          {!app.job && app.job_id && (
                            <span className="text-xs text-gray-400 ml-1">
                              (Job ID: {app.job_id})
                            </span>
                          )}
                        </p>
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
                                  // Show in a modal or use a better UI component
                                  // For now, using a more user-friendly approach
                                  toast.info(analysisText.split("\n\n")[0]); // Show first part as toast
                                  // Could also open a modal here for full details
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
                                <div className="border-t border-gray-200 my-1"></div>
                                <button
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    handleAnalyzeCV(app);
                                  }}
                                  disabled={analyzingApps.has(app.id)}
                                  className={`block w-full text-left px-3 py-1.5 text-xs transition-colors ${
                                    analyzingApps.has(app.id)
                                      ? "text-gray-400 cursor-not-allowed"
                                      : "text-blue-600 hover:bg-gray-50"
                                  }`}
                                >
                                  {analyzingApps.has(app.id) ? (
                                    <span className="flex items-center gap-1">
                                      <span className="inline-block w-3 h-3 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></span>
                                      Analyzing...
                                    </span>
                                  ) : (
                                    "🔍 Analyze CV"
                                  )}
                                </button>
                                {(app.status === "pending" ||
                                  app.status === "cv_viewed") && (
                                  <>
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
                                <div className="border-t border-gray-200 my-1"></div>
                                <button
                                  onClick={async (e) => {
                                    e.stopPropagation();
                                    setSelectedAppForDetails(app.id);
                                    await loadAppNotes(app.id);
                                    setShowNotesModal(true);
                                    setOpenDropdown(null);
                                  }}
                                  className="block w-full text-left px-3 py-1.5 text-xs text-blue-600 hover:bg-gray-50 transition-colors"
                                >
                                  📝 Notes
                                </button>
                                <button
                                  onClick={async (e) => {
                                    e.stopPropagation();
                                    setSelectedAppForDetails(app.id);
                                    await loadAppTimeline(app.id);
                                    setShowTimelineModal(true);
                                    setOpenDropdown(null);
                                  }}
                                  className="block w-full text-left px-3 py-1.5 text-xs text-blue-600 hover:bg-gray-50 transition-colors"
                                >
                                  📅 Timeline
                                </button>
                                <button
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    setSelectedAppForDetails(app.id);
                                    setShowReferralModal(true);
                                    setOpenDropdown(null);
                                  }}
                                  className="block w-full text-left px-3 py-1.5 text-xs text-blue-600 hover:bg-gray-50 transition-colors"
                                >
                                  👥 Referral Info
                                </button>
                                {app.in_talent_pool ? (
                                  <button
                                    onClick={async (e) => {
                                      e.stopPropagation();
                                      try {
                                        await crmAPI.removeFromTalentPool(
                                          app.id
                                        );
                                        toast.success(
                                          "Removed from talent pool"
                                        );
                                        fetchData();
                                      } catch (error) {
                                        toast.error(
                                          "Failed to remove from talent pool"
                                        );
                                      }
                                      setOpenDropdown(null);
                                    }}
                                    className="block w-full text-left px-3 py-1.5 text-xs text-yellow-600 hover:bg-gray-50 transition-colors"
                                  >
                                    ⭐ Remove from Talent Pool
                                  </button>
                                ) : (
                                  <button
                                    onClick={async (e) => {
                                      e.stopPropagation();
                                      try {
                                        await crmAPI.addToTalentPool(app.id);
                                        toast.success("Added to talent pool");
                                        fetchData();
                                      } catch (error) {
                                        toast.error(
                                          "Failed to add to talent pool"
                                        );
                                      }
                                      setOpenDropdown(null);
                                    }}
                                    className="block w-full text-left px-3 py-1.5 text-xs text-yellow-600 hover:bg-gray-50 transition-colors"
                                  >
                                    ⭐ Add to Talent Pool
                                  </button>
                                )}
                                <div className="border-t border-gray-200 my-1"></div>
                                <button
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    handleDelete(app.id, app.full_name);
                                    setOpenDropdown(null);
                                  }}
                                  className="block w-full text-left px-3 py-1.5 text-xs text-red-600 hover:bg-red-50 transition-colors font-medium"
                                >
                                  🗑️ Delete Application
                                </button>
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

      {/* Notes Modal */}
      {showNotesModal && selectedAppForDetails && (
        <NotesModal
          applicationId={selectedAppForDetails}
          notes={appNotes[selectedAppForDetails] || []}
          onClose={() => {
            setShowNotesModal(false);
            setSelectedAppForDetails(null);
            setNewNote("");
            setIsPrivateNote(false);
          }}
          onRefresh={async () => {
            if (selectedAppForDetails) {
              await loadAppNotes(selectedAppForDetails);
            }
          }}
        />
      )}

      {/* Timeline Modal */}
      {showTimelineModal && selectedAppForDetails && (
        <TimelineModal
          applicationId={selectedAppForDetails}
          timeline={appTimeline[selectedAppForDetails] || []}
          onClose={() => {
            setShowTimelineModal(false);
            setSelectedAppForDetails(null);
          }}
        />
      )}

      {/* Referral Modal */}
      {showReferralModal && selectedAppForDetails && (
        <ReferralModal
          application={applications.find((a) => a.id === selectedAppForDetails)}
          onClose={() => {
            setShowReferralModal(false);
            setSelectedAppForDetails(null);
          }}
          onSave={async () => {
            await fetchData();
            setShowReferralModal(false);
            setSelectedAppForDetails(null);
          }}
        />
      )}
    </div>
  );
}

// Notes Modal Component
function NotesModal({
  applicationId,
  notes,
  onClose,
  onRefresh,
}: {
  applicationId: string;
  notes: CandidateNote[];
  onClose: () => void;
  onRefresh: () => void;
}) {
  const [newNote, setNewNote] = useState("");
  const [isPrivate, setIsPrivate] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleAddNote = async () => {
    if (!newNote.trim()) {
      toast.error("Please enter a note");
      return;
    }
    setLoading(true);
    try {
      await crmAPI.addNote(applicationId, newNote.trim(), isPrivate);
      toast.success("Note added successfully");
      setNewNote("");
      setIsPrivate(false);
      onRefresh();
    } catch (error: any) {
      toast.error(error.response?.data?.error || "Failed to add note");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        <div className="p-6 border-b">
          <div className="flex justify-between items-center">
            <h2 className="text-2xl font-bold">Candidate Notes</h2>
            <button
              onClick={onClose}
              className="text-gray-500 hover:text-gray-700 text-2xl"
            >
              ×
            </button>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto p-6">
          <div className="space-y-4 mb-6">
            {notes.length === 0 ? (
              <p className="text-gray-500 text-center py-8">No notes yet</p>
            ) : (
              notes.map((note) => (
                <div key={note.id} className="border rounded-lg p-4 bg-gray-50">
                  <div className="flex justify-between items-start mb-2">
                    <div>
                      <p className="font-semibold text-sm">
                        {note.admin?.name || "Unknown"}
                      </p>
                      <p className="text-xs text-gray-500">
                        {new Date(note.created_at).toLocaleString()}
                        {note.is_private && (
                          <span className="ml-2 text-blue-600">🔒 Private</span>
                        )}
                      </p>
                    </div>
                  </div>
                  <p className="text-gray-800 whitespace-pre-wrap">
                    {note.note}
                  </p>
                </div>
              ))
            )}
          </div>

          <div className="border-t pt-4">
            <h3 className="font-semibold mb-3">Add New Note</h3>
            <textarea
              className="w-full px-4 py-2 border rounded-lg resize-none mb-3"
              placeholder="Enter your note here..."
              rows={4}
              value={newNote}
              onChange={(e) => setNewNote(e.target.value)}
            />
            <div className="flex items-center justify-between">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={isPrivate}
                  onChange={(e) => setIsPrivate(e.target.checked)}
                  className="mr-2"
                />
                <span className="text-sm text-gray-600">
                  Private note (only visible to you)
                </span>
              </label>
              <button
                onClick={handleAddNote}
                disabled={loading || !newNote.trim()}
                className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-300"
              >
                {loading ? "Adding..." : "Add Note"}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

// Timeline Modal Component
function TimelineModal({
  applicationId,
  timeline,
  onClose,
}: {
  applicationId: string;
  timeline: RelationshipTimelineItem[];
  onClose: () => void;
}) {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-3xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        <div className="p-6 border-b">
          <div className="flex justify-between items-center">
            <h2 className="text-2xl font-bold">Relationship Timeline</h2>
            <button
              onClick={onClose}
              className="text-gray-500 hover:text-gray-700 text-2xl"
            >
              ×
            </button>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto p-6">
          {timeline.length === 0 ? (
            <p className="text-gray-500 text-center py-8">
              No timeline events yet
            </p>
          ) : (
            <div className="space-y-4">
              {timeline.map((item, index) => (
                <div key={index} className="flex items-start gap-4">
                  <div className="text-2xl flex-shrink-0">{item.icon}</div>
                  <div className="flex-1 border-l-2 border-gray-200 pl-4 pb-4">
                    <div className="flex justify-between items-start">
                      <div>
                        <p className="font-semibold">{item.title}</p>
                        <p className="text-sm text-gray-600 mt-1">
                          {item.description}
                        </p>
                        {item.author && (
                          <p className="text-xs text-gray-500 mt-1">
                            By: {item.author}
                          </p>
                        )}
                        {item.admin && (
                          <p className="text-xs text-gray-500 mt-1">
                            By: {item.admin}
                          </p>
                        )}
                      </div>
                      <p className="text-xs text-gray-500">
                        {new Date(item.timestamp).toLocaleString()}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// Referral Modal Component
function ReferralModal({
  application,
  onClose,
  onSave,
}: {
  application: Application | undefined;
  onClose: () => void;
  onSave: () => void;
}) {
  const [referralSource, setReferralSource] = useState(
    application?.referral_source || ""
  );
  const [referredByName, setReferredByName] = useState(
    application?.referred_by_name || ""
  );
  const [referredByEmail, setReferredByEmail] = useState(
    application?.referred_by_email || ""
  );
  const [referredByPhone, setReferredByPhone] = useState(
    application?.referred_by_phone || ""
  );
  const [loading, setLoading] = useState(false);

  const handleSave = async () => {
    if (!application) return;
    setLoading(true);
    try {
      await crmAPI.updateReferral(
        application.id,
        referralSource,
        referredByName,
        referredByEmail,
        referredByPhone
      );
      toast.success("Referral information updated");
      onSave();
    } catch (error: any) {
      toast.error(
        error.response?.data?.error || "Failed to update referral info"
      );
    } finally {
      setLoading(false);
    }
  };

  if (!application) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
        <div className="p-6 border-b">
          <div className="flex justify-between items-center">
            <h2 className="text-2xl font-bold">Referral Information</h2>
            <button
              onClick={onClose}
              className="text-gray-500 hover:text-gray-700 text-2xl"
            >
              ×
            </button>
          </div>
        </div>
        <div className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">
              Referral Source
            </label>
            <input
              type="text"
              className="w-full px-4 py-2 border rounded-lg"
              placeholder="e.g., LinkedIn, Employee Referral, Job Board"
              value={referralSource}
              onChange={(e) => setReferralSource(e.target.value)}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">
              Referred By (Name)
            </label>
            <input
              type="text"
              className="w-full px-4 py-2 border rounded-lg"
              placeholder="Name of referrer"
              value={referredByName}
              onChange={(e) => setReferredByName(e.target.value)}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">
              Referred By (Email)
            </label>
            <input
              type="email"
              className="w-full px-4 py-2 border rounded-lg"
              placeholder="email@example.com"
              value={referredByEmail}
              onChange={(e) => setReferredByEmail(e.target.value)}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">
              Referred By (Phone)
            </label>
            <input
              type="tel"
              className="w-full px-4 py-2 border rounded-lg"
              placeholder="+1234567890"
              value={referredByPhone}
              onChange={(e) => setReferredByPhone(e.target.value)}
            />
          </div>
          <div className="flex gap-3 pt-4">
            <button
              onClick={onClose}
              className="flex-1 px-4 py-2 border rounded-lg hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              disabled={loading}
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-blue-300"
            >
              {loading ? "Saving..." : "Save"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
