"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { jobAPI, errorMessage, type Job } from "@/lib/api";
import { toast } from "@/components/Toast";

interface JobFormProps {
  jobId?: string;
  // Where to return after saving, e.g. "/admin/dashboard/jobs".
  basePath: string;
}

interface CriteriaState {
  required_skills: string;
  min_experience: number;
  required_languages: string;
  match_job_description: boolean;
}

const emptyJob = {
  title: "",
  description: "",
  requirements: "",
  location: "",
  job_type: "full-time",
  salary_range: "",
  deadline: "",
  status: "open",
};

const emptyCriteria: CriteriaState = {
  required_skills: "",
  min_experience: 0,
  required_languages: "",
  match_job_description: true,
};

const splitList = (value: string) =>
  value
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);

function parseCriteria(raw: Job["shortlist_criteria"]): CriteriaState | null {
  if (!raw) return null;
  try {
    const c = typeof raw === "string" ? JSON.parse(raw) : raw;
    return {
      required_skills: Array.isArray(c.required_skills) ? c.required_skills.join(", ") : "",
      min_experience: c.min_experience || 0,
      required_languages: Array.isArray(c.required_languages) ? c.required_languages.join(", ") : "",
      match_job_description: c.match_job_description !== false,
    };
  } catch {
    return null;
  }
}

function serializeCriteria(show: boolean, c: CriteriaState): string {
  if (!show || (!c.required_skills && c.min_experience <= 0 && !c.required_languages)) return "";
  return JSON.stringify({
    required_skills: splitList(c.required_skills),
    min_experience: c.min_experience,
    required_languages: splitList(c.required_languages),
    match_job_description: c.match_job_description,
  });
}

const inputClass = "w-full px-4 py-2 border rounded-lg";
// Local calendar date; toISOString() would give tomorrow/yesterday near midnight outside UTC.
const today = () => new Date().toLocaleDateString("en-CA");

