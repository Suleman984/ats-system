// Application pipeline, in order. Keep in sync with the applications_status_check constraint.
export const APPLICATION_STATUSES = [
  { value: "pending", label: "Pending", badge: "bg-yellow-100 text-yellow-800", icon: "⏳" },
  { value: "cv_viewed", label: "CV Viewed", badge: "bg-blue-100 text-blue-800", icon: "👁️" },
  { value: "shortlisted", label: "Shortlisted", badge: "bg-green-100 text-green-800", icon: "✓" },
  { value: "interview", label: "Interview", badge: "bg-indigo-100 text-indigo-800", icon: "🗓️" },
  { value: "offer", label: "Offer", badge: "bg-purple-100 text-purple-800", icon: "📨" },
  { value: "hired", label: "Hired", badge: "bg-emerald-100 text-emerald-800", icon: "🎉" },
  { value: "rejected", label: "Rejected", badge: "bg-red-100 text-red-800", icon: "✗" },
] as const;

const byValue = new Map<string, (typeof APPLICATION_STATUSES)[number]>(
  APPLICATION_STATUSES.map((s) => [s.value, s])
);

export const statusLabel = (status: string) => byValue.get(status)?.label ?? status;
export const statusBadge = (status: string) => byValue.get(status)?.badge ?? "bg-gray-100 text-gray-800";
export const statusIcon = (status: string) => byValue.get(status)?.icon ?? "⏳";

export const isUndecided = (status: string) => status === "pending" || status === "cv_viewed";

// The action that moves a candidate to the next stage, mirroring models.NextStage on the server.
export const NEXT_STAGE_ACTION: Record<string, string> = {
  shortlisted: "Move to Interview",
  interview: "Make Offer",
  offer: "Mark as Hired",
};
