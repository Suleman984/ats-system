import JobForm from "@/components/jobs/JobForm";

// The embed iframe needs company_id carried through every navigation.
export default async function CreateJobPage({
  searchParams,
}: {
  searchParams: Promise<{ company_id?: string }>;
}) {
  const { company_id } = await searchParams;
  const query = company_id ? `?company_id=${encodeURIComponent(company_id)}` : "";
  return <JobForm basePath={`/embed/dashboard/jobs${query}`} />;
}
