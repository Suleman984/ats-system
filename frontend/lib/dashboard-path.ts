"use client";

import { usePathname } from "next/navigation";
import { useAuthStore } from "./store";

// Dashboard pages are shared by the admin app and the embed iframe. Links are written against
// /admin/dashboard and rewritten here for the embed, which must carry company_id on every URL.
export function useDashboardHref() {
  const pathname = usePathname();
  const companyId = useAuthStore((state) => state.user?.company_id);
  const embedded = pathname.startsWith("/embed");

  return (path: string) => {
    if (!embedded) return path;
    const target = path.replace(/^\/admin\/dashboard/, "/embed/dashboard");
    if (!companyId) return target;
    return `${target}${target.includes("?") ? "&" : "?"}company_id=${encodeURIComponent(companyId)}`;
  };
}

export function useIsOwner() {
  return useAuthStore((state) => state.user?.role === "owner");
}
