"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { ListingCard } from "@/lib/types";
import { ListingCardView } from "@/components/ListingCard";
import { Button, Spinner } from "@/components/ui";

type Tab = "listings" | "favorites";

export default function DashboardPage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const [tab, setTab] = useState<Tab>("listings");
  const [mine, setMine] = useState<ListingCard[]>([]);
  const [favs, setFavs] = useState<ListingCard[]>([]);
  const [loading, setLoading] = useState(true);
  const [posted, setPosted] = useState(false);

  useEffect(() => {
    if (typeof window !== "undefined" && window.location.search.includes("posted=1")) setPosted(true);
  }, []);

  useEffect(() => {
    if (!authLoading && !user) router.replace("/login");
  }, [authLoading, user, router]);

  async function loadAll() {
    setLoading(true);
    try {
      const [m, f] = await Promise.all([
        api<ListingCard[]>("/me/listings", { auth: true }),
        api<ListingCard[]>("/me/favorites", { auth: true }),
      ]);
      setMine(m || []);
      setFavs(f || []);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    if (user) loadAll();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  if (authLoading || !user) return <Spinner />;

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-extrabold text-slate-800">مرحباً، {user.full_name}</h1>
        <div className="flex gap-2">
          {user.role === "admin" && (
            <Link href="/admin">
              <Button variant="outline">لوحة الإدارة</Button>
            </Link>
          )}
          <Link href="/post">
            <Button>+ أضف إعلان</Button>
          </Link>
        </div>
      </div>

      {posted && (
        <div className="rounded-lg border border-brand-200 bg-brand-50 p-3 text-sm text-brand-800">
          ✓ تم استلام إعلانك وهو الآن قيد المراجعة من قبل الإدارة.
        </div>
      )}

      <div className="flex gap-2 border-b border-slate-200">
        <TabBtn active={tab === "listings"} onClick={() => setTab("listings")}>
          إعلاناتي ({mine.length})
        </TabBtn>
        <TabBtn active={tab === "favorites"} onClick={() => setTab("favorites")}>
          المفضلة ({favs.length})
        </TabBtn>
      </div>

      {loading ? (
        <Spinner />
      ) : tab === "listings" ? (
        <Grid items={mine} showStatus empty="لم تنشر أي إعلان بعد." />
      ) : (
        <Grid items={favs} empty="لا توجد إعلانات في المفضلة." />
      )}
    </div>
  );
}

function TabBtn({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={`-mb-px border-b-2 px-4 py-2 text-sm font-semibold ${
        active ? "border-brand-600 text-brand-700" : "border-transparent text-slate-500 hover:text-slate-700"
      }`}
    >
      {children}
    </button>
  );
}

function Grid({ items, showStatus, empty }: { items: ListingCard[]; showStatus?: boolean; empty: string }) {
  if (items.length === 0)
    return <div className="card p-10 text-center text-slate-500">{empty}</div>;
  return (
    <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
      {items.map((i) => (
        <ListingCardView key={i.id} item={i} showStatus={showStatus} />
      ))}
    </div>
  );
}
