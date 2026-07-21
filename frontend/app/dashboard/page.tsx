"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { ListingCard, CreateListingResponse } from "@/lib/types";
import { ListingCardView } from "@/components/ListingCard";
import { daysUntil } from "@/lib/format";
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
        <MyGrid items={mine} onChanged={loadAll} empty="لم تنشر أي إعلان بعد." />
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

// MyGrid renders the seller's own listings with per-card actions: pay for a
// draft awaiting payment, or renew an expired listing.
function MyGrid({ items, onChanged, empty }: { items: ListingCard[]; onChanged: () => void; empty: string }) {
  if (items.length === 0) return <div className="card p-10 text-center text-slate-500">{empty}</div>;
  return (
    <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
      {items.map((i) => (
        <div key={i.id} className="space-y-2">
          <ListingCardView item={i} showStatus />
          <MyCardActions item={i} onChanged={onChanged} />
        </div>
      ))}
    </div>
  );
}

function MyCardActions({ item, onChanged }: { item: ListingCard; onChanged: () => void }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const payable = item.status === "draft" || item.status === "expired";
  const left = item.status === "active" ? daysUntil(item.expires_at) : null;

  async function pay() {
    setBusy(true);
    setError("");
    try {
      const res = await api<CreateListingResponse>(`/listings/${item.id}/pay`, { method: "POST", auth: true });
      if (res.payment?.checkout_url) {
        window.location.href = res.payment.checkout_url;
        return;
      }
      onChanged(); // free (no fee owed) → moved straight to review
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "تعذّر بدء الدفع");
      setBusy(false);
    }
  }

  if (left !== null && left <= 7) {
    return (
      <p className="text-center text-xs text-amber-600">
        {left > 0 ? `ينتهي خلال ${left} يوم` : "منتهٍ"}
      </p>
    );
  }
  if (!payable) return null;

  return (
    <div className="space-y-1">
      <Button variant="outline" className="w-full" onClick={pay} disabled={busy}>
        {busy ? "..." : item.status === "draft" ? "أكمل الدفع" : "تجديد الإعلان"}
      </Button>
      {error && <p className="text-center text-xs text-red-600">{error}</p>}
    </div>
  );
}
