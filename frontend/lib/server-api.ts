// Server-only API access for SSR pages, metadata, and the sitemap. Uses an
// internal URL (reachable from the Next.js server/container) which differs from
// the browser-facing NEXT_PUBLIC_API_URL — in Docker the frontend reaches the
// backend at http://backend:8080, not localhost.
const SERVER_API =
  process.env.API_INTERNAL_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

// Public site origin (your domain) — used for canonical URLs, Open Graph, and
// the sitemap. Set NEXT_PUBLIC_SITE_URL at build/deploy time.
export const SITE_URL = (process.env.NEXT_PUBLIC_SITE_URL || "http://localhost:3000").replace(/\/$/, "");

// serverFetch does a cached GET against the backend. Returns null on any error
// so callers can render a graceful fallback / notFound.
export async function serverFetch<T>(path: string, revalidateSeconds = 60): Promise<T | null> {
  try {
    const res = await fetch(`${SERVER_API}${path}`, { next: { revalidate: revalidateSeconds } });
    if (!res.ok) return null;
    return (await res.json()) as T;
  } catch {
    return null;
  }
}
