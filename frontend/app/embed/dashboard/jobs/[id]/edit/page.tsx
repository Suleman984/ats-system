import JobForm from "@/components/jobs/JobForm";

export default async function EditJobPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ company_id?: string }>;
}) {
  const { id } = await params;
  const { company_id } = await searchParams;
  const query = company_id ? `?company_id=${encodeURIComponent(company_id)}` : "";
  return <JobForm jobId={id} basePath={`/embed/dashboard/jobs${query}`} />;
}
