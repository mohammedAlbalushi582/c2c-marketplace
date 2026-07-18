import Link from "next/link";
import { ListingCard as Card } from "@/lib/types";
import { formatPrice, statusLabel, statusColor } from "@/lib/format";
import { Badge } from "@/components/ui";

export function ListingCardView({ item, showStatus = false }: { item: Card; showStatus?: boolean }) {
  return (
    <Link
      href={`/listings/${item.id}`}
      className="card group overflow-hidden transition hover:shadow-md"
    >
      <div className="relative aspect-[4/3] bg-slate-100">
        {item.primary_image ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={item.primary_image}
            alt={item.title}
            className="h-full w-full object-cover transition group-hover:scale-105"
          />
        ) : (
          <div className="grid h-full place-items-center text-slate-300">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              <rect x="3" y="3" width="18" height="18" rx="2" />
              <circle cx="8.5" cy="8.5" r="1.5" />
              <path d="m21 15-5-5L5 21" />
            </svg>
          </div>
        )}
        {item.is_featured && (
          <Badge className="absolute right-2 top-2 bg-amber-400 text-amber-900">مميّز</Badge>
        )}
        {showStatus && (
          <Badge className={`absolute left-2 top-2 ${statusColor(item.status)}`}>
            {statusLabel(item.status)}
          </Badge>
        )}
      </div>
      <div className="p-3">
        <h3 className="line-clamp-1 font-bold text-slate-800">{item.title}</h3>
        <div className="mt-1 text-brand-700 font-extrabold">
          {formatPrice(item.price, item.currency, item.price_type)}
        </div>
        <div className="mt-2 flex items-center justify-between text-xs text-slate-500">
          <span>{item.category_name_ar || ""}</span>
          {item.location_name_ar && (
            <span className="flex items-center gap-1">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z" />
                <circle cx="12" cy="10" r="3" />
              </svg>
              {item.location_name_ar}
            </span>
          )}
        </div>
      </div>
    </Link>
  );
}
