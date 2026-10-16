"use client";

import type { RelationshipTimelineItem } from "@/lib/api";

export default function TimelineModal({
  timeline,
  onClose,
}: {
  timeline: RelationshipTimelineItem[];
  onClose: () => void;
}) {
  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-3xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        <div className="p-6 border-b">
          <div className="flex justify-between items-center">
            <h2 className="text-2xl font-bold">Relationship Timeline</h2>
            <button
              onClick={onClose}
              className="text-gray-500 hover:text-gray-700 text-2xl"
            >
              ×
            </button>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto p-6">
          {timeline.length === 0 ? (
            <p className="text-gray-500 text-center py-8">
              No timeline events yet
            </p>
          ) : (
            <div className="space-y-4">
              {timeline.map((item, index) => (
                <div key={index} className="flex items-start gap-4">
                  <div className="text-2xl shrink-0">{item.icon}</div>
                  <div className="flex-1 border-l-2 border-gray-200 pl-4 pb-4">
                    <div className="flex justify-between items-start">
                      <div>
                        <p className="font-semibold">{item.title}</p>
                        <p className="text-sm text-gray-600 mt-1">
                          {item.description}
                        </p>
                        {item.author && (
                          <p className="text-xs text-gray-500 mt-1">
                            By: {item.author}
                          </p>
                        )}
                        {item.admin && (
                          <p className="text-xs text-gray-500 mt-1">
                            By: {item.admin}
                          </p>
                        )}
                      </div>
                      <p className="text-xs text-gray-500">
                        {new Date(item.timestamp).toLocaleString()}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
