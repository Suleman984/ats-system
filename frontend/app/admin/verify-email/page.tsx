"use client";

import { Suspense, useEffect, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { authAPI, errorMessage } from "@/lib/api";
import { useAuthStore } from "@/lib/store";

function VerifyEmail() {
  const router = useRouter();
  const token = useSearchParams().get("token") ?? "";
  const login = useAuthStore((state) => state.login);
  const [error, setError] = useState(token ? "" : "This link is missing its verification code.");
  // Tokens are single-use, so guard against React running the effect twice in development.
  const started = useRef(false);

  useEffect(() => {
    if (!token || started.current) return;
    started.current = true;
    authAPI
      .verifyEmail(token)
      .then(({ data }) => {
        login(data.admin);
        router.replace("/admin/dashboard");
      })
      .catch((err) => setError(errorMessage(err, "Couldn't verify your email")));
  }, [token, login, router]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full bg-white p-8 rounded-lg shadow-lg text-center">
        <h2 className="text-3xl font-bold mb-4">Verify email</h2>
        {error ? (
          <>
            <p className="text-red-700 mb-4">{error}</p>
            <a href="/admin/login" className="text-blue-600 hover:underline">
              Go to login
            </a>
          </>
        ) : (
          <p className="text-gray-600">Verifying your email…</p>
        )}
      </div>
    </div>
  );
}

// useSearchParams needs a Suspense boundary so the page can prerender.
export default function VerifyEmailPage() {
  return (
    <Suspense>
      <VerifyEmail />
    </Suspense>
  );
}
