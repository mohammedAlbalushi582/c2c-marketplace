"use client";

import { useEffect, useRef } from "react";
import { usePathname } from "next/navigation";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth";

// Emits a heartbeat (session id + current page) on navigation and every 20s so
// the admin's "who's online" panel stays current. Renders nothing.
function getSessionId(): string {
  const KEY = "amjad_session_id";
  let id = localStorage.getItem(KEY);
  if (!id) {
    id = (crypto.randomUUID?.() ?? String(Math.random()).slice(2)) as string;
    localStorage.setItem(KEY, id);
  }
  return id;
}

export function PresenceTracker() {
  const pathname = usePathname();
  const { user } = useAuth();
  const pathRef = useRef(pathname);
  const nameRef = useRef<string | undefined>(user?.full_name);
  pathRef.current = pathname;
  nameRef.current = user?.full_name;

  useEffect(() => {
    let stopped = false;
    async function ping() {
      try {
        await api("/presence/ping", {
          method: "POST",
          auth: true,
          body: { session_id: getSessionId(), path: pathRef.current, name: nameRef.current || "" },
        });
      } catch {
        /* heartbeat is best-effort */
      }
    }
    ping();
    const iv = setInterval(() => {
      if (!stopped) ping();
    }, 20000);
    return () => {
      stopped = true;
      clearInterval(iv);
    };
  }, [pathname, user?.id]);

  return null;
}
