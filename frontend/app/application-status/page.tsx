"use client";

import { Suspense, useEffect, useState } from "react";
import { candidatePortalAPI, errorMessage, type ApplicationStatus } from "@/lib/api";
import { toast } from "@/components/Toast";
import { useSearchParams } from "next/navigation";

function ApplicationStatus() {
  const searchParams = useSearchParams();
  const [email, setEmail] = useState("");
  const [applicationId, setApplicationId] = useState("");
  const [application, setApplication] = useState<ApplicationStatus | null>(
    null
  );
  const [loading, setLoading] = useState(false);

  const handleCheckStatus = async (
    emailParam?: string,
    applicationIdParam?: string
  ) => {
    const emailToUse = emailParam || email;
    const applicationIdToUse = applicationIdParam || applicationId;

    if (!emailToUse || !applicationIdToUse) {
      toast.error("Please enter both email and application ID");
      return;
    }

    setLoading(true);
    try {
      const response = await candidatePortalAPI.checkStatus(
        emailToUse,
        applicationIdToUse
      );
      setApplication(response.data.application);
    } catch (error) {
      toast.error(
        errorMessage(error, "Failed to load application status. Please check your email and application ID.")
      );
      setApplication(null);
    } finally {
      setLoading(false);
    }
  };

  // The confirmation email links here with both values so the status loads immediately.
  useEffect(() => {
    const emailParam = searchParams?.get("email");
    const applicationIdParam = searchParams?.get("applicationId");
    if (emailParam) setEmail(emailParam);
    if (applicationIdParam) setApplicationId(applicationIdParam);
    if (emailParam && applicationIdParam) {
      handleCheckStatus(emailParam, applicationIdParam);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case "shortlisted":
        return "bg-green-100 text-green-800 border-green-300";
      case "rejected":
        return "bg-red-100 text-red-800 border-red-300";
      case "cv_viewed":
        return "bg-blue-100 text-blue-800 border-blue-300";
      case "under_review":
        return "bg-yellow-100 text-yellow-800 border-yellow-300";
      case "interview_scheduled":
        return "bg-purple-100 text-purple-800 border-purple-300";
      default:
        return "bg-gray-100 text-gray-800 border-gray-300";
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "shortlisted":
        return "✓";
      case "rejected":
        return "✗";
      case "cv_viewed":
        return "👁️";
      case "under_review":
        return "📋";
      case "interview_scheduled":
        return "📅";
      default:
        return "⏳";
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
          <h1 className="text-3xl font-bold mb-2">Application Status Portal</h1>
          <p className="text-gray-600 mb-6">
            Enter your email and application ID to check your application status
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div>
              <label className="block text-sm font-medium mb-2">Email</label>
              <input
                type="email"
                className="w-full px-4 py-2 border rounded-lg"
                placeholder="your.email@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">
                Application ID
              </label>
              <input
                type="text"
                className="w-full px-4 py-2 border rounded-lg"
                placeholder="Application ID"
                value={applicationId}
                onChange={(e) => setApplicationId(e.target.value)}
              />
            </div>
          </div>

          <button
            onClick={() => handleCheckStatus()}
            disabled={loading}
            className="w-full bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 disabled:bg-blue-300 font-semibold"
          >
            {loading ? "Loading..." : "Check Status"}
          </button>
        </div>

        {application && (
          <>
            {/* Status Card */}
            <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-2xl font-bold">Application Status</h2>
                <span
                  className={`px-4 py-2 rounded-full text-sm font-semibold border ${getStatusColor(
                    application.status
                  )}`}
                >
                  {getStatusIcon(application.status)}{" "}
                  {application.status_label || application.status}
                </span>
              </div>

              <div className="space-y-4">
                <div>
                  <h3 className="font-semibold text-lg mb-2">
                    {application.job.title}
                  </h3>
                  <p className="text-gray-600">
                    Applied on:{" "}
                    {new Date(application.applied_at).toLocaleDateString()}
                  </p>
                </div>

                {application.expected_response_days !== undefined &&
                  application.expected_response_days > 0 && (
                    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                      <p className="text-blue-800 font-semibold">
                        📅 Expected Response: You&apos;ll hear from us within{" "}
                        {application.expected_response_days} day
                        {application.expected_response_days !== 1 ? "s" : ""}
                      </p>
                      {application.expected_response_date && (
                        <p className="text-blue-600 text-sm mt-1">
                          By:{" "}
                          {new Date(
                            application.expected_response_date
                          ).toLocaleDateString()}
                        </p>
                      )}
                    </div>
                  )}

                {/* Status Timeline */}
                {application.status_history && (
                  <div className="mt-6">
                    <h3 className="font-semibold mb-4">Application Timeline</h3>
                    <div className="space-y-3">
                      {application.status_history.map((step, index) => (
                        <div key={index} className="flex items-start gap-3">
                          <div
                            className={`w-8 h-8 rounded-full flex items-center justify-center font-semibold ${
                              step.completed
                                ? "bg-green-500 text-white"
                                : "bg-gray-300 text-gray-600"
                            }`}
                          >
                            {step.completed ? "✓" : index + 1}
                          </div>
                          <div className="flex-1">
                            <p
                              className={`font-medium ${
                                step.completed
                                  ? "text-gray-900"
                                  : "text-gray-500"
                              }`}
                            >
                              {step.label}
                            </p>
                            <p className="text-sm text-gray-500">
                              {new Date(step.timestamp).toLocaleString()}
                            </p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

// useSearchParams needs a Suspense boundary so the rest of the page can prerender.
export default function ApplicationStatusPage() {
  return (
    <Suspense>
      <ApplicationStatus />
    </Suspense>
  );
}
