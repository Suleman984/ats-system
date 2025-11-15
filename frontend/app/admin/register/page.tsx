"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { authAPI } from "@/lib/api";
import { useAuthStore } from "@/lib/store";

export default function RegisterPage() {
  const router = useRouter();
  const login = useAuthStore((state) => state.login);
  const [formData, setFormData] = useState({
    company_name: "",
    email: "",
    password: "",
    name: "",
    embedded_mode: false,
    embed_domain: "",
  });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const response = await authAPI.register(formData);
      login(response.data.admin, response.data.token);
      router.push("/admin/dashboard");
    } catch (err: any) {
      setError(err.response?.data?.error || "Registration failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full bg-white p-8 rounded-lg shadow-lg">
        <h2 className="text-3xl font-bold text-center mb-6">
          Company Registration
        </h2>
        {error && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
            {error}
          </div>
        )}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">
              Company Name
            </label>
            <input
              type="text"
              required
              className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
              value={formData.company_name}
              onChange={(e) =>
                setFormData({ ...formData, company_name: e.target.value })
              }
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">Your Name</label>
            <input
              type="text"
              required
              className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
              value={formData.name}
              onChange={(e) =>
                setFormData({ ...formData, name: e.target.value })
              }
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">Email</label>
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
          <div>
            <label className="block text-sm font-medium mb-2">Password</label>
            <input
              type="password"
              required
              minLength={6}
              className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
              value={formData.password}
              onChange={(e) =>
                setFormData({ ...formData, password: e.target.value })
              }
            />
          </div>

          {/* Embedded Mode Option */}
          <div className="border-t pt-4 mt-4">
            <div className="flex items-center mb-4">
              <input
                type="checkbox"
                id="embedded_mode"
                className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                checked={formData.embedded_mode}
                onChange={(e) =>
                  setFormData({ ...formData, embedded_mode: e.target.checked })
                }
              />
              <label
                htmlFor="embedded_mode"
                className="ml-2 text-sm font-medium"
              >
                Use Embedded Dashboard Mode
              </label>
            </div>
            {formData.embedded_mode && (
              <div className="ml-6 mb-4">
                <label className="block text-sm font-medium mb-2">
                  Your Website Domain (Optional)
                  <span className="text-gray-500 text-xs ml-2">
                    (e.g., example.com - for security)
                  </span>
                </label>
                <input
                  type="text"
                  placeholder="example.com"
                  className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 text-sm"
                  value={formData.embed_domain}
                  onChange={(e) =>
                    setFormData({ ...formData, embed_domain: e.target.value })
                  }
                />
                <p className="text-xs text-gray-500 mt-1">
                  If provided, the dashboard can only be embedded on this domain
                  for security.
                </p>
              </div>
            )}
            {formData.embedded_mode && (
              <div className="ml-6 bg-blue-50 border border-blue-200 rounded-lg p-3">
                <p className="text-xs text-blue-800">
                  <strong>Note:</strong> In embedded mode, you'll get an embed
                  code to integrate the entire ATS dashboard into your website.
                  You can still access it normally by logging in, but you'll
                  also have the option to embed it.
                </p>
              </div>
            )}
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-300"
          >
            {loading ? "Registering..." : "Register"}
          </button>
        </form>
        <p className="text-center mt-4 text-sm">
          Already have an account?{" "}
          <a href="/admin/login" className="text-blue-600 hover:underline">
            Login here
          </a>
        </p>
      </div>
    </div>
  );
}
