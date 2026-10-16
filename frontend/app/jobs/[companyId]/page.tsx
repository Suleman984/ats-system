"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { jobAPI, Job, errorMessage } from "@/lib/api";
import { toast } from "@/components/Toast";
import Link from "next/link";

export default function PublicJobsPage() {
  const params = useParams();
  const companyId = params.companyId as string;
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchJobs = useCallback(async () => {
    try {
      const response = await jobAPI.getPublic(companyId);
      setJobs(response.data.jobs);
    } catch (error) {
      toast.error(errorMessage(error, "Failed to fetch jobs"));
    } finally {
      setLoading(false);
    }
  }, [companyId]);

  useEffect(() => {
    fetchJobs();
  }, [fetchJobs]);

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-blue-600 text-white py-16">
        <div className="max-w-4xl mx-auto px-6">
          <h1 className="text-4xl font-bold mb-4">Career Opportunities</h1>
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
                <p className="text-gray-700 mb-4 whitespace-pre-line">
                  {job.description.substring(0, 200)}...
                </p>
                <div className="flex justify-between items-center">
                  <p className="text-sm text-gray-500">
                    Deadline: {new Date(job.deadline).toLocaleDateString()}
                  </p>
                  <Link
                    href={`/apply/${job.id}`}
                    className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                  >
                    Apply Now
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
