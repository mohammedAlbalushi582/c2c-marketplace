"use client";

import { useEffect, useState, useCallback } from "react";
import { api } from "@/lib/api";
import { Category, Location, SearchResponse } from "@/lib/types";
import { ListingCardView } from "@/components/ListingCard";
import { Button, Input, Select, Spinner } from "@/components/ui";

export default function HomePage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [governorates, setGovernorates] = useState<Location[]>([]);
  const [results, setResults] = useState<SearchResponse | null>(null);
  const [loading, setLoading] = useState(true);

  // filters
  const [categoryId, setCategoryId] = useState<number | null>(null);
  const [keyword, setKeyword] = useState("");
  const [locationId, setLocationId] = useState<string>("");
  const [minPrice, setMinPrice] = useState("");
  const [maxPrice, setMaxPrice] = useState("");

  useEffect(() => {
    api<Category[]>("/categories").then(setCategories).catch(() => {});
    api<Location[]>("/locations")
      .then((locs) => setGovernorates(locs.filter((l) => l.type === "governorate")))
      .catch(() => {});
  }, []);

  const search = useCallback(async () => {
    setLoading(true);
    const params = new URLSearchParams();
    if (categoryId) params.set("category_id", String(categoryId));
    if (keyword) params.set("keyword", keyword);
    if (locationId) params.set("location_id", locationId);
    if (minPrice) params.set("min_price", minPrice);
    if (maxPrice) params.set("max_price", maxPrice);
    try {
      const res = await api<SearchResponse>(`/listings?${params.toString()}`);
      setResults(res);
    } finally {
      setLoading(false);
    }
  }, [categoryId, keyword, locationId, minPrice, maxPrice]);

  useEffect(() => {
    search();
  }, [categoryId, locationId, search]);

  return (
    <div className="space-y-6">
      {/* Hero */}
      <section className="rounded-2xl bg-gradient-to-l from-brand-700 to-brand-500 p-8 text-white">
        <h1 className="text-2xl font-extrabold sm:text-3xl">ابحث عن أرضك أو عقارك القادم</h1>
        <p className="mt-1 text-brand-50">أراضٍ ومنازل وقوالب مواقع — في مكان واحد</p>
        <div className="mt-5 flex flex-col gap-2 sm:flex-row">
          <Input
            placeholder="ابحث بالكلمات المفتاحية..."
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && search()}
            className="flex-1 text-slate-900"
          />
          <Button variant="outline" onClick={search} className="text-slate-800">
            بحث
          </Button>
        </div>
      </section>

      {/* Category chips */}
      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => setCategoryId(null)}
          className={`rounded-full px-4 py-1.5 text-sm font-semibold ${
            categoryId === null ? "bg-brand-600 text-white" : "bg-white text-slate-600 border border-slate-200"
          }`}
        >
          الكل
        </button>
        {categories.map((c) => (
          <button
            key={c.id}
            onClick={() => setCategoryId(c.id)}
            className={`rounded-full px-4 py-1.5 text-sm font-semibold ${
              categoryId === c.id ? "bg-brand-600 text-white" : "bg-white text-slate-600 border border-slate-200"
            }`}
          >
            {c.name_ar}
          </button>
        ))}
      </div>

      <div className="grid gap-6 md:grid-cols-[260px_1fr]">
        {/* Filters */}
        <aside className="card h-fit space-y-4 p-4">
          <h2 className="font-bold text-slate-800">تصفية النتائج</h2>
          <div>
            <label className="mb-1 block text-sm font-semibold text-slate-600">المحافظة</label>
            <Select value={locationId} onChange={(e) => setLocationId(e.target.value)}>
              <option value="">كل المحافظات</option>
              {governorates.map((g) => (
                <option key={g.id} value={g.id}>
                  {g.name_ar}
                </option>
              ))}
            </Select>
          </div>
          <div>
            <label className="mb-1 block text-sm font-semibold text-slate-600">نطاق السعر (ر.ع)</label>
            <div className="flex gap-2">
              <Input type="number" placeholder="من" value={minPrice} onChange={(e) => setMinPrice(e.target.value)} />
              <Input type="number" placeholder="إلى" value={maxPrice} onChange={(e) => setMaxPrice(e.target.value)} />
            </div>
          </div>
          <Button onClick={search} className="w-full">
            تطبيق
          </Button>
        </aside>

        {/* Results */}
        <section>
          {loading ? (
            <Spinner />
          ) : results && results.items.length > 0 ? (
            <>
              <div className="mb-3 text-sm text-slate-500">{results.total} إعلان</div>
              <div className="grid grid-cols-2 gap-4 lg:grid-cols-3">
                {results.items.map((item) => (
                  <ListingCardView key={item.id} item={item} />
                ))}
              </div>
            </>
          ) : (
            <div className="card grid place-items-center p-12 text-center text-slate-500">
              <p className="text-lg font-semibold">لا توجد إعلانات مطابقة</p>
              <p className="mt-1 text-sm">جرّب تعديل معايير البحث أو أضف أول إعلان.</p>
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
