"use client";

import { useEffect, useRef, useState } from "react";
import type { Message } from "@/lib/api";

// Shared by the recruiter modal and the candidate portal; `viewer` decides which side is "mine".
export default function MessageThread({
  messages,
  viewer,
  loading,
  onSend,
}: {
  messages: Message[];
  viewer: Message["sender_type"];
  loading: boolean;
  onSend: (text: string) => Promise<boolean>;
}) {
  const [draft, setDraft] = useState("");
  const [sending, setSending] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ block: "end" });
  }, [messages.length]);

  const handleSend = async () => {
    const text = draft.trim();
    if (!text) return;
    setSending(true);
    if (await onSend(text)) setDraft("");
    setSending(false);
  };

  return (
    <div className="flex flex-col min-h-0">
      <div className="flex-1 overflow-y-auto space-y-3 mb-4 max-h-[50vh]">
        {loading ? (
          <p className="text-gray-500 text-center py-8">Loading messages...</p>
        ) : messages.length === 0 ? (
          <p className="text-gray-500 text-center py-8">No messages yet</p>
        ) : (
          messages.map((msg, i) => {
            const mine = msg.sender_type === viewer;
            const lastMine = mine && !messages.slice(i + 1).some((m) => m.sender_type === viewer);
            return (
              <div
                key={msg.id}
                className={`flex ${mine ? "justify-end" : "justify-start"}`}
              >
                <div
                  className={`max-w-[80%] rounded-lg px-4 py-2 ${
                    mine ? "bg-blue-600 text-white" : "bg-gray-100 text-gray-900"
                  }`}
                >
                  <p className="whitespace-pre-wrap break-words">{msg.message}</p>
                  <p
                    className={`text-xs mt-1 ${
                      mine ? "text-blue-100" : "text-gray-500"
                    }`}
                  >
                    {msg.sender_type === "recruiter" ? "Recruiter" : "Candidate"} ·{" "}
                    {new Date(msg.created_at).toLocaleString()}
                    {lastMine && msg.is_read && " · Seen"}
                  </p>
                </div>
              </div>
            );
          })
        )}
        <div ref={bottomRef} />
      </div>

      <div className="border-t pt-4">
        <textarea
          className="w-full px-4 py-2 border rounded-lg resize-none mb-3"
          placeholder="Write a message..."
          rows={3}
          maxLength={5000}
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
        />
        <div className="flex justify-end">
          <button
            onClick={handleSend}
            disabled={sending || !draft.trim()}
            className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-300"
          >
            {sending ? "Sending..." : "Send"}
          </button>
        </div>
      </div>
    </div>
  );
}
