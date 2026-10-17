"use client";

import Link from "next/link";
import { useDashboardHref } from "@/lib/dashboard-path";
import { useUnreadSummary } from "@/lib/unread";

// Nav bell: unread candidate messages; links to the applications that have them.
export default function UnreadBell() {
  const href = useDashboardHref();
  const { total } = useUnreadSummary();
  const label = total ? `${total} unread candidate message${total === 1 ? "" : "s"}` : "No unread messages";

  return (
    <Link
      href={href("/admin/dashboard/applications?unread=1")}
      title={label}
      aria-label={label}
      className="relative inline-flex items-center justify-center w-9 h-9 rounded-full text-gray-600 hover:text-blue-600 hover:bg-gray-100"
    >
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M15 17h5l-1.4-1.4A2 2 0 0118 14.2V11a6 6 0 10-12 0v3.2a2 2 0 01-.6 1.4L4 17h5m6 0a3 3 0 11-6 0"
        />
      </svg>
      {total > 0 && (
        <span className="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] px-1 rounded-full bg-red-600 text-white text-[10px] font-bold flex items-center justify-center">
          {total > 99 ? "99+" : total}
        </span>
      )}
    </Link>
  );
}
