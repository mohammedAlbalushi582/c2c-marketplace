import type { MetadataRoute } from "next";
import { SITE_URL } from "@/lib/server-api";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: "*",
      allow: "/",
      // Keep private/app pages out of the index.
      disallow: ["/admin", "/dashboard", "/login", "/register", "/post", "/pay/"],
    },
    sitemap: `${SITE_URL}/sitemap.xml`,
    host: SITE_URL,
  };
}
