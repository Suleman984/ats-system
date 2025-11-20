"use client";

import { useEffect, useState } from "react";
import { candidateSearchAPI, CandidateSearchResult } from "@/lib/api";
import { toast } from "@/components/Toast";

export default function FindCandidatesPage() {
  const [searchResults, setSearchResults] = useState<CandidateSearchResult[]>(
    []
  );
  const [loading, setLoading] = useState(false);
  const [searchCriteria, setSearchCriteria] = useState({
    query: "",
    skills: "",
    min_experience: "",
    max_experience: "",
    current_position: "",
    languages: "",
    has_portfolio: false,
    has_linkedin: false,
    status: "",
    limit: 50,
  });
  const [totalCandidates, setTotalCandidates] = useState(0);
  const [showDetails, setShowDetails] = useState<string | null>(null);
  const [candidateDetails, setCandidateDetails] = useState<any>(null);

  const handleSearch = async () => {
    setLoading(true);
    try {
      // Build search request
      const request: any = {
        limit: searchCriteria.limit || 50,
      };

      if (searchCriteria.query.trim()) {
        request.query = searchCriteria.query.trim();
      }

      if (searchCriteria.skills.trim()) {
        request.skills = searchCriteria.skills
          .split(",")
          .map((s) => s.trim())
          .filter((s) => s.length > 0);
      }

      if (searchCriteria.min_experience) {
        request.min_experience = parseInt(searchCriteria.min_experience);
      }

      if (searchCriteria.max_experience) {
        request.max_experience = parseInt(searchCriteria.max_experience);
      }

      if (searchCriteria.current_position.trim()) {
        request.current_position = searchCriteria.current_position.trim();
      }

      if (searchCriteria.languages.trim()) {
        request.languages = searchCriteria.languages
          .split(",")
          .map((l) => l.trim())
          .filter((l) => l.length > 0);
      }

      if (searchCriteria.has_portfolio) {
        request.has_portfolio = true;
      }

      if (searchCriteria.has_linkedin) {
        request.has_linkedin = true;
      }

      if (searchCriteria.status) {
        request.status = searchCriteria.status;
      }

      const response = await candidateSearchAPI.search(request);
      setSearchResults(response.data.candidates || []);
      setTotalCandidates(response.data.total || 0);
    } catch (error) {
      console.error("Failed to search candidates:", error);
      toast.error("Failed to search candidates. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const handleViewDetails = async (candidateId: string) => {
    try {
      const response = await candidateSearchAPI.getDetails(candidateId);
      setCandidateDetails(response.data);
      setShowDetails(candidateId);
    } catch (error) {
      console.error("Failed to fetch candidate details:", error);
      toast.error("Failed to load candidate details.");
    }
  };

  const getMatchScoreColor = (score: number) => {
    if (score >= 70) return "bg-green-100 text-green-800";
    if (score >= 40) return "bg-yellow-100 text-yellow-800";
    return "bg-gray-100 text-gray-800";
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">🔍 Find Candidates</h1>
        <p className="text-gray-600 mb-6">
          Search through all CVs in your database to find the perfect
          candidates.
        </p>

        {/* Search Form */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-lg font-semibold mb-4">Search Criteria</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* General Search */}
            <div className="md:col-span-2">
              <label className="block text-sm font-medium mb-1">
                General Search (Keywords, Skills, etc.)
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border rounded-lg"
                placeholder="e.g., React, Python, 5 years experience..."
                value={searchCriteria.query}
                onChange={(e) =>
                  setSearchCriteria({
                    ...searchCriteria,
                    query: e.target.value,
                  })
                }
              />
            </div>

            {/* Skills */}
            <div>
              <label className="block text-sm font-medium mb-1">
                Required Skills (comma-separated)
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border rounded-lg"
                placeholder="e.g., React, Node.js, TypeScript"
                value={searchCriteria.skills}
                onChange={(e) =>
                  setSearchCriteria({
                    ...searchCriteria,
                    skills: e.target.value,
                  })
                }
              />
            </div>

            {/* Languages */}
            <div>
              <label className="block text-sm font-medium mb-1">
                Languages (comma-separated)
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border rounded-lg"
                placeholder="e.g., English, Spanish, French"
                value={searchCriteria.languages}
                onChange={(e) =>
                  setSearchCriteria({
                    ...searchCriteria,
                    languages: e.target.value,
                  })
                }
              />
            </div>

            {/* Experience Range */}
            <div>
              <label className="block text-sm font-medium mb-1">
                Min Experience (years)
              </label>
              <input
                type="number"
                className="w-full px-3 py-2 border rounded-lg"
                placeholder="e.g., 3"
                value={searchCriteria.min_experience}
                onChange={(e) =>
                  setSearchCriteria({
                    ...searchCriteria,
                    min_experience: e.target.value,
                  })
                }
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">
                Max Experience (years)
              </label>
              <input
                type="number"
                className="w-full px-3 py-2 border rounded-lg"
                placeholder="e.g., 10"
                value={searchCriteria.max_experience}
                onChange={(e) =>
                  setSearchCriteria({
                    ...searchCriteria,
                    max_experience: e.target.value,
                  })
                }
              />
            </div>

            {/* Current Position */}
            <div>
              <label className="block text-sm font-medium mb-1">
                Current Position (keyword)
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border rounded-lg"
                placeholder="e.g., Developer, Manager"
                value={searchCriteria.current_position}
                onChange={(e) =>
                  setSearchCriteria({
                    ...searchCriteria,
                    current_position: e.target.value,
                  })
                }
              />
            </div>

            {/* Status */}
            <div>
              <label className="block text-sm font-medium mb-1">Status</label>
              <select
                className="w-full px-3 py-2 border rounded-lg"
                value={searchCriteria.status}
                onChange={(e) =>
                  setSearchCriteria({
                    ...searchCriteria,
                    status: e.target.value,
                  })
                }
              >
                <option value="">All Statuses</option>
                <option value="pending">Pending</option>
                <option value="shortlisted">Shortlisted</option>
                <option value="rejected">Rejected</option>
              </select>
            </div>

            {/* Filters */}
            <div className="md:col-span-2 flex gap-4">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  className="mr-2"
                  checked={searchCriteria.has_portfolio}
                  onChange={(e) =>
                    setSearchCriteria({
                      ...searchCriteria,
                      has_portfolio: e.target.checked,
                    })
                  }
                />
                Has Portfolio
              </label>
              <label className="flex items-center">
                <input
                  type="checkbox"
                  className="mr-2"
                  checked={searchCriteria.has_linkedin}
                  onChange={(e) =>
                    setSearchCriteria({
                      ...searchCriteria,
                      has_linkedin: e.target.checked,
                    })
                  }
                />
                Has LinkedIn
              </label>
            </div>
          </div>

          <button
            onClick={handleSearch}
            disabled={loading}
            className="mt-4 bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-300"
          >
            {loading ? "Searching..." : "🔍 Search Candidates"}
          </button>
        </div>

        {/* Results */}
        {searchResults.length > 0 && (
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-lg font-semibold">
                Search Results ({searchResults.length} of {totalCandidates})
              </h2>
            </div>

            <div className="space-y-4">
              {searchResults.map((result) => (
                <div
                  key={result.application.id}
                  className="border rounded-lg p-4 hover:shadow-md transition"
                >
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <h3 className="text-lg font-semibold">
                          {result.application.full_name}
                        </h3>
                        <span
                          className={`px-2 py-1 rounded-full text-xs font-semibold ${getMatchScoreColor(
                            result.match_score
                          )}`}
                        >
                          {result.match_score}% Match
                        </span>
                      </div>
                      <p className="text-sm text-gray-600 mb-1">
                        📧 {result.application.email}
                      </p>
                      {result.application.phone && (
                        <p className="text-sm text-gray-600 mb-1">
                          📞 {result.application.phone}
                        </p>
                      )}
                      {result.application.current_position && (
                        <p className="text-sm text-gray-600 mb-1">
                          💼 {result.application.current_position}
                        </p>
                      )}
                      {result.application.years_of_experience > 0 && (
                        <p className="text-sm text-gray-600 mb-1">
                          ⏱️ {result.application.years_of_experience} years of
                          experience
                        </p>
                      )}
                      {result.matched_skills.length > 0 && (
                        <div className="mt-2">
                          <p className="text-xs text-gray-500 mb-1">
                            Matched Skills:
                          </p>
                          <div className="flex flex-wrap gap-1">
                            {result.matched_skills.map((skill, idx) => (
                              <span
                                key={idx}
                                className="px-2 py-1 bg-blue-100 text-blue-800 rounded text-xs"
                              >
                                {skill}
                              </span>
                            ))}
                          </div>
                        </div>
                      )}
                      {result.matched_reasons.length > 0 && (
                        <div className="mt-2">
                          <p className="text-xs text-gray-500 mb-1">
                            Match Reasons:
                          </p>
                          <ul className="text-xs text-gray-600 list-disc list-inside">
                            {result.matched_reasons.map((reason, idx) => (
                              <li key={idx}>{reason}</li>
                            ))}
                          </ul>
                        </div>
                      )}
                      {result.application.job && (
                        <p className="text-xs text-gray-500 mt-2">
                          Applied for: {result.application.job.title || "N/A"}
                        </p>
                      )}
                    </div>
                    <div className="flex flex-col gap-2 ml-4">
                      <a
                        href={result.application.resume_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="px-3 py-1 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 text-center"
                      >
                        View CV
                      </a>
                      <button
                        onClick={() => handleViewDetails(result.application.id)}
                        className="px-3 py-1 bg-gray-600 text-white rounded text-sm hover:bg-gray-700"
                      >
                        Details
                      </button>
                    </div>
                  </div>

                  {/* Candidate Details Modal */}
                  {showDetails === result.application.id &&
                    candidateDetails && (
                      <div className="mt-4 p-4 bg-gray-50 rounded-lg border">
                        <h4 className="font-semibold mb-2">Full Details</h4>
                        <div className="space-y-2 text-sm">
                          {candidateDetails.skills.length > 0 && (
                            <div>
                              <strong>Skills:</strong>{" "}
                              {candidateDetails.skills.join(", ")}
                            </div>
                          )}
                          {candidateDetails.experience > 0 && (
                            <div>
                              <strong>Experience:</strong>{" "}
                              {candidateDetails.experience} years
                            </div>
                          )}
                          {candidateDetails.candidate.cover_letter && (
                            <div>
                              <strong>Cover Letter:</strong>
                              <p className="text-gray-600 mt-1">
                                {candidateDetails.candidate.cover_letter}
                              </p>
                            </div>
                          )}
                          {candidateDetails.candidate.linkedin_url && (
                            <div>
                              <strong>LinkedIn:</strong>{" "}
                              <a
                                href={candidateDetails.candidate.linkedin_url}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-blue-600 hover:underline"
                              >
                                {candidateDetails.candidate.linkedin_url}
                              </a>
                            </div>
                          )}
                          {candidateDetails.candidate.portfolio_url && (
                            <div>
                              <strong>Portfolio:</strong>{" "}
                              <a
                                href={candidateDetails.candidate.portfolio_url}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-blue-600 hover:underline"
                              >
                                {candidateDetails.candidate.portfolio_url}
                              </a>
                            </div>
                          )}
                          {candidateDetails.cv_text && (
                            <div className="mt-3">
                              <strong>CV Text Preview:</strong>
                              <div className="mt-1 p-3 bg-white rounded border max-h-40 overflow-y-auto text-xs">
                                {candidateDetails.cv_text.substring(0, 500)}
                                {candidateDetails.cv_text.length > 500 && "..."}
                              </div>
                            </div>
                          )}
                        </div>
                        <button
                          onClick={() => {
                            setShowDetails(null);
                            setCandidateDetails(null);
                          }}
                          className="mt-3 px-3 py-1 bg-gray-600 text-white rounded text-sm hover:bg-gray-700"
                        >
                          Close
                        </button>
                      </div>
                    )}
                </div>
              ))}
            </div>
          </div>
        )}

        {!loading && searchResults.length === 0 && (
          <div className="bg-white rounded-lg shadow p-6 text-center text-gray-500">
            No candidates found. Try adjusting your search criteria.
          </div>
        )}
      </div>
    </div>
  );
}
