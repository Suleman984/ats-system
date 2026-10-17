"use client";

import { useState } from "react";
import { authAPI, errorMessage } from "@/lib/api";
import { toast } from "@/components/Toast";

export default function ResendVerification({ email }: { email: string }) {
  const [sending, setSending] = useState(false);

  const resend = async () => {
    setSending(true);
    try {
      await authAPI.resendVerification(email);
      toast.success("If the account needs verifying, a new link is on its way.");
    } catch (error) {
      toast.error(errorMessage(error, "Couldn't resend the link"));
    } finally {
      setSending(false);
    }
  };

  return (
    <button
      type="button"
      onClick={resend}
      disabled={sending || !email}
      className="text-sm text-blue-600 hover:underline disabled:opacity-50"
    >
      {sending ? "Sending..." : "Resend verification email"}
    </button>
  );
}
