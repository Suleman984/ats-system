"use client";

import { useEffect, useRef } from "react";
import { API_BASE_URL, type Message } from "./api";

// Live chat over Server-Sent Events. Sending stays on the normal POST endpoints; these streams only
// push what happened. "reconnected" fires after a dropped stream comes back, so views can reload
// anything they missed while disconnected.
export type ChatEvent =
  | { type: "message"; application_id: string; message: Message }
  | { type: "read"; application_id: string; reader: "candidate" | "recruiter" }
  | { type: "reconnected" };

type Handler = (event: ChatEvent) => void;

function attach(source: EventSource, emit: Handler) {
  let opened = false;
  source.addEventListener("open", () => {
    if (opened) emit({ type: "reconnected" });
    opened = true;
  });
  for (const type of ["message", "read"] as const) {
    source.addEventListener(type, (e) => {
      try {
        emit(JSON.parse((e as MessageEvent).data));
      } catch {
        // Ignore malformed events; the next reload catches up.
      }
    });
  }
}

// Recruiters share one stream per tab (browsers allow only ~6 connections per host over HTTP/1.1).
const handlers = new Set<Handler>();
let recruiterSource: EventSource | null = null;

function subscribeRecruiter(handler: Handler) {
  handlers.add(handler);
  if (!recruiterSource) {
    recruiterSource = new EventSource(`${API_BASE_URL}/events`, { withCredentials: true });
    attach(recruiterSource, (event) => handlers.forEach((h) => h(event)));
  }
  return () => {
    handlers.delete(handler);
    if (handlers.size === 0 && recruiterSource) {
      recruiterSource.close();
      recruiterSource = null;
    }
  };
}

// Subscribes a dashboard component to its company's chat events for as long as it's mounted.
export function useChatEvents(handler: Handler) {
  const latest = useRef(handler);
  useEffect(() => {
    latest.current = handler;
  });
  useEffect(() => subscribeRecruiter((event) => latest.current(event)), []);
}

// Candidate stream for one application. The token comes from a POST so the candidate's email and
// application ID never appear in a URL; when it expires the stream is re-opened with a new one.
export function openCandidateStream(getToken: () => Promise<string>, handler: Handler) {
  let source: EventSource | null = null;
  let retry: number | undefined;
  let stopped = false;

  const connect = async () => {
    try {
      const token = await getToken();
      if (stopped) return;
      source = new EventSource(`${API_BASE_URL}/candidate/events?token=${encodeURIComponent(token)}`);
      attach(source, handler);
      source.addEventListener("error", () => {
        // CLOSED means the browser gave up (e.g. expired token); otherwise it retries by itself.
        if (source?.readyState === EventSource.CLOSED && !stopped) {
          source.close();
          retry = window.setTimeout(async () => {
            await connect();
            handler({ type: "reconnected" });
          }, 5000);
        }
      });
    } catch {
      if (!stopped) retry = window.setTimeout(connect, 15000);
    }
  };
  connect();

  return () => {
    stopped = true;
    window.clearTimeout(retry);
    source?.close();
  };
}
