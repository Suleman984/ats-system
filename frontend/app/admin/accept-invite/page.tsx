"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { authAPI, errorMessage } from "@/lib/api";
import { useAuthStore } from "@/lib/store";

function AcceptInvite() {
  const router = useRouter();
  const token = useSearchParams().get("token") ?? "";
  const login = useAuthStore((state) => state.login);
  const [invite, setInvite] = useState<{ email: string; role: string; company_name: string } | null>(null);
  const [lookupError, setLookupError] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!token) return;
    authAPI
      .getInvite(token)
      .then(({ data }) => setInvite(data))
      .catch((err) => setLookupError(errorMessage(err, "This invite is invalid or has expired")));
  }, [token]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const { data } = await authAPI.acceptInvite(token, name, password);
      login(data.admin);
      router.push("/admin/dashboard");
    } catch (err) {
      setError(errorMessage(err, "Couldn't accept the invite"));
      setLoading(false);
    }
  };

  const problem = !token ? "This link is missing its invite code." : lookupError;

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full bg-white p-8 rounded-lg shadow-lg">
        <h2 className="text-3xl font-bold text-center mb-2">Join your team</h2>
        {problem ? (
          <p className="text-gray-600 text-center mt-4">{problem} Ask your team owner to send a new invite.</p>
        ) : !invite ? (
          <p className="text-gray-600 text-center mt-4">Loading invite...</p>
        ) : (
          <>
            <p className="text-gray-600 text-center mb-6">
              You&apos;re joining <strong>{invite.company_name}</strong> as a {invite.role}. Your login
              email is <strong>{invite.email}</strong>.
            </p>
            {error && (
              <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">{error}</div>
            )}
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-2">Your name</label>
                <input
                  required
                  maxLength={255}
                  className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">Choose a password</label>
                <input
                  type="password"
                  required
                  minLength={8}
                  maxLength={72}
                  autoComplete="new-password"
                  className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
                <p className="text-xs text-gray-500 mt-1">At least 8 characters</p>
              </div>
              <button
                type="submit"
                disabled={loading}
                className="w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-300"
              >
                {loading ? "Joining..." : "Join team"}
              </button>
            </form>
          </>
        )}
      </div>
    </div>
  );
}

// useSearchParams needs a Suspense boundary so the page can prerender.
export default function AcceptInvitePage() {
  return (
    <Suspense>
      <AcceptInvite />
    </Suspense>
  );
}
