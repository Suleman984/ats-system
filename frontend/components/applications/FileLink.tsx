"use client";

import { errorMessage, openApplicationFile } from "@/lib/api";
import { toast } from "@/components/Toast";

// Opens a candidate's CV or portfolio: external links directly, uploaded files via a signed link.
export default function FileLink({
  applicationId,
  kind,
  url,
  hasFile,
  className,
  onOpen,
  children,
}: {
  applicationId: string;
  kind: "resume" | "portfolio";
  url?: string;
  hasFile?: boolean;
  className?: string;
  onOpen?: () => void;
  children: React.ReactNode;
}) {
  if (hasFile) {
    return (
      <button
        type="button"
        className={`text-left ${className ?? ""}`}
        onClick={() => {
          onOpen?.();
          openApplicationFile(applicationId, kind).catch((error) =>
            toast.error(errorMessage(error, "Couldn't open the file"))
          );
        }}
      >
        {children}
      </button>
    );
  }
  if (!url) return null;
  return (
    <a href={url} target="_blank" rel="noopener noreferrer" className={className} onClick={onOpen}>
      {children}
    </a>
  );
}
