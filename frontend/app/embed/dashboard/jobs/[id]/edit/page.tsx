import JobForm from "@/components/jobs/JobForm";

export default async function EditJobPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <JobForm jobId={id} basePath="/embed/dashboard/jobs" />;
}
