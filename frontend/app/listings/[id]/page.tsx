"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { ListingDetail, ListingAttribute } from "@/lib/types";
import { formatPrice, statusLabel, statusColor } from "@/lib/format";
import { Button, Spinner, Badge } from "@/components/ui";

function renderAttrValue(a: ListingAttribute): string {
  if (a.value === null || a.value === undefined) return "-";
  if (a.field_type === "boolean") return a.value ? "نعم" : "لا";
  if (Array.isArray(a.value)) return a.value.join("، ");
  return String(a.value) + (a.unit ? ` ${a.unit}` : "");
}

export default function ListingDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const router = useRouter();
  const [listing, setListing] = useState<ListingDetail | null>(null);
  const [activeImg, setActiveImg] = useState(0);
  const [loading, setLoading] = useState(true);
  const [fav, setFav] = useState(false);
  const [modError, setModError] = useState("");

  useEffect(() => {
    api<ListingDetail>(`/listings/${id}`)
      .then(setListing)
      .finally(() => setLoading(false));
  }, [id]);

  async function toggleFav() {
    if (!user) {
      window.location.href = "/login";
      return;
    }
    try {
      if (fav) {
        await api(`/listings/${id}/favorite`, { method: "DELETE", auth: true });
      } else {
        await api(`/listings/${id}/favorite`, { method: "POST", auth: true });
      }
      setFav(!fav);
    } catch {
      /* ignore */
    }
  }

  async function setStatus(status: string) {
    setModError("");
    try {
      await api(`/admin/listings/${id}/status`, { method: "PATCH", auth: true, body: { status } });
      setListing((l) => (l ? { ...l, status } : l));
    } catch (e) {
      setModError(e instanceof ApiError ? e.message : "تعذّر تحديث الحالة");
    }
  }

  async function remove() {
    if (!confirm("حذف هذا الإعلان؟ لن يظهر بعد الآن في الموقع.")) return;
    setModError("");
    try {
      await api(`/listings/${id}`, { method: "DELETE", auth: true });
      router.push(isAdmin ? "/admin" : "/dashboard");
    } catch (e) {
      setModError(e instanceof ApiError ? e.message : "تعذّر حذف الإعلان");
    }
  }

  if (loading) return <Spinner />;
  if (!listing)
    return <div className="card p-12 text-center text-slate-500">لم يتم العثور على الإعلان</div>;

  const wa = listing.whatsapp_number?.replace(/[^0-9]/g, "");
  const phone = listing.contact_phone;
  const isAdmin = user?.role === "admin";
  const isOwner = user?.id === listing.user_id;
  const canManage = isAdmin || isOwner;

  return (
    <div className="grid gap-6 lg:grid-cols-[1fr_340px]">
      {/* Main */}
      <div className="space-y-4">
        <div className="card overflow-hidden">
          <div className="aspect-[16/10] bg-slate-100">
            {listing.images.length > 0 ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={listing.images[activeImg].url} alt={listing.title} className="h-full w-full object-cover" />
            ) : (
              <div className="grid h-full place-items-center text-slate-300">لا توجد صور</div>
            )}
          </div>
          {listing.images.length > 1 && (
            <div className="flex gap-2 overflow-x-auto p-3">
              {listing.images.map((img, i) => (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  key={img.id}
                  src={img.url}
                  alt=""
                  onClick={() => setActiveImg(i)}
                  className={`h-16 w-20 flex-shrink-0 cursor-pointer rounded-lg object-cover ${
                    i === activeImg ? "ring-2 ring-brand-500" : ""
                  }`}
                />
              ))}
            </div>
          )}
        </div>

        <div className="card p-5">
          <div className="flex items-start justify-between gap-3">
            <h1 className="text-2xl font-extrabold text-slate-800">{listing.title}</h1>
            <Badge className="bg-brand-100 text-brand-700">{listing.category_name_ar}</Badge>
          </div>
          {listing.location_name_ar && (
            <div className="mt-1 text-sm text-slate-500">📍 {listing.location_name_ar}</div>
          )}
          <div className="mt-3 text-2xl font-black text-brand-700">
            {formatPrice(listing.price, listing.currency, listing.price_type)}
          </div>
          <p className="mt-4 whitespace-pre-line leading-7 text-slate-700">{listing.description}</p>
        </div>

        {listing.attributes.length > 0 && (
          <div className="card p-5">
            <h2 className="mb-3 font-bold text-slate-800">التفاصيل</h2>
            <dl className="grid grid-cols-2 gap-3 sm:grid-cols-3">
              {listing.attributes.map((a) => (
                <div key={a.field_key} className="rounded-lg bg-slate-50 p-3">
                  <dt className="text-xs text-slate-500">{a.label_ar}</dt>
                  <dd className="font-semibold text-slate-800">{renderAttrValue(a)}</dd>
                </div>
              ))}
            </dl>
          </div>
        )}
      </div>

      {/* Contact sidebar */}
      <aside className="space-y-3">
        {/* Moderation / owner controls */}
        {canManage && (
          <div className="card space-y-3 p-5">
            <div className="flex items-center justify-between">
              <h2 className="font-bold text-slate-800">{isAdmin ? "إدارة الإعلان" : "إعلانك"}</h2>
              <Badge className={statusColor(listing.status)}>{statusLabel(listing.status)}</Badge>
            </div>

            {isAdmin && (
              <div className="flex gap-2">
                {listing.status !== "active" && (
                  <Button className="flex-1" onClick={() => setStatus("active")}>
                    موافقة
                  </Button>
                )}
                {listing.status !== "rejected" && (
                  <Button variant="outline" className="flex-1" onClick={() => setStatus("rejected")}>
                    رفض
                  </Button>
                )}
              </div>
            )}

            <div className="flex gap-2">
              <Link href={`/post?edit=${listing.id}&from=/listings/${listing.id}`} className="flex-1">
                <Button variant="outline" className="w-full">
                  تعديل
                </Button>
              </Link>
              <Button variant="danger" className="flex-1" onClick={remove}>
                حذف
              </Button>
            </div>

            {modError && <p className="text-sm text-red-600">{modError}</p>}
          </div>
        )}

        <div className="card space-y-3 p-5">
          <h2 className="font-bold text-slate-800">تواصل مع المعلن</h2>
          {wa && (
            <a href={`https://wa.me/${wa}`} target="_blank" rel="noreferrer" className="block">
              <Button variant="whatsapp" className="w-full">
                واتساب
              </Button>
            </a>
          )}
          {phone && (
            <a href={`tel:${phone}`} className="block">
              <Button variant="outline" className="w-full">
                اتصال: {phone}
              </Button>
            </a>
          )}
          {!wa && !phone && <p className="text-sm text-slate-500">لا تتوفر معلومات تواصل.</p>}
          <Button variant="ghost" onClick={toggleFav} className="w-full">
            {fav ? "★ في المفضلة" : "☆ أضف للمفضلة"}
          </Button>
        </div>
        <div className="card p-4 text-center text-xs text-slate-400">
          👁 {listing.views_count} مشاهدة
        </div>
      </aside>
    </div>
  );
}
