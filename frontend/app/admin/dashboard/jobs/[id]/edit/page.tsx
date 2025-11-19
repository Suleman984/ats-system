"use client";

import { useState, useEffect } from "react";
import { useRouter, useParams } from "next/navigation";
import { jobAPI } from "@/lib/api";
import { Job } from "@/lib/api";
import { toast } from "@/components/Toast";

export default function EditJobPage() {
  const router = useRouter();
  const params = useParams();
  const jobId = params.id as string;
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [formData, setFormData] = useState({
    title: "",
    description: "",
    requirements: "",
    location: "",
    job_type: "full-time",
    salary_range: "",
    deadline: "",
    status: "open",
  });

  const [showCriteria, setShowCriteria] = useState(false);
  const [criteria, setCriteria] = useState({
    required_skills: "",
    min_experience: 0,
    required_languages: "",
    match_job_description: true,
  });

  useEffect(() => {
    fetchJob();
  }, [jobId]);

  const fetchJob = async () => {
    try {
      const response = await jobAPI.getById(jobId);
      const job = response.data.job;

      // Format deadline from DateOnly to YYYY-MM-DD
      const deadlineDate = new Date(job.deadline);
      const formattedDeadline = deadlineDate.toISOString().split("T")[0];

      setFormData({
        title: job.title || "",
        description: job.description || "",
        requirements: job.requirements || "",
        location: job.location || "",
        job_type: job.job_type || "full-time",
        salary_range: job.salary_range || "",
        deadline: formattedDeadline,
        status: job.status || "open",
      });

      // Load criteria if exists
      if (job.shortlist_criteria) {
        try {
          const criteriaObj =
            typeof job.shortlist_criteria === "string"
              ? JSON.parse(job.shortlist_criteria)
              : job.shortlist_criteria;
          setCriteria({
            required_skills: Array.isArray(criteriaObj.required_skills)
              ? criteriaObj.required_skills.join(", ")
              : "",
            min_experience: criteriaObj.min_experience || 0,
            required_languages: Array.isArray(criteriaObj.required_languages)
              ? criteriaObj.required_languages.join(", ")
              : "",
            match_job_description: criteriaObj.match_job_description !== false,
          });
          setShowCriteria(true);
        } catch (e) {
          // Invalid criteria JSON, ignore
        }
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error || "Failed to load job");
      setTimeout(() => router.push("/admin/dashboard/jobs"), 1500);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      // Build shortlist criteria JSON if criteria is provided
      let shortlistCriteria = "";
      if (
        showCriteria &&
        (criteria.required_skills ||
          criteria.min_experience > 0 ||
          criteria.required_languages)
      ) {
        const criteriaObj = {
          required_skills: criteria.required_skills
            .split(",")
            .map((s) => s.trim())
            .filter((s) => s),
          min_experience: criteria.min_experience,
          required_languages: criteria.required_languages
            .split(",")
            .map((s) => s.trim())
            .filter((s) => s),
          match_job_description: criteria.match_job_description,
        };
        shortlistCriteria = JSON.stringify(criteriaObj);
      }

      await jobAPI.update(jobId, {
        ...formData,
        shortlist_criteria: shortlistCriteria,
      });
      toast.success("Job updated successfully!");
      setTimeout(() => router.push("/admin/dashboard/jobs"), 1000);
    } catch (error: any) {
      toast.error(error.response?.data?.error || "Failed to update job");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 p-6">
        <div className="max-w-3xl mx-auto">
          <div className="text-center py-12">Loading job details...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-3xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">Edit Job</h1>
        <form
          onSubmit={handleSubmit}
          className="bg-white rounded-lg shadow p-6 space-y-4"
        >
          <div>
            <label className="block text-sm font-medium mb-2">
              Job Title *
            </label>
            <input
              type="text"
              required
              className="w-full px-4 py-2 border rounded-lg"
              value={formData.title}
              onChange={(e) =>
                setFormData({ ...formData, title: e.target.value })
              }
              placeholder="e.g. Frontend Developer"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">
              Description *
            </label>
            <textarea
              required
              rows={5}
              className="w-full px-4 py-2 border rounded-lg"
              value={formData.description}
              onChange={(e) =>
                setFormData({ ...formData, description: e.target.value })
              }
              placeholder="Job description..."
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">
              Requirements
            </label>
            <textarea
              rows={4}
              className="w-full px-4 py-2 border rounded-lg"
              value={formData.requirements}
              onChange={(e) =>
                setFormData({ ...formData, requirements: e.target.value })
              }
              placeholder="Required skills and experience..."
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2">Location</label>
              <input
                type="text"
                className="w-full px-4 py-2 border rounded-lg"
                value={formData.location}
                onChange={(e) =>
                  setFormData({ ...formData, location: e.target.value })
                }
                placeholder="e.g. Remote, Islamabad"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">Job Type</label>
              <select
                className="w-full px-4 py-2 border rounded-lg"
                value={formData.job_type}
                onChange={(e) =>
                  setFormData({ ...formData, job_type: e.target.value })
                }
              >
                <option value="full-time">Full Time</option>
                <option value="part-time">Part Time</option>
                <option value="contract">Contract</option>
                <option value="internship">Internship</option>
              </select>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2">
                Salary Range
              </label>
              <input
                type="text"
                className="w-full px-4 py-2 border rounded-lg"
                value={formData.salary_range}
                onChange={(e) =>
                  setFormData({ ...formData, salary_range: e.target.value })
                }
                placeholder="e.g. $50k - $80k"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">
                Application Deadline *
              </label>
              <input
                type="date"
                required
                className="w-full px-4 py-2 border rounded-lg"
                value={formData.deadline}
                onChange={(e) =>
                  setFormData({ ...formData, deadline: e.target.value })
                }
              />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">Status</label>
            <select
              className="w-full px-4 py-2 border rounded-lg"
              value={formData.status}
              onChange={(e) =>
                setFormData({ ...formData, status: e.target.value })
              }
            >
              <option value="open">Open</option>
              <option value="closed">Closed</option>
              <option value="archived">Archived</option>
            </select>
          </div>

          {/* Shortlisting Criteria Section */}
          <div className="border-t pt-4 mt-4">
            <button
              type="button"
              onClick={() => setShowCriteria(!showCriteria)}
              className="flex items-center justify-between w-full text-left mb-4"
            >
              <div>
                <h3 className="text-lg font-semibold">
                  Shortlisting Criteria (Optional)
                </h3>
                <p className="text-sm text-gray-500">
                  Set criteria for automatic CV matching and scoring
                </p>
              </div>
              <span className="text-2xl">{showCriteria ? "−" : "+"}</span>
            </button>

            {showCriteria && (
              <div className="space-y-4 bg-gray-50 p-4 rounded-lg">
                <div>
                  <label className="block text-sm font-medium mb-2">
                    Required Skills (comma-separated)
                  </label>
                  <input
                    type="text"
                    className="w-full px-4 py-2 border rounded-lg"
                    value={criteria.required_skills}
                    onChange={(e) =>
                      setCriteria({
                        ...criteria,
                        required_skills: e.target.value,
                      })
                    }
                    placeholder="e.g. JavaScript, React, Node.js, TypeScript"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">
                    Minimum Years of Experience
                  </label>
                  <input
                    type="number"
                    min="0"
                    className="w-full px-4 py-2 border rounded-lg"
                    value={criteria.min_experience}
                    onChange={(e) =>
                      setCriteria({
                        ...criteria,
                        min_experience: parseInt(e.target.value) || 0,
                      })
                    }
                    placeholder="e.g. 3"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">
                    Required Languages (comma-separated)
                  </label>
                  <input
                    type="text"
                    className="w-full px-4 py-2 border rounded-lg"
                    value={criteria.required_languages}
                    onChange={(e) =>
                      setCriteria({
                        ...criteria,
                        required_languages: e.target.value,
                      })
                    }
                    placeholder="e.g. English, Spanish"
                  />
                </div>

                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="matchJobDescEdit"
                    className="mr-2"
                    checked={criteria.match_job_description}
                    onChange={(e) =>
                      setCriteria({
                        ...criteria,
                        match_job_description: e.target.checked,
                      })
                    }
                  />
                  <label htmlFor="matchJobDescEdit" className="text-sm">
                    Match job description and requirements in scoring
                  </label>
                </div>
              </div>
            )}
          </div>

          <div className="flex gap-4 pt-4">
            <button
              type="submit"
              disabled={submitting}
              className="flex-1 bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-300"
            >
              {submitting ? "Updating..." : "Update Job"}
            </button>
            <button
              type="button"
              onClick={() => router.back()}
              className="px-6 py-2 border rounded-lg hover:bg-gray-50"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
