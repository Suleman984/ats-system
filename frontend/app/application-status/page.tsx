"use client";

import { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { candidatePortalAPI, ApplicationStatus } from "@/lib/api";

export default function ApplicationStatusPage() {
  const searchParams = useSearchParams();
  const [email, setEmail] = useState("");
  const [applicationId, setApplicationId] = useState("");
  const [applications, setApplications] = useState<ApplicationStatus[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [searchMethod, setSearchMethod] = useState<"email" | "id">("email");

  // Auto-fill email from URL params if available
  useEffect(() => {
    const emailParam = searchParams.get("email");
    if (emailParam) {
      setEmail(emailParam);
      setSearchMethod("email");
    }
  }, [searchParams]);

  // Auto-search when email is set from URL
  useEffect(() => {
    const emailParam = searchParams.get("email");
    if (emailParam && email === emailParam) {
      // Small delay to ensure state is set
      const timer = setTimeout(() => {
        handleSearchByEmail();
      }, 500);
      return () => clearTimeout(timer);
    }
  }, [email, searchParams]);

  const handleSearchByEmail = async () => {
    if (!email) {
      setError("Please enter your email address");
      return;
    }

    setLoading(true);
    setError("");
    try {
      const response = await candidatePortalAPI.getByEmail(email);
      setApplications(response.data.applications || []);
      if (response.data.applications.length === 0) {
        setError("No applications found for this email address.");
      }
    } catch (err: any) {
      setError(
        err.response?.data?.error ||
          "Failed to fetch applications. Please try again."
      );
      setApplications([]);
    } finally {
      setLoading(false);
    }
  };

  const handleSearchById = async () => {
    if (!email || !applicationId) {
      setError("Please enter both email and application ID");
      return;
    }

    setLoading(true);
    setError("");
    try {
      const response = await candidatePortalAPI.checkStatus(
        email,
        applicationId
      );
      setApplications([response.data.application]);
    } catch (err: any) {
      setError(
        err.response?.data?.error ||
          "Application not found. Please check your details."
      );
      setApplications([]);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case "shortlisted":
        return "bg-green-100 text-green-800 border-green-300";
      case "rejected":
        return "bg-red-100 text-red-800 border-red-300";
      case "pending":
        return "bg-yellow-100 text-yellow-800 border-yellow-300";
      default:
        return "bg-gray-100 text-gray-800 border-gray-300";
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status.toLowerCase()) {
      case "shortlisted":
        return "✅";
      case "rejected":
        return "❌";
      case "pending":
        return "⏳";
      default:
        return "📋";
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4">
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-lg shadow-xl p-8 mb-8">
          <h1 className="text-4xl font-bold text-center mb-2 text-gray-800">
            📋 Application Status
          </h1>
          <p className="text-center text-gray-600 mb-8">
            Check the status of your job application
          </p>

          {/* Search Method Toggle */}
          <div className="flex justify-center gap-4 mb-6">
            <button
              onClick={() => {
                setSearchMethod("email");
                setApplications([]);
                setError("");
              }}
              className={`px-4 py-2 rounded-lg font-medium ${
                searchMethod === "email"
                  ? "bg-blue-600 text-white"
                  : "bg-gray-200 text-gray-700"
              }`}
            >
              Search by Email
            </button>
            <button
              onClick={() => {
                setSearchMethod("id");
                setApplications([]);
                setError("");
              }}
              className={`px-4 py-2 rounded-lg font-medium ${
                searchMethod === "id"
                  ? "bg-blue-600 text-white"
                  : "bg-gray-200 text-gray-700"
              }`}
            >
              Search by Application ID
            </button>
          </div>

          {/* Search Form */}
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">
                Email Address *
              </label>
              <input
                type="email"
                className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
                placeholder="your.email@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>

            {searchMethod === "id" && (
              <div>
                <label className="block text-sm font-medium mb-2">
                  Application ID *
                </label>
                <input
                  type="text"
                  className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  placeholder="Enter your application ID"
                  value={applicationId}
                  onChange={(e) => setApplicationId(e.target.value)}
                />
                <p className="text-xs text-gray-500 mt-1">
                  You received this in your confirmation email
                </p>
              </div>
            )}

            {error && (
              <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
                {error}
              </div>
            )}

            <button
              onClick={
                searchMethod === "email"
                  ? handleSearchByEmail
                  : handleSearchById
              }
              disabled={loading}
              className="w-full bg-blue-600 text-white py-3 rounded-lg hover:bg-blue-700 disabled:bg-blue-300 font-semibold"
            >
              {loading ? "Searching..." : "🔍 Check Status"}
            </button>
          </div>
        </div>

        {/* Results */}
        {applications.length > 0 && (
          <div className="bg-white rounded-lg shadow-xl p-8">
            <h2 className="text-2xl font-bold mb-6">
              Your Applications ({applications.length})
            </h2>
            <div className="space-y-4">
              {applications.map((app) => (
                <div
                  key={app.id}
                  className="border-2 rounded-lg p-6 hover:shadow-md transition"
                >
                  <div className="flex justify-between items-start mb-4">
                    <div className="flex-1">
                      <h3 className="text-xl font-semibold mb-2">
                        {app.job.title}
                      </h3>
                      <p className="text-sm text-gray-600 mb-1">
                        Application ID:{" "}
                        <code className="bg-gray-100 px-2 py-1 rounded text-xs">
                          {app.id}
                        </code>
                      </p>
                      <p className="text-sm text-gray-600">
                        Applied on:{" "}
                        {new Date(app.applied_at).toLocaleDateString("en-US", {
                          year: "numeric",
                          month: "long",
                          day: "numeric",
                        })}
                      </p>
                      {app.reviewed_at && (
                        <p className="text-sm text-gray-600">
                          Reviewed on:{" "}
                          {new Date(app.reviewed_at).toLocaleDateString(
                            "en-US",
                            {
                              year: "numeric",
                              month: "long",
                              day: "numeric",
                            }
                          )}
                        </p>
                      )}
                      {app.score > 0 && (
                        <p className="text-sm text-gray-600 mt-1">
                          Match Score:{" "}
                          <span className="font-semibold">{app.score}%</span>
                        </p>
                      )}
                    </div>
                    <div
                      className={`px-4 py-2 rounded-lg border-2 font-semibold flex items-center gap-2 ${getStatusColor(
                        app.status
                      )}`}
                    >
                      <span className="text-xl">
                        {getStatusIcon(app.status)}
                      </span>
                      <span className="capitalize">{app.status}</span>
                    </div>
                  </div>

                  {/* Status Messages */}
                  <div className="mt-4 p-4 rounded-lg">
                    {app.status.toLowerCase() === "pending" && (
                      <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                        <p className="text-yellow-800">
                          <strong>⏳ Your application is under review.</strong>
                          <br />
                          We're currently reviewing your application. You'll be
                          notified via email once a decision has been made.
                        </p>
                      </div>
                    )}
                    {app.status.toLowerCase() === "shortlisted" && (
                      <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                        <p className="text-green-800">
                          <strong>
                            🎉 Congratulations! You've been shortlisted!
                          </strong>
                          <br />
                          Your application has been selected for further review.
                          Our team will be in touch with you soon regarding the
                          next steps.
                        </p>
                      </div>
                    )}
                    {app.status.toLowerCase() === "rejected" && (
                      <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                        <p className="text-red-800">
                          <strong>Thank you for your interest.</strong>
                          <br />
                          After careful consideration, we've decided to move
                          forward with other candidates at this time. We
                          appreciate your time and encourage you to apply for
                          future opportunities.
                        </p>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Help Section */}
        <div className="bg-white rounded-lg shadow-xl p-8 mt-8">
          <h3 className="text-xl font-semibold mb-4">💡 Need Help?</h3>
          <div className="space-y-2 text-gray-600">
            <p>
              <strong>Can't find your application?</strong>
            </p>
            <ul className="list-disc list-inside space-y-1 ml-4">
              <li>
                Make sure you're using the same email address you used when
                applying
              </li>
              <li>Check your confirmation email for your Application ID</li>
              <li>
                If you applied recently, it may take a few minutes to appear
              </li>
            </ul>
            <p className="mt-4">
              <strong>Didn't receive a confirmation email?</strong>
            </p>
            <ul className="list-disc list-inside space-y-1 ml-4">
              <li>Check your spam/junk folder</li>
              <li>Verify the email address you used</li>
              <li>Contact the company directly if the issue persists</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}
