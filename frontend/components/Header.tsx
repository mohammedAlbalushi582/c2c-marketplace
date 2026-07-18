"use client";

import Link from "next/link";
import { useAuth } from "@/lib/auth";
import { Button } from "@/components/ui";

export function Header() {
  const { user, logout } = useAuth();

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
          {user ? (
            <>
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
