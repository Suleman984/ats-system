"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { jobAPI } from "@/lib/api";

export default function CreateJobPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    title: "",
    description: "",
    requirements: "",
    location: "",
    job_type: "full-time",
    salary_range: "",
    deadline: "",
    auto_shortlist: true,
    shortlist_criteria: "",
  });

  const [showCriteria, setShowCriteria] = useState(false);
  const [criteria, setCriteria] = useState({
    required_skills: "",
    min_experience: 0,
    required_languages: "",
    match_job_description: true,
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
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

      await jobAPI.create({
        ...formData,
        shortlist_criteria: shortlistCriteria,
      });
      alert("Job posted successfully!");
      router.push("/admin/dashboard/jobs");
    } catch (error: any) {
      alert(error.response?.data?.error || "Failed to create job");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-3xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">Post New Job</h1>
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
                  <p className="text-xs text-gray-500 mt-1">
                    List all skills candidates must have
                  </p>
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
                    id="matchJobDesc"
                    className="mr-2"
                    checked={criteria.match_job_description}
                    onChange={(e) =>
                      setCriteria({
                        ...criteria,
                        match_job_description: e.target.checked,
                      })
                    }
                  />
                  <label htmlFor="matchJobDesc" className="text-sm">
                    Match job description and requirements in scoring
                  </label>
                </div>

                <div className="bg-blue-50 border border-blue-200 rounded p-3 text-sm text-blue-800">
                  <strong>💡 How it works:</strong> When applications are
                  submitted, CVs will be automatically analyzed and scored based
                  on these criteria. Candidates with scores above 70% will be
                  auto-shortlisted.
                </div>
              </div>
            )}
          </div>

          <div className="flex gap-4 pt-4">
            <button
              type="submit"
              disabled={loading}
              className="flex-1 bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-300"
            >
              {loading ? "Creating..." : "Post Job"}
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
