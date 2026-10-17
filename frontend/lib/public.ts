import { cache } from "react";
import type { PublicJob } from "./api";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

// Server-side reads for public pages. Cached 60s: all SSR traffic shares one IP against the API rate limit.
async function getJSON<T>(path: string): Promise<T | null> {
  const res = await fetch(`${API_BASE_URL}${path}`, { next: { revalidate: 60 } });
  if (res.status === 404 || res.status === 400) return null;
  if (!res.ok) throw new Error(`GET ${path} failed with ${res.status}`);
  return res.json();
}

export const getCompanyJobs = cache((companyId: string) =>
  getJSON<{ jobs: PublicJob[]; company_name: string }>(`/jobs/public/${encodeURIComponent(companyId)}`)
);

export const getPublicJob = cache((jobId: string) =>
  getJSON<{
    job: PublicJob;
    company_id: string;
    company_name: string;
    accepting_applications: boolean;
  }>(`/public/jobs/${encodeURIComponent(jobId)}`)
);

// Plain-text summary for meta descriptions.
export function summarize(text: string, max = 160): string {
  const flat = text.replace(/\s+/g, " ").trim();
  return flat.length > max ? `${flat.slice(0, max - 1).trimEnd()}…` : flat;
}
