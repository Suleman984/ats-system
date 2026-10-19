"use client";

import { toast } from "@/components/Toast";

// Public pages live on the frontend's own domain, even when the dashboard runs inside an embed.
export const publicUrl = (path: string) =>
  `${process.env.NEXT_PUBLIC_FRONTEND_URL || window.location.origin}${path}`;

export default function CopyLinkButton({
  path,
  label,
  className,
}: {
  path: string;
  label: string;
  className?: string;
}) {
  const copy = () => {
    const url = publicUrl(path);
    navigator.clipboard
      .writeText(url)
      .then(() => toast.success("Link copied"))
      // Clipboard can be blocked (e.g. some iframes); show the link so it can be copied by hand.
      .catch(() => window.prompt("Copy this link:", url));
  };
  return (
    <button type="button" onClick={copy} className={className}>
      {label}
    </button>
  );
}
