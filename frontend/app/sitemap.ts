import type { MetadataRoute } from "next";

const SITE_URL = process.env.NEXT_PUBLIC_FRONTEND_URL || "http://localhost:3000";
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

// Rebuilt hourly; open job pages plus each hiring company's careers page.
export const revalidate = 3600;

type Entry = { job_id: string; company_id: string; updated_at: string };

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const home: MetadataRoute.Sitemap = [{ url: SITE_URL, changeFrequency: "weekly", priority: 1 }];
  let jobs: Entry[] = [];
  try {
    const res = await fetch(`${API_BASE_URL}/public/sitemap`, { next: { revalidate } });
    if (res.ok) jobs = (await res.json()).jobs ?? [];
  } catch {
    // API unreachable (e.g. during a CI build): publish the static entries and retry next hour.
  }

  const careers = new Map<string, Date>();
  for (const job of jobs) {
    const updated = new Date(job.updated_at);
    const prev = careers.get(job.company_id);
    if (!prev || updated > prev) careers.set(job.company_id, updated);
  }

  return [
    ...home,
    ...[...careers].map(([companyId, lastModified]) => ({
      url: `${SITE_URL}/jobs/${companyId}`,
      lastModified,
      changeFrequency: "daily" as const,
      priority: 0.7,
    })),
    ...jobs.map((job) => ({
      url: `${SITE_URL}/apply/${job.job_id}`,
      lastModified: new Date(job.updated_at),
      changeFrequency: "daily" as const,
      priority: 0.8,
    })),
  ];
}
