"use client";

import type { CVAnalysisResult } from "@/lib/api";

// Older rows may hold the analysis as a JSON string; anything unparseable shows as unavailable.
export function parseAnalysis(raw: CVAnalysisResult | string | undefined): CVAnalysisResult | null {
  if (!raw) return null;
  if (typeof raw !== "string") return raw;
  try {
    return JSON.parse(raw) as CVAnalysisResult;
  } catch {
    return null;
  }
}

function List({ title, items, empty }: { title: string; items?: string[]; empty: string }) {
  return (
    <div>
      <h3 className="font-semibold text-sm mb-1">{title}</h3>
      {items && items.length > 0 ? (
        <ul className="list-disc list-inside text-sm text-gray-700 space-y-0.5">
          {items.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      ) : (
        <p className="text-sm text-gray-500">{empty}</p>
      )}
    </div>
  );
}

export default function AnalysisModal({
  candidateName,
  analysis,
  onClose,
}: {
  candidateName?: string;
  analysis: CVAnalysisResult | null;
  onClose: () => void;
}) {
  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        <div className="p-6 border-b">
          <div className="flex justify-between items-center">
            <h2 className="text-2xl font-bold">
              CV Analysis{candidateName ? ` · ${candidateName}` : ""}
            </h2>
            <button onClick={onClose} className="text-gray-500 hover:text-gray-700 text-2xl">
              ×
            </button>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto p-6">
          {!analysis ? (
            <p className="text-gray-500 text-center py-8">Analysis details are unavailable.</p>
          ) : (
            <div className="space-y-5">
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-center">
                {[
                  ["Overall", analysis.match_score],
                  ["Skills", analysis.skills_match],
                  ["Experience", analysis.experience_match],
                  ["Languages", analysis.language_match],
                ].map(([label, value]) => (
                  <div key={label} className="rounded-lg bg-gray-50 border p-3">
                    <p className="text-xs text-gray-500">{label}</p>
                    <p className="text-xl font-bold">{value ?? "—"}{value != null ? "%" : ""}</p>
                  </div>
                ))}
              </div>
              {analysis.match_reason && (
                <p className="text-sm text-gray-700">{analysis.match_reason}</p>
              )}
              <p className="text-sm">
                <span className="font-semibold">Experience detected:</span>{" "}
                {analysis.experience} year{analysis.experience === 1 ? "" : "s"}
                {analysis.education ? ` · ${analysis.education}` : ""}
              </p>
              <div className="grid sm:grid-cols-2 gap-5">
                <List title="Strengths" items={analysis.strengths} empty="None noted" />
                <List title="Missing skills" items={analysis.missing_skills} empty="None" />
                <List title="Skills found" items={analysis.skills} empty="None detected" />
                <List title="Languages" items={analysis.languages} empty="None detected" />
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
