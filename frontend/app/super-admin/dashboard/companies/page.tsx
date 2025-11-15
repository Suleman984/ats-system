"use client";

import { useEffect, useState } from "react";
import { superAdminAPI, CompanyWithStats } from "@/lib/api";

export default function CompaniesPage() {
  const [companies, setCompanies] = useState<CompanyWithStats[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchCompanies();
  }, []);

  const fetchCompanies = async () => {
    try {
      const response = await superAdminAPI.getAllCompanies();
      setCompanies(response.data.companies);
    } catch (error) {
      console.error("Failed to fetch companies:", error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">All Companies</h1>

        <div className="bg-white rounded-lg shadow overflow-hidden">
          {companies.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              No companies registered yet
            </div>
          ) : (
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Company
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Email
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Subscription
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Jobs
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Applications
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Created
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {companies.map((company) => (
                  <tr key={company.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div>
                        <p className="font-semibold">{company.company_name}</p>
                        {company.company_website && (
                          <p className="text-sm text-gray-500">
                            {company.company_website}
                          </p>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm">{company.email}</td>
                    <td className="px-6 py-4">
                      <div>
                        <span
                          className={`px-2 py-1 rounded-full text-xs ${
                            company.subscription_status === "active"
                              ? "bg-green-100 text-green-800"
                              : company.subscription_status === "trial"
                              ? "bg-yellow-100 text-yellow-800"
                              : "bg-gray-100 text-gray-800"
                          }`}
                        >
                          {company.subscription_status}
                        </span>
                        <p className="text-xs text-gray-500 mt-1">
                          {company.subscription_tier}
                        </p>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm">{company.job_count}</td>
                    <td className="px-6 py-4 text-sm">
                      {company.application_count}
                    </td>
                    <td className="px-6 py-4 text-sm">
                      {new Date(company.created_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  );
}
