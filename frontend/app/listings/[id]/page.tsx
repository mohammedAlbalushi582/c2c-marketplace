import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { serverFetch, SITE_URL } from "@/lib/server-api";
import { ListingDetail, ListingAttribute } from "@/lib/types";
import { formatPrice } from "@/lib/format";
import { Badge, Button } from "@/components/ui";
import { Gallery } from "@/components/listing/Gallery";
import { FavoriteButton } from "@/components/listing/FavoriteButton";
import { ManageCard } from "@/components/listing/ManageCard";

function getListing(id: string) {
  return serverFetch<ListingDetail>(`/listings/${id}`);
}

function renderAttrValue(a: ListingAttribute): string {
  if (a.value === null || a.value === undefined) return "-";
  if (a.field_type === "boolean") return a.value ? "نعم" : "لا";
  if (Array.isArray(a.value)) return a.value.join("، ");
  return String(a.value) + (a.unit ? ` ${a.unit}` : "");
}

// generateMetadata + the page fetch the same URL; Next dedupes it into one call.
export async function generateMetadata({ params }: { params: { id: string } }): Promise<Metadata> {
  const l = await getListing(params.id);
  if (!l) return { title: "إعلان غير موجود", robots: { index: false } };
  const description = l.description.replace(/\s+/g, " ").trim().slice(0, 160);
  const images = l.images[0]?.url ? [l.images[0].url] : [];
  return {
    title: l.title,
    description,
    alternates: { canonical: `/listings/${l.id}` },
    openGraph: {
      title: l.title,
      description,
      url: `${SITE_URL}/listings/${l.id}`,
      type: "article",
      images,
    },
    twitter: { card: "summary_large_image", title: l.title, description, images },
  };
}

export default async function ListingDetailPage({ params }: { params: { id: string } }) {
  const listing = await getListing(params.id);
  if (!listing) notFound();

  const wa = listing.whatsapp_number?.replace(/[^0-9]/g, "");
  const phone = listing.contact_phone;

  // Product structured data so Google can show a rich result.
  const jsonLd: Record<string, unknown> = {
    "@context": "https://schema.org",
    "@type": "Product",
    name: listing.title,
    description: listing.description,
    image: listing.images.map((i) => i.url),
    category: listing.category_name_ar,
    url: `${SITE_URL}/listings/${listing.id}`,
  };
  if (listing.price !== null && listing.price_type !== "on_request") {
    jsonLd.offers = {
      "@type": "Offer",
      price: listing.price,
      priceCurrency: listing.currency || "OMR",
      availability: listing.status === "active" ? "https://schema.org/InStock" : "https://schema.org/OutOfStock",
      url: `${SITE_URL}/listings/${listing.id}`,
    };
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[1fr_340px]">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />

      {/* Main */}
      <div className="space-y-4">
        <Gallery images={listing.images} title={listing.title} />

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
        <ManageCard id={listing.id} userId={listing.user_id} initialStatus={listing.status} />

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
          <FavoriteButton id={listing.id} />
        </div>

        <div className="card p-4 text-center text-xs text-slate-400">👁 {listing.views_count} مشاهدة</div>
      </aside>
    </div>
  );
}
