"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api, uploadImage } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { Category, CategoryField, Location, ListingDetail } from "@/lib/types";
import { Button, Input, Textarea, Select, Label, Spinner } from "@/components/ui";

export default function PostPage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();

  const [categories, setCategories] = useState<Category[]>([]);
  const [locations, setLocations] = useState<Location[]>([]);
  const [fields, setFields] = useState<CategoryField[]>([]);

  const [categoryId, setCategoryId] = useState<number | "">("");
  const [governorateId, setGovernorateId] = useState<string>("");
  const [wilayatId, setWilayatId] = useState<string>("");
  const [form, setForm] = useState({
    title: "",
    description: "",
    price: "",
    price_type: "fixed",
    contact_phone: user?.phone || "",
    whatsapp_number: user?.whatsapp_number || "",
  });
  const [attrs, setAttrs] = useState<Record<number, unknown>>({});
  const [files, setFiles] = useState<File[]>([]);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!authLoading && !user) router.replace("/login");
  }, [authLoading, user, router]);

  useEffect(() => {
    api<Category[]>("/categories").then(setCategories).catch(() => {});
    api<Location[]>("/locations").then(setLocations).catch(() => {});
  }, []);

  useEffect(() => {
    if (categoryId === "") {
      setFields([]);
      return;
    }
    api<CategoryField[]>(`/categories/${categoryId}/fields`).then(setFields).catch(() => setFields([]));
    setAttrs({});
  }, [categoryId]);

  const governorates = locations.filter((l) => l.type === "governorate");
  const wilayats = locations.filter((l) => l.parent_id && String(l.parent_id) === governorateId);

  function upd(k: string, v: string) {
    setForm((f) => ({ ...f, [k]: v }));
  }
  function setAttr(id: number, v: unknown) {
    setAttrs((a) => ({ ...a, [id]: v }));
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (categoryId === "") {
      setError("اختر القسم");
      return;
    }
    setSubmitting(true);
    try {
      const attributes = fields
        .filter((f) => attrs[f.id] !== undefined && attrs[f.id] !== "" && attrs[f.id] !== null)
        .map((f) => ({ field_id: f.id, value: attrs[f.id] }));

      const location_id = wilayatId ? Number(wilayatId) : governorateId ? Number(governorateId) : null;

      const created = await api<ListingDetail>("/listings", {
        method: "POST",
        auth: true,
        body: {
          category_id: Number(categoryId),
          location_id,
          title: form.title,
          description: form.description,
          price: form.price ? Number(form.price) : null,
          price_type: form.price_type,
          contact_phone: form.contact_phone || null,
          whatsapp_number: form.whatsapp_number || null,
          attributes,
        },
      });

      for (let i = 0; i < files.length; i++) {
        await uploadImage(created.id, files[i], i === 0);
      }

      router.push("/dashboard?posted=1");
    } catch (err) {
      setError("تعذّر نشر الإعلان، تأكد من تعبئة الحقول المطلوبة");
    } finally {
      setSubmitting(false);
    }
  }

  if (authLoading || !user) return <Spinner />;

  return (
    <div className="mx-auto max-w-2xl">
      <h1 className="mb-4 text-2xl font-extrabold text-slate-800">أضف إعلاناً جديداً</h1>
      <form onSubmit={onSubmit} className="card space-y-5 p-6">
        <div>
          <Label>القسم *</Label>
          <Select value={categoryId} onChange={(e) => setCategoryId(e.target.value ? Number(e.target.value) : "")} required>
            <option value="">اختر القسم</option>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name_ar}
              </option>
            ))}
          </Select>
        </div>

        <div>
          <Label>العنوان *</Label>
          <Input value={form.title} onChange={(e) => upd("title", e.target.value)} required />
        </div>

        <div>
          <Label>الوصف *</Label>
          <Textarea rows={4} value={form.description} onChange={(e) => upd("description", e.target.value)} required />
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <Label>السعر (ر.ع)</Label>
            <Input type="number" value={form.price} onChange={(e) => upd("price", e.target.value)} />
          </div>
          <div>
            <Label>نوع السعر</Label>
            <Select value={form.price_type} onChange={(e) => upd("price_type", e.target.value)}>
              <option value="fixed">ثابت</option>
              <option value="negotiable">قابل للتفاوض</option>
              <option value="on_request">عند الطلب</option>
            </Select>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <Label>المحافظة</Label>
            <Select
              value={governorateId}
              onChange={(e) => {
                setGovernorateId(e.target.value);
                setWilayatId("");
              }}
            >
              <option value="">اختر</option>
              {governorates.map((g) => (
                <option key={g.id} value={g.id}>
                  {g.name_ar}
                </option>
              ))}
            </Select>
          </div>
          <div>
            <Label>الولاية</Label>
            <Select value={wilayatId} onChange={(e) => setWilayatId(e.target.value)} disabled={!wilayats.length}>
              <option value="">اختر</option>
              {wilayats.map((wl) => (
                <option key={wl.id} value={wl.id}>
                  {wl.name_ar}
                </option>
              ))}
            </Select>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <Label>هاتف التواصل</Label>
            <Input value={form.contact_phone} onChange={(e) => upd("contact_phone", e.target.value)} placeholder="968..." />
          </div>
          <div>
            <Label>رقم واتساب</Label>
            <Input value={form.whatsapp_number} onChange={(e) => upd("whatsapp_number", e.target.value)} placeholder="968..." />
          </div>
        </div>

        {/* Dynamic category fields */}
        {fields.length > 0 && (
          <div className="space-y-4 rounded-xl bg-slate-50 p-4">
            <h2 className="font-bold text-slate-700">تفاصيل القسم</h2>
            {fields.map((f) => (
              <DynamicField key={f.id} field={f} value={attrs[f.id]} onChange={(v) => setAttr(f.id, v)} />
            ))}
          </div>
        )}

        <div>
          <Label>الصور</Label>
          <input
            type="file"
            accept="image/*"
            multiple
            onChange={(e) => setFiles(Array.from(e.target.files || []))}
            className="block w-full text-sm text-slate-500 file:ml-3 file:rounded-lg file:border-0 file:bg-brand-50 file:px-4 file:py-2 file:text-brand-700"
          />
          {files.length > 0 && <p className="mt-1 text-xs text-slate-500">{files.length} صورة محددة</p>}
        </div>

        {error && <p className="text-sm text-red-600">{error}</p>}
        <p className="text-xs text-slate-500">سيُراجع إعلانك من قبل الإدارة قبل نشره.</p>
        <Button type="submit" className="w-full" disabled={submitting}>
          {submitting ? "جارٍ النشر..." : "نشر الإعلان"}
        </Button>
      </form>
    </div>
  );
}

