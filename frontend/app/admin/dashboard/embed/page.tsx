"use client";

import { useAuthStore } from "@/lib/store";
import { useState } from "react";

export default function EmbedPage() {
  const { user } = useAuthStore();
  const [copied, setCopied] = useState(false);

  const frontendUrl =
    process.env.NEXT_PUBLIC_FRONTEND_URL || "http://localhost:3000";
  const embedCode = `<iframe
  src="${frontendUrl}/jobs/${user?.company_id}"
  width="100%"
  height="800px"
  frameborder="0">
</iframe>`;

  const copyToClipboard = () => {
    navigator.clipboard.writeText(embedCode);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">Embed Job Portal</h1>
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-xl font-bold mb-4">How to Integrate</h2>
          <p className="text-gray-700 mb-4">
            Copy the code below and paste it into any page on your website where
            you want to display your job openings.
          </p>
          <div className="bg-gray-900 text-green-400 p-4 rounded-lg mb-4 relative">
            <pre className="text-sm overflow-x-auto">{embedCode}</pre>
            <button
              onClick={copyToClipboard}
              className="absolute top-2 right-2 px-3 py-1 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
            >
              {copied ? "Copied!" : "Copy"}
            </button>
          </div>
          <div className="space-y-4">
            <div>
              <h3 className="font-bold mb-2">For WordPress:</h3>
              <ol className="list-decimal list-inside space-y-1 text-sm text-gray-700">
                <li>Go to your WordPress admin panel</li>
                <li>Edit the page where you want to show jobs</li>
                <li>Add a &quot;Custom HTML&quot; block</li>
                <li>Paste the embed code</li>
                <li>Publish the page</li>
              </ol>
            </div>
            <div>
              <h3 className="font-bold mb-2">For HTML Website:</h3>
              <ol className="list-decimal list-inside space-y-1 text-sm text-gray-700">
                <li>Open your HTML file in a text editor</li>
                <li>Find where you want to display jobs</li>
                <li>Paste the embed code</li>
                <li>Save and upload the file</li>
              </ol>
            </div>
            <div>
              <h3 className="font-bold mb-2">For Shopify:</h3>
              <ol className="list-decimal list-inside space-y-1 text-sm text-gray-700">
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
            <code className="bg-white px-2 py-1 rounded mt-2 inline-block">
              {frontendUrl}/jobs/{user?.company_id}
            </code>
          </p>
        </div>
      </div>
    </div>
  );
}
