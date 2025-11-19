"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { jobAPI, Job, uploadAPI } from "@/lib/api";
import { toast } from "@/components/Toast";

export default function ManualAddCandidatePage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [uploadingCV, setUploadingCV] = useState(false);
  const [cvUploadMessage, setCvUploadMessage] = useState("");
  const [jobs, setJobs] = useState<Job[]>([]);
  const [formData, setFormData] = useState({
    job_id: "",
    full_name: "",
    email: "",
    phone: "",
    resume_url: "",
    cover_letter: "",
    years_of_experience: 0,
    current_position: "",
    linkedin_url: "",
    portfolio_url: "",
    status: "pending",
    notes: "",
    resume_type: "url" as "url" | "file",
  });

  useEffect(() => {
    fetchJobs();
  }, []);

  const fetchJobs = async () => {
    try {
      const response = await jobAPI.getAll();
      setJobs(response.data.jobs);
    } catch (error) {
      console.error("Failed to fetch jobs:", error);
      toast.error("Failed to load jobs");
    }
  };

  const handleCVFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const allowedTypes = [
      "application/pdf",
      "application/msword",
      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    ];
    if (!allowedTypes.includes(file.type)) {
      setCvUploadMessage("");
      toast.error(
        "Invalid file type. Please upload PDF, DOC, or DOCX files only."
      );
      return;
    }

    if (file.size > 10 * 1024 * 1024) {
      setCvUploadMessage("");
      toast.error("File size exceeds 10MB limit.");
      return;
    }

    setUploadingCV(true);
    setCvUploadMessage("");
    try {
      const response = await uploadAPI.uploadCV(file);
      setFormData({ ...formData, resume_url: response.data.file_url });
      setCvUploadMessage("✓ CV uploaded successfully!");
      setTimeout(() => setCvUploadMessage(""), 5000);
    } catch (error: any) {
      setCvUploadMessage("");
      toast.error(error.response?.data?.error || "Failed to upload CV");
    } finally {
      setUploadingCV(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.job_id) {
      toast.error("Please select a job");
      return;
    }

    if (!formData.resume_url) {
      toast.error("Please provide a CV/resume (either URL or file upload)");
      return;
    }

    setLoading(true);
    try {
      const response = await fetch("/api/candidates/manual", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${localStorage.getItem("token")}`,
        },
        body: JSON.stringify(formData),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Failed to add candidate");
      }

      toast.success(
        "Candidate added successfully! CV will be analyzed automatically."
      );
      setTimeout(() => router.push("/admin/dashboard/applications"), 1500);
    } catch (error: any) {
      toast.error(
        error.message || "Failed to add candidate. Please try again."
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-3xl mx-auto">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-3xl font-bold">Add Candidate Manually</h1>
          <button
            onClick={() => router.back()}
            className="px-4 py-2 border rounded-lg hover:bg-gray-100"
          >
            Cancel
          </button>
        </div>

        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-blue-800">
            <strong>💡 Use this feature to:</strong> Add candidates that the AI
            might have missed, import candidates from other sources, or manually
            add promising candidates you found elsewhere. The system will
            automatically analyze their CV and calculate a match score.
          </p>
        </div>

        <form
          onSubmit={handleSubmit}
          className="bg-white rounded-lg shadow p-6 space-y-4"
        >
          <div>
            <label className="block text-sm font-medium mb-2">
              Select Job *
            </label>
            <select
              required
              className="w-full px-4 py-2 border rounded-lg"
              value={formData.job_id}
              onChange={(e) =>
                setFormData({ ...formData, job_id: e.target.value })
              }
            >
              <option value="">Select a job...</option>
              {jobs.map((job) => (
                <option key={job.id} value={job.id}>
                  {job.title}
                </option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2">
                Full Name *
              </label>
              <input
                type="text"
                required
                className="w-full px-4 py-2 border rounded-lg"
                value={formData.full_name}
                onChange={(e) =>
                  setFormData({ ...formData, full_name: e.target.value })
                }
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">Email *</label>
              <input
                type="email"
                required
                className="w-full px-4 py-2 border rounded-lg"
                value={formData.email}
                onChange={(e) =>
                  setFormData({ ...formData, email: e.target.value })
                }
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2">Phone</label>
              <input
                type="tel"
                className="w-full px-4 py-2 border rounded-lg"
                value={formData.phone}
                onChange={(e) =>
                  setFormData({ ...formData, phone: e.target.value })
                }
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">
                Years of Experience
              </label>
              <input
                type="number"
                min="0"
                className="w-full px-4 py-2 border rounded-lg"
                value={formData.years_of_experience}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    years_of_experience: parseInt(e.target.value) || 0,
                  })
                }
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Current Position
            </label>
            <input
              type="text"
              className="w-full px-4 py-2 border rounded-lg"
              value={formData.current_position}
              onChange={(e) =>
                setFormData({ ...formData, current_position: e.target.value })
              }
              placeholder="e.g., Senior Developer"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2">
                LinkedIn URL
              </label>
              <input
                type="url"
                className="w-full px-4 py-2 border rounded-lg"
                value={formData.linkedin_url}
                onChange={(e) =>
                  setFormData({ ...formData, linkedin_url: e.target.value })
                }
                placeholder="https://linkedin.com/in/..."
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">
                Portfolio URL
              </label>
              <input
                type="url"
                className="w-full px-4 py-2 border rounded-lg"
                value={formData.portfolio_url}
                onChange={(e) =>
                  setFormData({ ...formData, portfolio_url: e.target.value })
                }
                placeholder="https://portfolio.com"
              />
            </div>
          </div>

          {/* CV/Resume Section */}
          <div>
            <label className="block text-sm font-medium mb-2">
              CV/Resume *
            </label>
            <div className="mb-3">
              <div className="flex gap-4 mb-3">
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="resume_type"
                    value="url"
                    checked={formData.resume_type === "url"}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        resume_type: "url" as const,
                      })
                    }
                    className="mr-2"
                  />
                  Provide URL
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="resume_type"
                    value="file"
                    checked={formData.resume_type === "file"}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        resume_type: "file" as const,
                      })
                    }
                    className="mr-2"
                  />
                  Upload File
                </label>
              </div>
            </div>
            {formData.resume_type === "url" ? (
              <div>
                <input
                  type="url"
                  required
                  className="w-full px-4 py-2 border rounded-lg"
                  value={formData.resume_url}
                  onChange={(e) =>
                    setFormData({ ...formData, resume_url: e.target.value })
                  }
                  placeholder="https://drive.google.com/your-resume"
                />
              </div>
            ) : (
              <div>
                <input
                  type="file"
                  accept=".pdf,.doc,.docx"
                  onChange={handleCVFileChange}
                  className="w-full px-4 py-2 border rounded-lg"
                  disabled={uploadingCV}
                />
                {uploadingCV && (
                  <p className="text-sm text-blue-600 mt-1">Uploading...</p>
                )}
                {cvUploadMessage && (
                  <p className="text-sm text-green-600 mt-1">
                    {cvUploadMessage}
                  </p>
                )}
              </div>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Cover Letter (Optional)
            </label>
            <textarea
              rows={4}
              className="w-full px-4 py-2 border rounded-lg"
              value={formData.cover_letter}
              onChange={(e) =>
                setFormData({ ...formData, cover_letter: e.target.value })
              }
              placeholder="Cover letter or notes about the candidate..."
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Admin Notes (Optional)
            </label>
            <textarea
              rows={3}
              className="w-full px-4 py-2 border rounded-lg"
              value={formData.notes}
              onChange={(e) =>
                setFormData({ ...formData, notes: e.target.value })
              }
              placeholder="Why are you adding this candidate manually? (e.g., Found on LinkedIn, Referred by team, etc.)"
            />
            <p className="text-xs text-gray-500 mt-1">
              This note will be saved in activity logs for reference.
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Initial Status
            </label>
            <select
              className="w-full px-4 py-2 border rounded-lg"
              value={formData.status}
              onChange={(e) =>
                setFormData({ ...formData, status: e.target.value })
              }
            >
              <option value="pending">Pending</option>
              <option value="shortlisted">Shortlisted</option>
              <option value="rejected">Rejected</option>
            </select>
          </div>

          <div className="flex gap-4 pt-4">
            <button
              type="submit"
              disabled={loading || uploadingCV}
              className="flex-1 bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-300"
            >
              {loading ? "Adding Candidate..." : "Add Candidate"}
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