function DynamicField({
  field,
  value,
  onChange,
}: {
  field: CategoryField;
  value: unknown;
  onChange: (v: unknown) => void;
}) {
  const label = (
    <Label>
      {field.label_ar}
      {field.unit ? ` (${field.unit})` : ""}
      {field.is_required ? " *" : ""}
    </Label>
  );

  switch (field.field_type) {
    case "number":
      return (
        <div>
          {label}
          <Input
            type="number"
            value={(value as string) ?? ""}
            onChange={(e) => onChange(e.target.value === "" ? undefined : Number(e.target.value))}
            required={field.is_required}
          />
        </div>
      );
    case "textarea":
      return (
        <div>
          {label}
          <Textarea value={(value as string) ?? ""} onChange={(e) => onChange(e.target.value)} required={field.is_required} />
        </div>
      );
    case "boolean":
      return (
        <label className="flex items-center gap-2">
          <input type="checkbox" checked={!!value} onChange={(e) => onChange(e.target.checked)} />
          <span className="text-sm font-semibold text-slate-700">{field.label_ar}</span>
        </label>
      );
    case "select":
      return (
        <div>
          {label}
          <Select value={(value as string) ?? ""} onChange={(e) => onChange(e.target.value || undefined)} required={field.is_required}>
            <option value="">اختر</option>
            {(field.options || []).map((o) => (
              <option key={o.value} value={o.value}>
                {o.label_ar}
              </option>
            ))}
          </Select>
        </div>
      );
    case "multiselect":
      return (
        <div>
          {label}
          <div className="flex flex-wrap gap-2">
            {(field.options || []).map((o) => {
              const arr = Array.isArray(value) ? (value as string[]) : [];
              const checked = arr.includes(o.value);
              return (
                <label
                  key={o.value}
                  className={`cursor-pointer rounded-full border px-3 py-1 text-sm ${
                    checked ? "border-brand-500 bg-brand-50 text-brand-700" : "border-slate-300 text-slate-600"
                  }`}
                >
                  <input
                    type="checkbox"
                    className="hidden"
                    checked={checked}
                    onChange={() => {
                      const next = checked ? arr.filter((x) => x !== o.value) : [...arr, o.value];
                      onChange(next.length ? next : undefined);
                    }}
                  />
                  {o.label_ar}
                </label>
              );
            })}
          </div>
        </div>
      );
    case "date":
      return (
        <div>
          {label}
          <Input type="date" value={(value as string) ?? ""} onChange={(e) => onChange(e.target.value || undefined)} required={field.is_required} />
        </div>
      );
    default: // text, url
      return (
        <div>
          {label}
          <Input
            type={field.field_type === "url" ? "url" : "text"}
            value={(value as string) ?? ""}
            onChange={(e) => onChange(e.target.value || undefined)}
            required={field.is_required}
          />
        </div>
      );
  }
}
