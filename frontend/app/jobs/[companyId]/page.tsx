import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { formatDateOnly } from "@/lib/api";
import { getCompanyJobs, summarize } from "@/lib/public";

type Props = { params: Promise<{ companyId: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { companyId } = await params;
  const data = await getCompanyJobs(companyId).catch(() => null);
  if (!data) return { title: "Careers" };
  const title = `Careers at ${data.company_name}`;
  const description = `${data.jobs.length} open position${data.jobs.length === 1 ? "" : "s"} at ${data.company_name}. Browse roles and apply online.`;
  return {
    title,
    description,
    alternates: { canonical: `/jobs/${companyId}` },
    openGraph: { title, description, type: "website", url: `/jobs/${companyId}` },
  };
}

export default async function PublicJobsPage({ params }: Props) {
  const { companyId } = await params;
  let data;
  try {
    data = await getCompanyJobs(companyId);
  } catch {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-6">
        <p className="text-gray-600">
          We couldn&apos;t load open positions right now. Please try again in a moment.
        </p>
      </div>
    );
  }
  if (!data) notFound();
  const { jobs, company_name: companyName } = data;

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-blue-600 text-white py-16">
        <div className="max-w-4xl mx-auto px-6">
          <h1 className="text-4xl font-bold mb-4">Careers at {companyName}</h1>
          <p className="text-xl">Join our team and make an impact</p>
        </div>
      </div>
      <div className="max-w-4xl mx-auto px-6 py-12">
        {jobs.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-xl text-gray-500">
              No open positions at the moment
            </p>
            <p className="text-gray-400 mt-2">
              Check back later for new opportunities
            </p>
          </div>
        ) : (
          <div className="space-y-6">
            {jobs.map((job) => (
              <div
                key={job.id}
                className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition"
              >
                <div className="flex justify-between items-start mb-4">
                  <div>
                    <h2 className="text-2xl font-bold text-gray-900">
                      {job.title}
                    </h2>
                    <div className="flex gap-4 mt-2 text-sm text-gray-600">
                      {job.location && <span>📍 {job.location}</span>}
                      {job.job_type && <span>💼 {job.job_type}</span>}
                      {job.salary_range && <span>💰 {job.salary_range}</span>}
                    </div>
                  </div>
                  <span className="px-3 py-1 bg-green-100 text-green-800 rounded-full text-sm">
                    Open
                  </span>
                </div>
                <p className="text-gray-700 mb-4">{summarize(job.description, 200)}</p>
                <div className="flex justify-between items-center">
                  <p className="text-sm text-gray-500">
                    Deadline: {formatDateOnly(job.deadline)}
                  </p>
                  <Link
                    href={`/apply/${job.id}`}
                    className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                  >
                    View &amp; Apply
                  </Link>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
