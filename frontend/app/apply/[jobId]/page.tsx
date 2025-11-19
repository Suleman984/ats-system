"use client";

import { useEffect, useState, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import { applicationAPI, jobAPI, uploadAPI, Job } from "@/lib/api";

export default function ApplyPage() {
  const params = useParams();
  const router = useRouter();
  const jobId = params.jobId as string;
  const [job, setJob] = useState<Job | null>(null);
  const [loading, setLoading] = useState(false);
  const [uploadingCV, setUploadingCV] = useState(false);
  const [uploadingPortfolio, setUploadingPortfolio] = useState(false);
  const cvFileRef = useRef<HTMLInputElement>(null);
  const portfolioFileRef = useRef<HTMLInputElement>(null);
  const [formData, setFormData] = useState({
    job_id: jobId,
    full_name: "",
    email: "",
    phone: "",
    years_of_experience: 0,
    current_position: "",
    linkedin_url: "",
    portfolio_url: "",
    cover_letter: "",
    resume_url: "",
    resume_type: "url" as "url" | "file",
    portfolio_type: "url" as "url" | "file",
  });

  useEffect(() => {
    fetchJob();
  }, [jobId]);

  const fetchJob = async () => {
    try {
      // Try to get job from public endpoint - we need companyId
      // For now, we'll handle this in the form
    } catch (error) {
      console.error("Failed to fetch job:", error);
    }
  };

  const handleCVFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file type
    const allowedTypes = [
      "application/pdf",
      "application/msword",
      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    ];
    if (!allowedTypes.includes(file.type)) {
      alert("Invalid file type. Please upload PDF, DOC, or DOCX files only.");
      return;
    }

    // Validate file size (10MB)
    if (file.size > 10 * 1024 * 1024) {
      alert("File size exceeds 10MB limit.");
      return;
    }

    setUploadingCV(true);
    try {
      const response = await uploadAPI.uploadCV(file);
      setFormData({ ...formData, resume_url: response.data.file_url });
      alert("CV uploaded successfully!");
    } catch (error: any) {
      alert(error.response?.data?.error || "Failed to upload CV");
    } finally {
      setUploadingCV(false);
    }
  };

  const handlePortfolioFileChange = async (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file type
    const allowedTypes = [
      "application/pdf",
      "application/zip",
      "application/x-rar-compressed",
      "application/x-7z-compressed",
    ];
    if (!allowedTypes.includes(file.type)) {
      alert(
        "Invalid file type. Please upload PDF, ZIP, RAR, or 7Z files only."
      );
      return;
    }

    // Validate file size (10MB)
    if (file.size > 10 * 1024 * 1024) {
      alert("File size exceeds 10MB limit.");
      return;
    }

    setUploadingPortfolio(true);
    try {
      const response = await uploadAPI.uploadPortfolio(file);
      setFormData({ ...formData, portfolio_url: response.data.file_url });
      alert("Portfolio uploaded successfully!");
    } catch (error: any) {
      alert(error.response?.data?.error || "Failed to upload portfolio");
    } finally {
      setUploadingPortfolio(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validate resume
    if (!formData.resume_url) {
      alert("Please provide a CV/resume (either URL or file upload)");
      return;
    }

    setLoading(true);
    try {
      const response = await applicationAPI.submit(formData);
      const applicationId = response.data.application?.id;

      const message = applicationId
        ? `Application submitted successfully!\n\nYour Application ID: ${applicationId}\n\nCheck your email for confirmation, or visit /application-status to track your application status anytime.`
        : "Application submitted successfully! Check your email for confirmation.";

      alert(message);

      // Optionally redirect to status page
      if (applicationId && formData.email) {
        const redirect = confirm(
          "Would you like to check your application status now?"
        );
        if (redirect) {
          router.push(
            `/application-status?email=${encodeURIComponent(formData.email)}`
          );
        } else {
          router.push("/");
        }
      } else {
        router.push("/");
      }
    } catch (error: any) {
      if (error.response?.data?.message) {
        alert(error.response.data.message);
      } else {
        alert("Failed to submit application. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 py-12">
      <div className="max-w-3xl mx-auto px-6">
        <div className="bg-white rounded-lg shadow-lg p-8">
          <h1 className="text-3xl font-bold mb-6">Job Application</h1>
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-2">
                  Full Name *
                </label>
                <input
                  type="text"
                  required
                  className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  value={formData.full_name}
                  onChange={(e) =>
                    setFormData({ ...formData, full_name: e.target.value })
                  }
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">
                  Email *
                </label>
                <input
                  type="email"
                  required
                  className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
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
                  className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  value={formData.phone}
                  onChange={(e) =>
                    setFormData({ ...formData, phone: e.target.value })
                  }
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">
                  Years of Experience *
                </label>
                <input
                  type="number"
                  required
                  min="0"
                  className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  value={formData.years_of_experience}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      years_of_experience: parseInt(e.target.value),
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
                className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                value={formData.current_position}
                onChange={(e) =>
                  setFormData({ ...formData, current_position: e.target.value })
                }
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-2">
                  LinkedIn URL
                </label>
                <input
                  type="url"
                  className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  value={formData.linkedin_url}
                  onChange={(e) =>
                    setFormData({ ...formData, linkedin_url: e.target.value })
                  }
                  placeholder="https://linkedin.com/in/yourprofile"
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
                    className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                    value={formData.resume_url}
                    onChange={(e) =>
                      setFormData({ ...formData, resume_url: e.target.value })
                    }
                    placeholder="https://drive.google.com/your-resume"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Paste the shareable link to your resume
                  </p>
                </div>
              ) : (
                <div>
                  <input
                    ref={cvFileRef}
                    type="file"
                    accept=".pdf,.doc,.docx"
                    onChange={handleCVFileChange}
                    className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                    disabled={uploadingCV}
                  />
                  {uploadingCV && (
                    <p className="text-sm text-blue-600 mt-1">Uploading...</p>
                  )}
                  {formData.resume_url && !uploadingCV && (
                    <p className="text-sm text-green-600 mt-1">
                      ✓ CV uploaded successfully
                    </p>
                  )}
                  <p className="text-xs text-gray-500 mt-1">
                    Accepted formats: PDF, DOC, DOCX (Max 10MB)
                  </p>
                </div>
              )}
            </div>

            {/* Portfolio Section */}
            <div>
              <label className="block text-sm font-medium mb-2">
                Portfolio (Optional)
              </label>
              <div className="mb-3">
                <div className="flex gap-4 mb-3">
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="portfolio_type"
                      value="url"
                      checked={formData.portfolio_type === "url"}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          portfolio_type: "url" as const,
                        })
                      }
                      className="mr-2"
                    />
                    Provide URL
                  </label>
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="portfolio_type"
                      value="file"
                      checked={formData.portfolio_type === "file"}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          portfolio_type: "file" as const,
                        })
                      }
                      className="mr-2"
                    />
                    Upload File
                  </label>
                </div>
              </div>
              {formData.portfolio_type === "url" ? (
                <div>
                  <input
                    type="url"
                    className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                    value={formData.portfolio_url}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        portfolio_url: e.target.value,
                      })
                    }
                    placeholder="https://yourportfolio.com"
                  />
                </div>
              ) : (
                <div>
                  <input
                    ref={portfolioFileRef}
                    type="file"
                    accept=".pdf,.zip,.rar,.7z"
                    onChange={handlePortfolioFileChange}
                    className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                    disabled={uploadingPortfolio}
                  />
                  {uploadingPortfolio && (
                    <p className="text-sm text-blue-600 mt-1">Uploading...</p>
                  )}
                  {formData.portfolio_url && !uploadingPortfolio && (
                    <p className="text-sm text-green-600 mt-1">
                      ✓ Portfolio uploaded successfully
                    </p>
                  )}
                  <p className="text-xs text-gray-500 mt-1">
                    Accepted formats: PDF, ZIP, RAR, 7Z (Max 10MB)
                  </p>
                </div>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">
                Cover Letter (Optional)
              </label>
              <textarea
                rows={5}
                className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                value={formData.cover_letter}
                onChange={(e) =>
                  setFormData({ ...formData, cover_letter: e.target.value })
                }
                placeholder="Tell us why you're a great fit for this position..."
              />
            </div>
            <button
              type="submit"
              disabled={loading || uploadingCV || uploadingPortfolio}
              className="w-full bg-blue-600 text-white py-3 rounded-lg hover:bg-blue-700 disabled:bg-blue-300 font-semibold"
            >
              {loading ? "Submitting..." : "Submit Application"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
