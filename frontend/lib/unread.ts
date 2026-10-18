"use client";

import { useCallback, useEffect, useState } from "react";
import { messagesAPI } from "./api";
import { useChatEvents } from "./realtime";

const UNREAD_CHANGED = "ats:unread-changed";
// Live events drive updates; the slow poll only covers missed events.
const POLL_MS = 5 * 60_000;

// Tells every mounted bell/badge to refetch, e.g. right after a conversation is opened and read.
export function notifyUnreadChanged() {
  window.dispatchEvent(new Event(UNREAD_CHANGED));
}

// Unread candidate messages, total and per job. Refreshes on live chat events, on focus, and when a
// conversation is read, with a slow poll as a safety net.
export function useUnreadSummary() {
  const [summary, setSummary] = useState<{ total: number; byJob: Record<string, number> }>({
    total: 0,
    byJob: {},
  });

  const refresh = useCallback(() => {
    if (document.visibilityState !== "visible") return;
    messagesAPI
      .unreadSummary()
      .then(({ data }) => setSummary({ total: data.total, byJob: data.by_job ?? {} }))
      .catch(() => undefined);
  }, []);

  useChatEvents((event) => {
    if (event.type === "reconnected" || event.type === "read" || event.message.sender_type === "candidate") {
      refresh();
    }
  });

  useEffect(() => {
    refresh();
    const timer = window.setInterval(refresh, POLL_MS);
    window.addEventListener(UNREAD_CHANGED, refresh);
    document.addEventListener("visibilitychange", refresh);
    return () => {
      window.clearInterval(timer);
      window.removeEventListener(UNREAD_CHANGED, refresh);
      document.removeEventListener("visibilitychange", refresh);
    };
  }, [refresh]);

  return summary;
}