export default function JobForm({ jobId, basePath }: JobFormProps) {
  const router = useRouter();
  const isEdit = Boolean(jobId);
  const [loading, setLoading] = useState(isEdit);
  const [submitting, setSubmitting] = useState(false);
  const [formData, setFormData] = useState(emptyJob);
  const [showCriteria, setShowCriteria] = useState(false);
  const [criteria, setCriteria] = useState(emptyCriteria);
  const [savedDeadline, setSavedDeadline] = useState("");

  useEffect(() => {
    if (!jobId) return;
    jobAPI
      .getById(jobId)
      .then(({ data: { job } }) => {
        setFormData({
          title: job.title,
          description: job.description,
          requirements: job.requirements || "",
          location: job.location || "",
          job_type: job.job_type || "full-time",
          salary_range: job.salary_range || "",
          deadline: job.deadline.slice(0, 10),
          status: job.status,
        });
        setSavedDeadline(job.deadline.slice(0, 10));
        const saved = parseCriteria(job.shortlist_criteria);
        if (saved) {
          setCriteria(saved);
          setShowCriteria(true);
        }
      })
      .catch((error) => {
        toast.error(errorMessage(error, "Failed to load job"));
        router.push(basePath);
      })
      .finally(() => setLoading(false));
  }, [jobId, basePath, router]);

  const update = (field: keyof typeof emptyJob) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) =>
      setFormData((prev) => ({ ...prev, [field]: e.target.value }));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    const payload = {
      ...formData,
      auto_shortlist: true,
      shortlist_criteria: serializeCriteria(showCriteria, criteria),
    };
    try {
      if (jobId) {
        await jobAPI.update(jobId, payload);
        toast.success("Job updated successfully!");
      } else {
        await jobAPI.create(payload);
        toast.success("Job posted successfully!");
      }
      router.push(basePath);
    } catch (error) {
      toast.error(errorMessage(error, isEdit ? "Failed to update job" : "Failed to create job"));
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return <div className="min-h-screen bg-gray-50 p-6 text-center py-12">Loading job details...</div>;
  }

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-3xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">{isEdit ? "Edit Job" : "Post New Job"}</h1>
        <form onSubmit={handleSubmit} className="bg-white rounded-lg shadow p-6 space-y-4">
          <Field label="Job Title *">
            <input type="text" required className={inputClass} value={formData.title}
              onChange={update("title")} placeholder="e.g. Frontend Developer" />
          </Field>
          <Field label="Description *">
            <textarea required rows={5} className={inputClass} value={formData.description}
              onChange={update("description")} placeholder="Job description..." />
          </Field>
          <Field label="Requirements">
            <textarea rows={4} className={inputClass} value={formData.requirements}
              onChange={update("requirements")} placeholder="Required skills and experience..." />
          </Field>

          <div className="grid grid-cols-2 gap-4">
            <Field label="Location">
              <input type="text" className={inputClass} value={formData.location}
                onChange={update("location")} placeholder="e.g. Remote, Islamabad" />
            </Field>
            <Field label="Job Type">
              <select className={inputClass} value={formData.job_type} onChange={update("job_type")}>
                <option value="full-time">Full Time</option>
                <option value="part-time">Part Time</option>
                <option value="contract">Contract</option>
                <option value="internship">Internship</option>
              </select>
            </Field>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <Field label="Salary Range">
              <input type="text" className={inputClass} value={formData.salary_range}
                onChange={update("salary_range")} placeholder="e.g. $50k - $80k" />
            </Field>
            {/* An expired job can still be edited without moving its deadline. */}
            <Field label="Application Deadline *" hint="Cannot select past dates">
              <input type="date" required className={inputClass}
                min={savedDeadline && savedDeadline < today() ? savedDeadline : today()}
                value={formData.deadline} onChange={update("deadline")} />
            </Field>
          </div>

          {isEdit && (
            <Field label="Status">
              <select className={inputClass} value={formData.status} onChange={update("status")}>
                <option value="open">Open</option>
                <option value="closed">Closed</option>
                <option value="archived">Archived</option>
              </select>
            </Field>
          )}

          <div className="border-t pt-4 mt-4">
            <button type="button" onClick={() => setShowCriteria(!showCriteria)}
              className="flex items-center justify-between w-full text-left mb-4" aria-expanded={showCriteria}>
              <div>
                <h3 className="text-lg font-semibold">Shortlisting Criteria (Optional)</h3>
                <p className="text-sm text-gray-500">Set criteria for automatic CV matching and scoring</p>
              </div>
              <span className="text-2xl" aria-hidden>{showCriteria ? "−" : "+"}</span>
            </button>

            {showCriteria && (
              <div className="space-y-4 bg-gray-50 p-4 rounded-lg">
                <Field label="Required Skills (comma-separated)" hint="List all skills candidates must have">
                  <input type="text" className={inputClass} value={criteria.required_skills}
                    onChange={(e) => setCriteria({ ...criteria, required_skills: e.target.value })}
                    placeholder="e.g. JavaScript, React, Node.js, TypeScript" />
                </Field>
                <Field label="Minimum Years of Experience">
                  <input type="number" min="0" className={inputClass} value={criteria.min_experience}
                    onChange={(e) => setCriteria({ ...criteria, min_experience: parseInt(e.target.value) || 0 })}
                    placeholder="e.g. 3" />
                </Field>
                <Field label="Required Languages (comma-separated)">
                  <input type="text" className={inputClass} value={criteria.required_languages}
                    onChange={(e) => setCriteria({ ...criteria, required_languages: e.target.value })}
                    placeholder="e.g. English, Spanish" />
                </Field>
                <label className="flex items-center text-sm">
                  <input type="checkbox" className="mr-2" checked={criteria.match_job_description}
                    onChange={(e) => setCriteria({ ...criteria, match_job_description: e.target.checked })} />
                  Match job description and requirements in scoring
                </label>
                <div className="bg-blue-50 border border-blue-200 rounded p-3 text-sm text-blue-800">
                  <strong>How it works:</strong> each new CV is scored against these criteria.
                  Scores appear on the Applications page; you decide who to shortlist.
                </div>
              </div>
            )}
          </div>

          <div className="flex gap-4 pt-4">
            <button type="submit" disabled={submitting}
              className="flex-1 bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-300">
              {submitting ? "Saving..." : isEdit ? "Update Job" : "Post Job"}
            </button>
            <button type="button" onClick={() => router.back()} className="px-6 py-2 border rounded-lg hover:bg-gray-50">
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="block text-sm font-medium mb-2">{label}</span>
      {children}
      {hint && <span className="block text-xs text-gray-500 mt-1">{hint}</span>}
    </label>
  );
}
