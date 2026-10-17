"use client";

import { useEffect, useState } from "react";
import { applicationAPI, errorMessage, type Message } from "@/lib/api";
import { toast } from "@/components/Toast";
import MessageThread from "./MessageThread";

export default function MessagesModal({
  applicationId,
  candidateName,
  onClose,
}: {
  applicationId: string;
  candidateName?: string;
  onClose: () => void;
}) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    applicationAPI
      .getMessages(applicationId)
      .then((res) => setMessages(res.data.messages))
      .catch((error) => toast.error(errorMessage(error, "Failed to load messages")))
      .finally(() => setLoading(false));
  }, [applicationId]);

  const handleSend = async (text: string) => {
    try {
      const res = await applicationAPI.sendMessage(applicationId, text);
      setMessages((prev) => [...prev, res.data.data]);
      return true;
    } catch (error) {
      toast.error(errorMessage(error, "Failed to send message"));
      return false;
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        <div className="p-6 border-b">
          <div className="flex justify-between items-center">
            <h2 className="text-2xl font-bold">
              Messages{candidateName ? ` · ${candidateName}` : ""}
            </h2>
            <button
              onClick={onClose}
              className="text-gray-500 hover:text-gray-700 text-2xl"
            >
              ×
            </button>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto p-6">
          <MessageThread
            messages={messages}
            viewer="recruiter"
            loading={loading}
            onSend={handleSend}
          />
        </div>
      </div>
    </div>
  );
}
