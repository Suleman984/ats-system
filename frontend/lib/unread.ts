"use client";

import { useCallback, useEffect, useState } from "react";
import { messagesAPI } from "./api";

const UNREAD_CHANGED = "ats:unread-changed";
const POLL_MS = 60_000;

// Tells every mounted bell/badge to refetch, e.g. right after a conversation is opened and read.
export function notifyUnreadChanged() {
  window.dispatchEvent(new Event(UNREAD_CHANGED));
}

// Unread candidate messages, total and per job. Polls once a minute while the tab is visible and
// refreshes immediately on focus or when a conversation is read.
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
