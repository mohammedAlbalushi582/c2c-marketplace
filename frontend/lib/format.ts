export function formatPrice(price: number | null, currency: string, priceType: string): string {
  if (priceType === "on_request" || price === null) return "السعر عند الطلب";
  const n = new Intl.NumberFormat("ar-OM").format(price);
  const suffix = priceType === "negotiable" ? " (قابل للتفاوض)" : "";
  return `${n} ${currency === "OMR" ? "ر.ع" : currency}${suffix}`;
}

export function priceTypeLabel(t: string): string {
  return { fixed: "ثابت", negotiable: "قابل للتفاوض", on_request: "عند الطلب" }[t] || t;
}

export function statusLabel(s: string): string {
  return (
    {
      draft: "مسودة",
      pending: "قيد المراجعة",
      active: "منشور",
      sold: "مباع",
      expired: "منتهي",
      rejected: "مرفوض",
    }[s] || s
  );
}

export function statusColor(s: string): string {
  return (
    {
      active: "bg-emerald-100 text-emerald-700",
      pending: "bg-amber-100 text-amber-700",
      rejected: "bg-red-100 text-red-700",
      sold: "bg-slate-200 text-slate-600",
    }[s] || "bg-slate-100 text-slate-600"
  );
}
