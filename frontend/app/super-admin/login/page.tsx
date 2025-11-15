"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { superAdminAPI } from "@/lib/api";
import { useSuperAdminStore } from "@/lib/store";

export default function SuperAdminLoginPage() {
  const router = useRouter();
  const { login, isAuthenticated, checkAuth, initialized } =
    useSuperAdminStore();
  const [mounted, setMounted] = useState(false);
  const [formData, setFormData] = useState({
    email: "",
    password: "",
  });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setMounted(true);
    checkAuth();
  }, [checkAuth]);

  useEffect(() => {
    if (mounted && initialized && isAuthenticated) {
      router.push("/super-admin/dashboard");
    }
  }, [isAuthenticated, initialized, mounted, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const response = await superAdminAPI.login(formData);
      login(response.data.super_admin, response.data.token);
      router.push("/super-admin/dashboard");
    } catch (err: any) {
      setError(err.response?.data?.error || "Login failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full bg-white p-8 rounded-lg shadow-lg">
        <h2 className="text-3xl font-bold text-center mb-2">
          Super Admin Login
        </h2>
        <p className="text-center text-gray-600 mb-6">Platform Owner Access</p>
        {error && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
            {error}
          </div>
        )}
        <form onSubmit={handleSubmit} className="space-y-4">
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
              className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
              value={formData.password}
              onChange={(e) =>
                setFormData({ ...formData, password: e.target.value })
              }
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full bg-purple-600 text-white py-2 rounded-lg hover:bg-purple-700 disabled:bg-purple-300"
          >
            {loading ? "Logging in..." : "Login as Super Admin"}
          </button>
        </form>
        <p className="text-center mt-4 text-sm text-gray-500">
          Super admin accounts are managed by platform administrators
        </p>
      </div>
    </div>
  );
}
