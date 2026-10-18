"use client";

import { Suspense, useCallback, useEffect, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import {
  applicationAPI,
  jobAPI,
  aiShortlistAPI,
  crmAPI,
  errorMessage,
  type Application,
  type Job,
  type CandidateNote,
  type RelationshipTimelineItem,
} from "@/lib/api";
import { toast } from "@/components/Toast";
import NotesModal from "@/components/applications/NotesModal";
import TimelineModal from "@/components/applications/TimelineModal";
import ReferralModal from "@/components/applications/ReferralModal";
import MessagesModal from "@/components/applications/MessagesModal";
import AnalysisModal, { parseAnalysis } from "@/components/applications/AnalysisModal";
import { APPLICATION_STATUSES, NEXT_STAGE_ACTION, isUndecided, statusBadge, statusLabel } from "@/lib/status";
import { useDashboardHref, useIsOwner } from "@/lib/dashboard-path";
import Link from "next/link";
import FileLink from "@/components/applications/FileLink";
import { notifyUnreadChanged, useUnreadSummary } from "@/lib/unread";
import { useChatEvents } from "@/lib/realtime";

const PAGE_SIZE = 50;

function Applications() {
  const href = useDashboardHref();
  // The nav bell links here with ?unread=1.
  const unreadParam = useSearchParams().get("unread") === "1";
  const [unreadOnly, setUnreadOnly] = useState(unreadParam);
  const [seenUnreadParam, setSeenUnreadParam] = useState(unreadParam);
  if (seenUnreadParam !== unreadParam) {
    setSeenUnreadParam(unreadParam);
    setUnreadOnly(unreadParam);
  }
  const unreadSummary = useUnreadSummary();
  const isOwner = useIsOwner();
  const [applications, setApplications] = useState<Application[]>([]);
  const [total, setTotal] = useState(0);
  const [loadingMore, setLoadingMore] = useState(false);
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
  const [showMessagesModal, setShowMessagesModal] = useState(false);
  const [analysisApp, setAnalysisApp] = useState<Application | null>(null);

  // Live unread badges: a candidate message bumps its row (unless that thread is open right now),
  // and a recruiter reading the thread clears it.
  useChatEvents((event) => {
    if (event.type === "reconnected") return;
    const open = showMessagesModal && selectedAppForDetails === event.application_id;
    setApplications((prev) =>
      prev.map((a) => {
        if (a.id !== event.application_id) return a;
        if (event.type === "read" && event.reader === "recruiter") return { ...a, unread_messages: 0 };
        if (event.type === "message" && event.message.sender_type === "candidate" && !open) {
          return { ...a, unread_messages: (a.unread_messages ?? 0) + 1 };
        }
        return a;
      })
    );
  });
  const [batchScoring, setBatchScoring] = useState(false);

  const fetchData = useCallback(async () => {
    const params = Object.fromEntries(
      Object.entries({
        job_id: selectedJob,
        status: selectedStatus,
        date_from: dateFrom,
        date_to: dateTo,
        unread: unreadOnly,
      }).filter(([, value]) => value)
    );
    try {
      const { data } = await applicationAPI.getAll({ ...params, limit: PAGE_SIZE });
      setApplications(data.applications);
      setTotal(data.total);
    } catch (error) {
      toast.error(errorMessage(error, "Failed to fetch applications"));
    } finally {
      setLoading(false);
      setIsFiltering(false);
    }
  }, [selectedJob, selectedStatus, dateFrom, dateTo, unreadOnly]);

  useEffect(() => {
    jobAPI
      .getAll()
      .then(({ data }) => setJobs(data.jobs))
      .catch((error) => toast.error(errorMessage(error, "Failed to fetch jobs")));
  }, []);

  // Filter changes refetch after a 300ms debounce; the first load runs immediately.
  const isFirstLoad = useRef(true);
  useEffect(() => {
    const delay = isFirstLoad.current ? 0 : 300;
    if (!isFirstLoad.current) setIsFiltering(true);
    isFirstLoad.current = false;
    const timeoutId = setTimeout(fetchData, delay);
    return () => clearTimeout(timeoutId);
  }, [fetchData]);

  const loadMore = async () => {
    setLoadingMore(true);
    try {
      const { data } = await applicationAPI.getAll({
        ...Object.fromEntries(
          Object.entries({
            job_id: selectedJob,
            status: selectedStatus,
            date_from: dateFrom,
            date_to: dateTo,
            unread: unreadOnly,
          }).filter(([, value]) => value)
        ),
        limit: PAGE_SIZE,
        offset: applications.length,
      });
      // Skip rows already shown in case new applications shifted the offset.
      setApplications((prev) => {
        const seen = new Set(prev.map((a) => a.id));
        return [...prev, ...data.applications.filter((a) => !seen.has(a.id))];
      });
      setTotal(data.total);
    } catch (error) {
      toast.error(errorMessage(error, "Failed to load more applications"));
    } finally {
      setLoadingMore(false);
    }
  };

  const handleShortlist = async (id: string) => {
    if (!window.confirm("Shortlist this candidate?")) return;
    try {
      await applicationAPI.shortlist(id);
      toast.success("Candidate shortlisted. They'll be notified by email.");
      fetchData();
    } catch (error) {
      toast.error(errorMessage(error, "Failed to shortlist candidate"));
    }
  };

  const handleReject = async (id: string) => {
    if (!window.confirm("Reject this candidate?")) return;
    try {
      await applicationAPI.reject(id);
      toast.success("Candidate rejected. They'll be notified by email.");
      fetchData();
    } catch (error) {
      toast.error(errorMessage(error, "Failed to reject candidate"));
    }
  };

  const handleAdvance = async (app: Application) => {
    const action = NEXT_STAGE_ACTION[app.status];
    if (!window.confirm(`${action} for ${app.full_name}? They'll be notified by email.`)) return;
    try {
      await applicationAPI.advance(app.id);
      toast.success(`${app.full_name}: ${action.toLowerCase()} done`);
      fetchData();
    } catch (error) {
      toast.error(errorMessage(error, "Failed to update application"));
    }
  };

  const handleReopen = async (id: string) => {
    if (!window.confirm("Move this candidate back to review? They won't be notified.")) return;
    try {
      await applicationAPI.reopen(id);
      toast.success("Application moved back to review");
      fetchData();
    } catch (error) {
      toast.error(errorMessage(error, "Failed to reopen application"));
    }
  };

  const handleBatchScore = async () => {
    const job = jobs.find((j) => j.id === selectedJob);
    if (!job) return;
    setBatchScoring(true);
    try {
      const { data } = await aiShortlistAPI.analyzeBatch({ job_id: job.id });
      const failed = data.results.filter((r) => r.error).length;
      toast.success(
        `Scored ${data.total_analyzed - failed} of ${data.total_analyzed} undecided application(s) for ${job.title}` +
          (failed ? ` — ${failed} CV(s) couldn't be read` : "")
      );
      fetchData();
    } catch (error) {
      toast.error(errorMessage(error, "Failed to score applications"));
    } finally {
      setBatchScoring(false);
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
    } catch (error) {
      toast.error(
        errorMessage(error, "Failed to delete application")
      );
    }
  };

  const handleBulkDelete = async (status: string) => {
    const label = statusLabel(status).toLowerCase();
    const jobTitle = jobs.find((j) => j.id === selectedJob)?.title;
    const scope = jobTitle ? ` for "${jobTitle}"` : " across all jobs";

    if (
      !window.confirm(
        `Are you sure you want to delete ALL ${label} applications${scope}? This action cannot be undone.`
      )
    )
      return;

    try {
      const response = await applicationAPI.bulkDelete(status, selectedJob);
      toast.success(
        response.data.message ||
          `Deleted ${response.data.deleted_count} ${label} application(s)`
      );
      fetchData();
    } catch (error) {
      toast.error(
        errorMessage(error, "Failed to delete applications")
      );
    }
  };

  const handleAnalyzeCV = async (app: Application) => {
    setAnalyzingApps((prev) => new Set(prev).add(app.id));
    setOpenDropdown(null);

    try {
      const job = jobs.find((j) => j.id === app.job_id);
      if (!job) {
        toast.error("Job not found for this application");
        return;
      }

      // Empty criteria makes the API fall back to the job's saved criteria.
      const response = await aiShortlistAPI.analyze({
        application_id: app.id,
        required_skills: [],
        min_experience: 0,
        required_languages: [],
        match_job_description: true,
      });

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
    } catch (error) {
      toast.error(errorMessage(error, "Failed to analyze CV. Please try again."));
    } finally {
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
      toast.error(errorMessage(error, "Failed to load notes"));
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
      toast.error(errorMessage(error, "Failed to load timeline"));
    }
  };

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold">Applications</h1>
          <div className="flex items-center gap-4">
            {isFiltering && (
              <div className="text-sm text-gray-500">Filtering...</div>
            )}
            <Link
              href={href("/admin/dashboard/applications/manual-add")}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm"
            >
              + Add candidate
            </Link>
          </div>
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
                    {unreadSummary.byJob[job.id] ? ` (${unreadSummary.byJob[job.id]} new)` : ""}
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
                {APPLICATION_STATUSES.map((s) => (
                  <option key={s.value} value={s.value}>
                    {s.label}
                  </option>
                ))}
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
          <div className="flex justify-between items-center gap-4">
            {selectedJob ? (
              <button
                onClick={handleBatchScore}
                disabled={batchScoring}
                className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-blue-300"
              >
                {batchScoring ? "Scoring CVs..." : "🔍 Score all undecided CVs for this job"}
              </button>
            ) : (
              <span className="text-xs text-gray-500">Select a job to score all its CVs at once.</span>
            )}
            <label className="flex items-center gap-2 text-sm text-gray-700 ml-auto">
              <input
                type="checkbox"
                checked={unreadOnly}
                onChange={(e) => setUnreadOnly(e.target.checked)}
              />
              Unread messages only
              {unreadSummary.total > 0 && (
                <span className="px-1.5 rounded-full bg-red-600 text-white text-xs">{unreadSummary.total}</span>
              )}
            </label>
            {(dateFrom || dateTo) && (
              <button
                onClick={clearDateFilters}
                className="text-sm text-blue-600 hover:underline"
              >
                Clear Date Filters
              </button>
            )}
          </div>

          {/* Bulk Delete Actions (owners only; the API enforces this too) */}
          {isOwner && (
          <div className="mt-4 pt-4 border-t">
            <label className="block text-sm font-medium mb-2 text-gray-700">
              Bulk Delete by Status
              {selectedJob ? " (selected job only)" : " (all jobs)"}
            </label>
            <div className="flex gap-2 flex-wrap">
              {APPLICATION_STATUSES.map((s) => (
                <button
                  key={s.value}
                  onClick={() => handleBulkDelete(s.value)}
                  className={`px-3 py-1.5 text-xs rounded hover:opacity-80 transition-opacity ${s.badge}`}
                >
                  Delete All {s.label}
                </button>
              ))}
            </div>
            <p className="text-xs text-gray-500 mt-2">
              ⚠️ Warning: Bulk delete actions cannot be undone
            </p>
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
                          <p className="font-semibold flex items-center gap-2">
                            {app.full_name}
                            {!!app.unread_messages && (
                              <button
                                onClick={() => {
                                  setSelectedAppForDetails(app.id);
                                  setShowMessagesModal(true);
                                }}
                                title="Unread messages from this candidate"
                                className="px-1.5 py-0.5 rounded-full bg-red-600 text-white text-xs font-medium hover:bg-red-700"
                              >
                                💬 {app.unread_messages}
                              </button>
                            )}
                          </p>
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
                                onClick={() => setAnalysisApp(app)}
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
                          className={`px-2 py-1 rounded-full text-xs ${statusBadge(app.status)}`}
                        >
                          {statusLabel(app.status)}
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
                                className="fixed inset-0 z-100"
                                onClick={() => setOpenDropdown(null)}
                              ></div>
                              {/* Dropdown menu */}
                              <div className="absolute right-0 top-full mt-1 w-36 bg-white rounded-md shadow-2xl z-200 border border-gray-300 py-1.5">
                                <FileLink
                                  applicationId={app.id}
                                  kind="resume"
                                  url={app.resume_url}
                                  hasFile={app.has_resume_file}
                                  onOpen={async () => {
                                    setOpenDropdown(null);
                                    try {
                                      await applicationAPI.trackCVView(app.id);
                                      fetchData();
                                    } catch (error) {
                                      toast.error(errorMessage(error, "Failed to track CV view"));
                                    }
                                  }}
                                  className="block w-full px-3 py-1.5 text-xs text-gray-700 hover:bg-gray-50 transition-colors"
                                >
                                  📄 View CV
                                </FileLink>
                                <FileLink
                                  applicationId={app.id}
                                  kind="portfolio"
                                  url={app.portfolio_url}
                                  hasFile={app.has_portfolio_file}
                                  onOpen={() => setOpenDropdown(null)}
                                  className="block w-full px-3 py-1.5 text-xs text-gray-700 hover:bg-gray-50 transition-colors"
                                >
                                  🎨 Portfolio
                                </FileLink>
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
                                {isUndecided(app.status) && (
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
                                )}
                                {NEXT_STAGE_ACTION[app.status] && (
                                  <button
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      handleAdvance(app);
                                      setOpenDropdown(null);
                                    }}
                                    className="block w-full text-left px-3 py-1.5 text-xs text-green-600 hover:bg-gray-50 transition-colors"
                                  >
                                    → {NEXT_STAGE_ACTION[app.status]}
                                  </button>
                                )}
                                {app.status !== "rejected" && app.status !== "hired" && (
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
                                )}
                                {!isUndecided(app.status) && (
                                  <button
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      handleReopen(app.id);
                                      setOpenDropdown(null);
                                    }}
                                    className="block w-full text-left px-3 py-1.5 text-xs text-gray-700 hover:bg-gray-50 transition-colors"
                                  >
                                    ↺ Move back to review
                                  </button>
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
                                    setShowMessagesModal(true);
                                    setOpenDropdown(null);
                                  }}
                                  className="block w-full text-left px-3 py-1.5 text-xs text-blue-600 hover:bg-gray-50 transition-colors"
                                >
                                  💬 Messages
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
                                      } catch {
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
                                      } catch {
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
        {applications.length > 0 && (
          <div className="flex items-center justify-between mt-4 text-sm text-gray-600">
            <span>
              Showing {applications.length} of {total}
            </span>
            {applications.length < total && (
              <button
                onClick={loadMore}
                disabled={loadingMore}
                className="px-4 py-2 bg-white border rounded-lg hover:bg-gray-50 disabled:opacity-50"
              >
                {loadingMore ? "Loading..." : "Load more"}
              </button>
            )}
          </div>
        )}
      </div>

      {/* Notes Modal */}
      {showNotesModal && selectedAppForDetails && (
        <NotesModal
          applicationId={selectedAppForDetails}
          notes={appNotes[selectedAppForDetails] || []}
          onClose={() => {
            setShowNotesModal(false);
            setSelectedAppForDetails(null);
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
          timeline={appTimeline[selectedAppForDetails] || []}
          onClose={() => {
            setShowTimelineModal(false);
            setSelectedAppForDetails(null);
          }}
        />
      )}

      {/* Analysis Modal */}
      {analysisApp && (
        <AnalysisModal
          candidateName={analysisApp.full_name}
          analysis={parseAnalysis(analysisApp.analysis_result)}
          onClose={() => setAnalysisApp(null)}
        />
      )}

      {/* Messages Modal */}
      {showMessagesModal && selectedAppForDetails && (
        <MessagesModal
          applicationId={selectedAppForDetails}
          candidateName={
            applications.find((a) => a.id === selectedAppForDetails)?.full_name
          }
          onClose={() => {
            // Opening the thread marked the candidate's messages read on the server.
            setApplications((prev) =>
              prev.map((a) => (a.id === selectedAppForDetails ? { ...a, unread_messages: 0 } : a))
            );
            notifyUnreadChanged();
            setShowMessagesModal(false);
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

// useSearchParams needs a Suspense boundary so the page can prerender.
export default function ApplicationsPage() {
  return (
    <Suspense>
      <Applications />
    </Suspense>
  );
}
