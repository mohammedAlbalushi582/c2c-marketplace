"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { SearchResponse } from "@/lib/types";
import { Button } from "@/components/ui";

export function Header() {
  const { user, logout } = useAuth();
  const pathname = usePathname();
  const [pending, setPending] = useState<number | null>(null);

  // Keep the moderation badge current as the admin moves around the site.
  useEffect(() => {
    if (user?.role !== "admin") {
      setPending(null);
      return;
    }
    api<SearchResponse>("/admin/listings?status=pending&page_size=1", { auth: true })
      .then((r) => setPending(r.total))
      .catch(() => setPending(null));
  }, [user, pathname]);

  return (
    <header className="sticky top-0 z-20 border-b border-slate-200 bg-white/90 backdrop-blur">
      <div className="container flex h-16 items-center justify-between gap-4">
        <Link href="/" className="flex items-center gap-2">
          <span className="grid h-9 w-9 place-items-center rounded-lg bg-brand-600 text-lg font-black text-white">
            أ
          </span>
          <div className="leading-tight">
            <div className="text-base font-extrabold text-brand-800">الأمجاد</div>
            <div className="text-[11px] text-slate-500">للأعمال والعقارات</div>
          </div>
        </Link>

        <nav className="flex items-center gap-2">
          <Link
            href="/contact"
            title="راسلنا"
            className="flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm font-semibold text-slate-600 hover:bg-slate-100"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
              <rect x="3" y="5" width="18" height="14" rx="2" />
              <path d="m3 7 9 6 9-6" />
            </svg>
            <span className="hidden sm:inline">راسلنا</span>
          </Link>
          {user ? (
            <>
              {user.role === "admin" && (
                <Link href="/admin">
                  <Button variant="outline">
                    مراجعة الإعلانات
                    {pending !== null && pending > 0 && (
                      <span className="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-bold text-amber-700">
                        {pending}
                      </span>
                    )}
                  </Button>
                </Link>
              )}
              <Link href="/post">
                <Button>+ أضف إعلان</Button>
              </Link>
              <Link href="/dashboard">
                <Button variant="outline">لوحتي</Button>
              </Link>
              <Button variant="ghost" onClick={logout}>
                خروج
              </Button>
            </>
          ) : (
            <>
              <Link href="/post">
                <Button>+ أضف إعلان</Button>
              </Link>
              <Link href="/login">
                <Button variant="outline">دخول</Button>
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
