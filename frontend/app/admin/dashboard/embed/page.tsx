"use client";

import { useAuthStore } from "@/lib/store";
import { useState } from "react";

export default function EmbedPage() {
  const { user } = useAuthStore();
  const [copied, setCopied] = useState(false);
  const [copiedDashboard, setCopiedDashboard] = useState(false);

  const frontendUrl =
    process.env.NEXT_PUBLIC_FRONTEND_URL || "http://localhost:3000";

  // Job Portal Embed Code (existing)
  const jobPortalEmbedCode = `<iframe
  src="${frontendUrl}/jobs/${user?.company_id}"
  width="100%"
  height="800px"
  frameborder="0">
</iframe>`;

  // Full Dashboard Embed Code (new) - includes company_id for security
  const dashboardEmbedCode = `<iframe
  src="${frontendUrl}/embed/dashboard?company_id=${user?.company_id}"
  width="100%"
  height="900px"
  frameborder="0"
  allow="clipboard-read; clipboard-write">
</iframe>`;

  const copyToClipboard = (code: string, type: "job" | "dashboard") => {
    navigator.clipboard.writeText(code);
    if (type === "job") {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } else {
      setCopiedDashboard(true);
      setTimeout(() => setCopiedDashboard(false), 2000);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">Embed Codes</h1>

        {/* Full Dashboard Embed Section */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-xl font-bold mb-4">📊 Embed Full Dashboard</h2>
          <p className="text-gray-700 mb-4 text-sm">
            Embed the entire ATS dashboard into your website. This includes
            login, dashboard, jobs management, and applications - everything in
            one iframe.
          </p>
          <div className="bg-gray-900 text-green-400 p-4 rounded-lg mb-4 relative">
            <pre className="text-sm overflow-x-auto">{dashboardEmbedCode}</pre>
            <button
              onClick={() => copyToClipboard(dashboardEmbedCode, "dashboard")}
              className="absolute top-2 right-2 px-3 py-1 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
            >
              {copiedDashboard ? "Copied!" : "Copy"}
            </button>
          </div>
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3 mb-4">
            <p className="text-xs text-yellow-800">
              <strong>Security Note:</strong> This embed code is unique to your
              company. The company ID is included in the URL to ensure only
              authorized access. Users will need to log in within the embedded
              iframe. Make sure the iframe height is sufficient (recommended:
              900px or more).
            </p>
          </div>
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 mb-4">
            <p className="text-xs text-blue-800">
              <strong>🔒 Security:</strong> Each company has a unique embed
              code. If someone copies your embed code, they will only see the
              login page, and only users from your company can successfully log
              in. The system validates that the logged-in user belongs to the
              company specified in the embed URL.
            </p>
          </div>
          <div className="space-y-3">
            <div>
              <h3 className="font-bold mb-2 text-sm">For WordPress:</h3>
              <ol className="list-decimal list-inside space-y-1 text-xs text-gray-700">
                <li>Go to your WordPress admin panel</li>
                <li>Edit the page where you want to embed the dashboard</li>
                <li>Add a &quot;Custom HTML&quot; block</li>
                <li>Paste the embed code above</li>
                <li>Publish the page</li>
              </ol>
            </div>
            <div>
              <h3 className="font-bold mb-2 text-sm">For HTML Website:</h3>
              <ol className="list-decimal list-inside space-y-1 text-xs text-gray-700">
                <li>Open your HTML file in a text editor</li>
                <li>Find where you want to display the dashboard</li>
                <li>Paste the embed code</li>
                <li>Save and upload the file</li>
              </ol>
            </div>
          </div>
        </div>

        {/* Job Portal Embed Section (existing) */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-xl font-bold mb-4">💼 Embed Job Portal Only</h2>
          <p className="text-gray-700 mb-4 text-sm">
            Copy the code below and paste it into any page on your website where
            you want to display your job openings (public view only).
          </p>
          <div className="bg-gray-900 text-green-400 p-4 rounded-lg mb-4 relative">
            <pre className="text-sm overflow-x-auto">{jobPortalEmbedCode}</pre>
            <button
              onClick={() => copyToClipboard(jobPortalEmbedCode, "job")}
              className="absolute top-2 right-2 px-3 py-1 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
            >
              {copied ? "Copied!" : "Copy"}
            </button>
          </div>
          <div className="space-y-3">
            <div>
              <h3 className="font-bold mb-2 text-sm">For WordPress:</h3>
              <ol className="list-decimal list-inside space-y-1 text-xs text-gray-700">
                <li>Go to your WordPress admin panel</li>
                <li>Edit the page where you want to show jobs</li>
                <li>Add a &quot;Custom HTML&quot; block</li>
                <li>Paste the embed code</li>
                <li>Publish the page</li>
              </ol>
            </div>
            <div>
              <h3 className="font-bold mb-2 text-sm">For HTML Website:</h3>
              <ol className="list-decimal list-inside space-y-1 text-xs text-gray-700">
                <li>Open your HTML file in a text editor</li>
                <li>Find where you want to display jobs</li>
                <li>Paste the embed code</li>
                <li>Save and upload the file</li>
              </ol>
            </div>
            <div>
              <h3 className="font-bold mb-2 text-sm">For Shopify:</h3>
              <ol className="list-decimal list-inside space-y-1 text-xs text-gray-700">
                <li>Go to Online Store → Pages</li>
                <li>Create or edit a page (e.g., &quot;Careers&quot;)</li>
                <li>Click &quot;Show HTML&quot;</li>
                <li>Paste the embed code</li>
                <li>Save</li>
              </ol>
            </div>
          </div>
        </div>

        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
          <h3 className="font-bold text-blue-900 mb-2">💡 Pro Tip</h3>
          <p className="text-blue-800 text-sm">
            You can also share the direct link with candidates:
            <br />
            <code className="bg-white px-2 py-1 rounded mt-2 inline-block text-xs">
              {frontendUrl}/jobs/{user?.company_id}
            </code>
          </p>
        </div>
      </div>
    </div>
  );
}
