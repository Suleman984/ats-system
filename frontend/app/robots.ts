import type { MetadataRoute } from "next";

const SITE_URL = process.env.NEXT_PUBLIC_FRONTEND_URL || "http://localhost:3000";

// Public job pages are crawlable; dashboards, auth flows and the candidate portal are not.
export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: "*",
      allow: ["/", "/jobs/", "/apply/"],
      disallow: ["/admin/", "/super-admin/", "/embed/", "/application-status"],
    },
    sitemap: `${SITE_URL}/sitemap.xml`,
  };
}
