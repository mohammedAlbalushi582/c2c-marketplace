"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { ListingCard } from "@/lib/types";
import { ListingCardView } from "@/components/ListingCard";
import { formatPrice, statusLabel, statusColor } from "@/lib/format";
import { Button, Spinner, Badge } from "@/components/ui";

type Tab = "listings" | "favorites" | "admin";

export default function DashboardPage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const [tab, setTab] = useState<Tab>("listings");
  const [mine, setMine] = useState<ListingCard[]>([]);
  const [favs, setFavs] = useState<ListingCard[]>([]);
  const [pending, setPending] = useState<ListingCard[]>([]);
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
      if (user?.role === "admin") {
        const p = await api<ListingCard[]>("/admin/listings?status=pending", { auth: true });
        setPending(p || []);
      }
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    if (user) loadAll();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  async function approve(id: number, status: "active" | "rejected") {
    await api(`/admin/listings/${id}/status`, { method: "PATCH", auth: true, body: { status } });
    setPending((p) => p.filter((x) => x.id !== id));
  }

  if (authLoading || !user) return <Spinner />;

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-extrabold text-slate-800">مرحباً، {user.full_name}</h1>
        <Link href="/post">
          <Button>+ أضف إعلان</Button>
        </Link>
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
        {user.role === "admin" && (
          <TabBtn active={tab === "admin"} onClick={() => setTab("admin")}>
            مراجعة الإعلانات ({pending.length})
          </TabBtn>
        )}
      </div>

      {loading ? (
        <Spinner />
      ) : tab === "listings" ? (
        <Grid items={mine} showStatus empty="لم تنشر أي إعلان بعد." />
      ) : tab === "favorites" ? (
        <Grid items={favs} empty="لا توجد إعلانات في المفضلة." />
      ) : (
        <AdminPanel items={pending} onApprove={approve} />
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

function AdminPanel({
  items,
  onApprove,
}: {
  items: ListingCard[];
  onApprove: (id: number, status: "active" | "rejected") => void;
}) {
  if (items.length === 0)
    return <div className="card p-10 text-center text-slate-500">لا توجد إعلانات قيد المراجعة 🎉</div>;
  return (
    <div className="space-y-3">
      {items.map((i) => (
        <div key={i.id} className="card flex items-center justify-between gap-4 p-4">
          <div className="flex items-center gap-3">
            <div className="h-14 w-16 overflow-hidden rounded-lg bg-slate-100">
              {i.primary_image && (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={i.primary_image} alt="" className="h-full w-full object-cover" />
              )}
            </div>
            <div>
              <Link href={`/listings/${i.id}`} className="font-bold text-slate-800 hover:text-brand-700">
                {i.title}
              </Link>
              <div className="text-xs text-slate-500">
                {i.category_name_ar} · {formatPrice(i.price, i.currency, i.price_type)}
              </div>
              <Badge className={`mt-1 ${statusColor(i.status)}`}>{statusLabel(i.status)}</Badge>
            </div>
          </div>
          <div className="flex gap-2">
            <Button onClick={() => onApprove(i.id, "active")}>موافقة</Button>
            <Button variant="danger" onClick={() => onApprove(i.id, "rejected")}>
              رفض
            </Button>
          </div>
        </div>
      ))}
    </div>
  );
}
