import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { formatDateOnly } from "@/lib/api";
import { getPublicJob, summarize } from "@/lib/public";
import ApplyForm from "./ApplyForm";

type Props = { params: Promise<{ jobId: string }> };

const employmentTypes: Record<string, string> = {
  "full-time": "FULL_TIME",
  "part-time": "PART_TIME",
  contract: "CONTRACTOR",
  internship: "INTERN",
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { jobId } = await params;
  const data = await getPublicJob(jobId).catch(() => null);
  if (!data) return { title: "Job not found" };
  const title = `${data.job.title} at ${data.company_name}`;
  const description = summarize(data.job.description);
  return {
    title,
    description,
    alternates: { canonical: `/apply/${jobId}` },
    openGraph: { title, description, type: "website", url: `/apply/${jobId}` },
    // Closed roles stay reachable for existing links but shouldn't be indexed.
    robots: data.accepting_applications ? undefined : { index: false },
  };
}

export default async function ApplyPage({ params }: Props) {
  const { jobId } = await params;
  let data;
  try {
    data = await getPublicJob(jobId);
  } catch {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-6">
        <p className="text-gray-600">
          We couldn&apos;t load this job right now. Please try again in a moment.
        </p>
      </div>
    );
  }
  if (!data) notFound();
  const { job, company_id: companyId, company_name: companyName, accepting_applications: accepting } = data;
  const remote = /remote/i.test(job.location ?? "");

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "JobPosting",
    title: job.title,
    description: [job.description, job.requirements].filter(Boolean).join("\n\n"),
    datePosted: job.created_at,
    validThrough: `${job.deadline}T23:59:59`,
    employmentType: employmentTypes[job.job_type ?? ""] ?? undefined,
    hiringOrganization: { "@type": "Organization", name: companyName },
    ...(remote
      ? { jobLocationType: "TELECOMMUTE" }
      : job.location
        ? { jobLocation: { "@type": "Place", address: { "@type": "PostalAddress", addressLocality: job.location } } }
        : {}),
    directApply: true,
  };

  return (
    <div className="min-h-screen bg-gray-50 py-12">
      {accepting && (
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd).replace(/</g, "\\u003c") }}
        />
      )}
      <div className="max-w-3xl mx-auto px-6 space-y-6">
        <Link href={`/jobs/${companyId}`} className="text-sm text-blue-600 hover:text-blue-700">
          ← All jobs at {companyName}
        </Link>
        <div className="bg-white rounded-lg shadow-lg p-8">
          <p className="text-sm text-gray-500 mb-1">{companyName}</p>
          <h1 className="text-3xl font-bold mb-2">{job.title}</h1>
          <p className="text-gray-600 mb-6">
            {[job.location, job.job_type, job.salary_range, `Apply by ${formatDateOnly(job.deadline)}`]
              .filter(Boolean)
              .join(" · ")}
          </p>
          <h2 className="text-lg font-semibold mb-2">About the role</h2>
          <p className="text-gray-700 whitespace-pre-line mb-6">{job.description}</p>
          {job.requirements && (
            <>
              <h2 className="text-lg font-semibold mb-2">Requirements</h2>
              <p className="text-gray-700 whitespace-pre-line">{job.requirements}</p>
            </>
          )}
        </div>
        {accepting ? (
          <ApplyForm jobId={job.id} jobTitle={job.title} />
        ) : (
          <div className="rounded-lg border border-yellow-300 bg-yellow-50 p-4 text-yellow-800">
            This job is no longer accepting applications.
          </div>
        )}
      </div>
    </div>
  );
}
