import type { MetadataRoute } from "next";
import { SITE_URL, serverFetch } from "@/lib/server-api";
import { Category, SearchResponse } from "@/lib/types";

// Render at request time: at Docker build time the backend isn't reachable, so a
// statically-generated sitemap would miss every listing. The underlying data
// fetches are still cached (see serverFetch revalidate), so this stays cheap.
export const dynamic = "force-dynamic";

const MAX_PAGES = 20; // cap: up to 20×50 = 1000 listings

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const entries: MetadataRoute.Sitemap = [
    { url: `${SITE_URL}/`, changeFrequency: "daily", priority: 1 },
    { url: `${SITE_URL}/contact`, changeFrequency: "yearly", priority: 0.3 },
  ];

  // Category-filtered listing pages (the home page reads ?category_id).
  const cats = (await serverFetch<Category[]>("/categories", 3600)) || [];
  for (const c of cats) {
    entries.push({ url: `${SITE_URL}/?category_id=${c.id}`, changeFrequency: "daily", priority: 0.7 });
  }

  // Every active listing.
  const seen = new Set<number>();
  for (let page = 1; page <= MAX_PAGES; page++) {
    const res = await serverFetch<SearchResponse>(`/listings?page=${page}&page_size=50`, 3600);
    if (!res || !res.items || res.items.length === 0) break;
    for (const it of res.items) {
      if (seen.has(it.id)) continue;
      seen.add(it.id);
      entries.push({ url: `${SITE_URL}/listings/${it.id}`, changeFrequency: "weekly", priority: 0.8 });
    }
    if (res.items.length < 50) break;
  }

  return entries;
}
