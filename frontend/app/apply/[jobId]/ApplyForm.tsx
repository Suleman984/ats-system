"use client";

import { useRef, useState } from "react";
import Link from "next/link";
import { applicationAPI, uploadAPI, errorMessage } from "@/lib/api";
import { toast } from "@/components/Toast";

// Browsers report inconsistent MIME types for archives, so files are checked by extension;
// the server re-validates the file signature.
const hasExtension = (file: File, extensions: string[]) =>
  extensions.some((ext) => file.name.toLowerCase().endsWith(ext));

export default function ApplyForm({
  jobId,
  jobTitle,
}: {
  jobId: string;
  jobTitle: string;
}) {
  const [submitted, setSubmitted] = useState<{ id: string; email: string } | null>(null);
  const [loading, setLoading] = useState(false);
  const [uploadingCV, setUploadingCV] = useState(false);
  const [uploadingPortfolio, setUploadingPortfolio] = useState(false);
  const [cvUploadMessage, setCvUploadMessage] = useState("");
  const [portfolioUploadMessage, setPortfolioUploadMessage] = useState("");
  const cvFileRef = useRef<HTMLInputElement>(null);
  const portfolioFileRef = useRef<HTMLInputElement>(null);
  const successRef = useRef<HTMLDivElement>(null);
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
    resume_upload_id: "",
    portfolio_upload_id: "",
    resume_type: "url" as "url" | "file",
    portfolio_type: "url" as "url" | "file",
  });

  const handleCVFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!hasExtension(file, [".pdf", ".doc", ".docx"])) {
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
      setFormData((prev) => ({ ...prev, resume_upload_id: response.data.upload_id }));
      setCvUploadMessage(`✓ ${response.data.file_name} uploaded`);
    } catch (error) {
      setCvUploadMessage("");
      toast.error(errorMessage(error, "Failed to upload CV"));
    } finally {
      setUploadingCV(false);
    }
  };

  const handlePortfolioFileChange = async (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!hasExtension(file, [".pdf", ".zip", ".rar", ".7z"])) {
      setPortfolioUploadMessage("");
      toast.error(
        "Invalid file type. Please upload PDF, ZIP, RAR, or 7Z files only."
      );
      return;
    }

    if (file.size > 10 * 1024 * 1024) {
      setPortfolioUploadMessage("");
      toast.error("File size exceeds 10MB limit.");
      return;
    }

    setUploadingPortfolio(true);
    setPortfolioUploadMessage("");
    try {
      const response = await uploadAPI.uploadPortfolio(file);
      setFormData((prev) => ({ ...prev, portfolio_upload_id: response.data.upload_id }));
      setPortfolioUploadMessage(`✓ ${response.data.file_name} uploaded`);
    } catch (error) {
      setPortfolioUploadMessage("");
      toast.error(errorMessage(error, "Failed to upload portfolio"));
    } finally {
      setUploadingPortfolio(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const usingCVFile = formData.resume_type === "file";
    if (usingCVFile ? !formData.resume_upload_id : !formData.resume_url) {
      toast.error("Please provide a CV/resume (either URL or file upload)");
      return;
    }

    setLoading(true);
    try {
      // Send only the source the candidate picked: a link or an upload ID.
      const usingPortfolioFile = formData.portfolio_type === "file";
      const response = await applicationAPI.submit({
        job_id: formData.job_id,
        full_name: formData.full_name,
        email: formData.email,
        phone: formData.phone,
        years_of_experience: formData.years_of_experience,
        current_position: formData.current_position,
        linkedin_url: formData.linkedin_url,
        cover_letter: formData.cover_letter,
        ...(usingCVFile
          ? { resume_upload_id: formData.resume_upload_id }
          : { resume_url: formData.resume_url }),
        ...(usingPortfolioFile
          ? formData.portfolio_upload_id && { portfolio_upload_id: formData.portfolio_upload_id }
          : formData.portfolio_url && { portfolio_url: formData.portfolio_url }),
      });
      setSubmitted({ id: response.data.application.id, email: formData.email });
      requestAnimationFrame(() => successRef.current?.scrollIntoView({ block: "start" }));
    } catch (error) {
      toast.error(errorMessage(error, "Failed to submit application. Please try again."));
    } finally {
      setLoading(false);
    }
  };

  if (submitted) {
    return (
      <div ref={successRef} className="bg-white rounded-lg shadow-lg p-8 text-center">
        <div className="w-14 h-14 mx-auto mb-4 rounded-full bg-green-100 text-green-600 flex items-center justify-center text-2xl">
          ✓
        </div>
        <h1 className="text-2xl font-bold mb-2">Application submitted</h1>
        <p className="text-gray-600 mb-6">
          Thanks for applying for {jobTitle}. We&apos;ve also emailed
          these details to {submitted.email}.
        </p>
        <div className="text-left bg-gray-50 border rounded-lg p-4 mb-6">
          <p className="text-sm font-medium text-gray-700 mb-1">Your application ID</p>
          <div className="flex items-center gap-2">
            <code className="flex-1 text-sm break-all">{submitted.id}</code>
            <button
              type="button"
              onClick={() =>
                navigator.clipboard
                  .writeText(submitted.id)
                  .then(() => toast.success("Application ID copied"))
                  .catch(() => toast.error("Couldn't copy — please copy it manually"))
              }
              className="text-sm text-blue-600 hover:text-blue-700 font-medium"
            >
              Copy
            </button>
          </div>
          <p className="text-xs text-gray-500 mt-2">
            Keep this ID — you&apos;ll need it with your email to check your status or
            message the hiring team.
          </p>
        </div>
        <Link
          href={`/application-status?email=${encodeURIComponent(submitted.email)}&applicationId=${submitted.id}`}
          className="block w-full bg-blue-600 text-white py-3 rounded-lg hover:bg-blue-700 font-semibold"
        >
          Check application status
        </Link>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-lg p-8">
      <h2 className="text-2xl font-bold mb-6">Apply for this job</h2>
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
                  onChange={() =>
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
                  onChange={() =>
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
              {cvUploadMessage && (
                <p className="text-sm text-green-600 mt-1">
                  {cvUploadMessage}
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
                  onChange={() =>
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
                  onChange={() =>
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
              {portfolioUploadMessage && (
                <p className="text-sm text-green-600 mt-1">
                  {portfolioUploadMessage}
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
  );
}
