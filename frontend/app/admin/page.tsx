"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import {
  AdminUser,
  AdminUsersResponse,
  AppSettings,
  Category,
  CategoryField,
  ContactListResponse,
  ContactMessage,
  ListingCard,
  PresenceResponse,
  SearchResponse,
} from "@/lib/types";
import { formatPrice, statusLabel, statusColor } from "@/lib/format";
import { Button, Input, Select, Label, Spinner, Badge } from "@/components/ui";

type Tab = "listings" | "users" | "categories" | "messages" | "settings" | "presence";

export default function AdminPage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const [tab, setTab] = useState<Tab>("listings");

  useEffect(() => {
    if (authLoading) return;
    if (!user) router.replace("/login");
    else if (user.role !== "admin") router.replace("/dashboard");
  }, [authLoading, user, router]);

  if (authLoading || !user || user.role !== "admin") return <Spinner />;

  return (
    <div className="space-y-5">
      <h1 className="text-2xl font-extrabold text-slate-800">لوحة الإدارة</h1>

      <div className="flex gap-2 border-b border-slate-200">
        <TabBtn active={tab === "listings"} onClick={() => setTab("listings")}>
          الإعلانات
        </TabBtn>
        <TabBtn active={tab === "users"} onClick={() => setTab("users")}>
          المستخدمون
        </TabBtn>
        <TabBtn active={tab === "categories"} onClick={() => setTab("categories")}>
          الأقسام
        </TabBtn>
        <TabBtn active={tab === "messages"} onClick={() => setTab("messages")}>
          الرسائل
        </TabBtn>
        <TabBtn active={tab === "settings"} onClick={() => setTab("settings")}>
          التسعير
        </TabBtn>
        <TabBtn active={tab === "presence"} onClick={() => setTab("presence")}>
          المتواجدون الآن
        </TabBtn>
      </div>

      {tab === "listings" ? (
        <ListingsTab />
      ) : tab === "users" ? (
        <UsersTab meId={user.id} />
      ) : tab === "categories" ? (
        <CategoriesTab />
      ) : tab === "messages" ? (
        <MessagesTab />
      ) : tab === "settings" ? (
        <SettingsTab />
      ) : (
        <PresenceTab />
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

function ErrorBar({ msg }: { msg: string }) {
  if (!msg) return null;
  return <div className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">{msg}</div>;
}

function Pager({ page, total, size, onPage }: { page: number; total: number; size: number; onPage: (p: number) => void }) {
  const pages = Math.ceil(total / size);
  if (pages <= 1) return null;
  return (
    <div className="flex items-center justify-center gap-3 pt-2">
      <Button variant="outline" disabled={page <= 1} onClick={() => onPage(page - 1)}>
        السابق
      </Button>
      <span className="text-sm text-slate-500">
        صفحة {page} من {pages}
      </span>
      <Button variant="outline" disabled={page >= pages} onClick={() => onPage(page + 1)}>
        التالي
      </Button>
    </div>
  );
}

// ---- listings ----

const PAGE_SIZE = 20;

function ListingsTab() {
  const [items, setItems] = useState<ListingCard[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState("pending");
  const [keyword, setKeyword] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const qs = new URLSearchParams({ page: String(page), page_size: String(PAGE_SIZE) });
      if (status) qs.set("status", status);
      if (query) qs.set("keyword", query);
      const res = await api<SearchResponse>(`/admin/listings?${qs}`, { auth: true });
      setItems(res.items || []);
      setTotal(res.total);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "تعذّر تحميل الإعلانات");
    } finally {
      setLoading(false);
    }
  }, [page, status, query]);

  useEffect(() => {
    load();
  }, [load]);

  async function setListingStatus(id: number, next: string) {
    try {
      await api(`/admin/listings/${id}/status`, { method: "PATCH", auth: true, body: { status: next } });
      load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "تعذّر تحديث الحالة");
    }
  }

  async function remove(id: number, title: string) {
    if (!confirm(`حذف الإعلان "${title}" نهائياً من الموقع؟`)) return;
    try {
      await api(`/listings/${id}`, { method: "DELETE", auth: true });
      load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "تعذّر حذف الإعلان");
    }
  }

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-end gap-3">
        <div className="w-44">
          <Label>الحالة</Label>
          <Select
            value={status}
            onChange={(e) => {
              setStatus(e.target.value);
              setPage(1);
            }}
          >
            <option value="">كل الحالات</option>
            <option value="pending">قيد المراجعة</option>
            <option value="active">منشور</option>
            <option value="rejected">مرفوض</option>
            <option value="sold">مباع</option>
            <option value="expired">منتهي</option>
          </Select>
        </div>
        <form
          className="flex flex-1 items-end gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            setQuery(keyword);
            setPage(1);
          }}
        >
          <div className="flex-1">
            <Label>بحث</Label>
            <Input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder="عنوان أو وصف الإعلان" />
          </div>
          <Button type="submit" variant="outline">
            بحث
          </Button>
        </form>
      </div>

      <ErrorBar msg={error} />

      {loading ? (
        <Spinner />
      ) : items.length === 0 ? (
        <div className="card p-10 text-center text-slate-500">
          {status === "pending" && !query ? (
            <>
              <div className="text-lg font-semibold text-slate-600">لا توجد إعلانات بانتظار المراجعة 🎉</div>
              <p className="mt-1 text-sm">
                الإعلانات الجديدة تظهر هنا تلقائياً. اختر «كل الحالات» لتصفّح جميع الإعلانات.
              </p>
            </>
          ) : (
            "لا توجد إعلانات مطابقة."
          )}
        </div>
      ) : (
        <>
          <div className="text-sm text-slate-500">{total} إعلان</div>
          <div className="space-y-3">
            {items.map((i) => (
              <div key={i.id} className="card flex flex-wrap items-center justify-between gap-4 p-4">
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
                <div className="flex flex-wrap gap-2">
                  {i.status !== "active" && <Button onClick={() => setListingStatus(i.id, "active")}>موافقة</Button>}
                  {i.status !== "rejected" && (
                    <Button variant="outline" onClick={() => setListingStatus(i.id, "rejected")}>
                      رفض
                    </Button>
                  )}
                  <Link href={`/post?edit=${i.id}&from=/admin`}>
                    <Button variant="outline">تعديل</Button>
                  </Link>
                  <Button variant="danger" onClick={() => remove(i.id, i.title)}>
                    حذف
                  </Button>
                </div>
              </div>
            ))}
          </div>
          <Pager page={page} total={total} size={PAGE_SIZE} onPage={setPage} />
        </>
      )}
    </div>
  );
}

// ---- users ----

function UsersTab({ meId }: { meId: number }) {
  const [items, setItems] = useState<AdminUser[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [role, setRole] = useState("");
  const [keyword, setKeyword] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const qs = new URLSearchParams({ page: String(page), page_size: String(PAGE_SIZE) });
      if (role) qs.set("role", role);
      if (query) qs.set("keyword", query);
      const res = await api<AdminUsersResponse>(`/admin/users?${qs}`, { auth: true });
      setItems(res.items || []);
      setTotal(res.total);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "تعذّر تحميل المستخدمين");
    } finally {
      setLoading(false);
    }
  }, [page, role, query]);

  useEffect(() => {
    load();
  }, [load]);

  // The API refuses to demote or suspend the last remaining admin, and to let
  // an admin act on their own account; surface that as a readable message.
  function explain(e: unknown, fallback: string) {
    if (e instanceof ApiError && e.status === 403) {
      return "غير مسموح: لا يمكنك تعديل حسابك الخاص أو إزالة آخر مدير في النظام.";
    }
    return e instanceof ApiError ? e.message : fallback;
  }

  async function changeRole(u: AdminUser) {
    const next = u.role === "admin" ? "user" : "admin";
    const msg = next === "admin" ? `منح ${u.full_name} صلاحيات مدير كاملة؟` : `سحب صلاحيات المدير من ${u.full_name}؟`;
    if (!confirm(msg)) return;
    try {
      await api(`/admin/users/${u.id}/role`, { method: "PATCH", auth: true, body: { role: next } });
      load();
    } catch (e) {
      setError(explain(e, "تعذّر تغيير الصلاحية"));
    }
  }

  async function changeStatus(u: AdminUser) {
    const next = u.status === "active" ? "suspended" : "active";
    const msg = next === "suspended" ? `حظر ${u.full_name}؟ سيتم إنهاء جلساته فوراً.` : `إعادة تفعيل حساب ${u.full_name}؟`;
    if (!confirm(msg)) return;
    try {
      await api(`/admin/users/${u.id}/status`, { method: "PATCH", auth: true, body: { status: next } });
      load();
    } catch (e) {
      setError(explain(e, "تعذّر تغيير حالة الحساب"));
    }
  }

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-end gap-3">
        <div className="w-44">
          <Label>الصلاحية</Label>
          <Select
            value={role}
            onChange={(e) => {
              setRole(e.target.value);
              setPage(1);
            }}
          >
            <option value="">الكل</option>
            <option value="admin">مدير</option>
            <option value="user">مستخدم</option>
          </Select>
        </div>
        <form
          className="flex flex-1 items-end gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            setQuery(keyword);
            setPage(1);
          }}
        >
          <div className="flex-1">
            <Label>بحث</Label>
            <Input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder="الاسم أو البريد" />
          </div>
          <Button type="submit" variant="outline">
            بحث
          </Button>
        </form>
      </div>

      <ErrorBar msg={error} />

      {loading ? (
        <Spinner />
      ) : items.length === 0 ? (
        <div className="card p-10 text-center text-slate-500">لا يوجد مستخدمون مطابقون.</div>
      ) : (
        <>
          <div className="text-sm text-slate-500">{total} مستخدم</div>
          <div className="space-y-2">
            {items.map((u) => {
              const isMe = u.id === meId;
              return (
                <div key={u.id} className="card flex flex-wrap items-center justify-between gap-4 p-4">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-bold text-slate-800">{u.full_name}</span>
                      {u.role === "admin" && <Badge className="bg-brand-100 text-brand-700">مدير</Badge>}
                      {u.status === "suspended" && <Badge className="bg-red-100 text-red-700">محظور</Badge>}
                      {isMe && <Badge className="bg-slate-200 text-slate-600">أنت</Badge>}
                    </div>
                    <div className="text-xs text-slate-500">
                      {u.email} · {u.listings_count} إعلان
                    </div>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Button variant="outline" disabled={isMe} onClick={() => changeRole(u)}>
                      {u.role === "admin" ? "سحب الإدارة" : "ترقية إلى مدير"}
                    </Button>
                    <Button
                      variant={u.status === "active" ? "danger" : "primary"}
                      disabled={isMe}
                      onClick={() => changeStatus(u)}
                    >
                      {u.status === "active" ? "حظر" : "إلغاء الحظر"}
                    </Button>
                  </div>
                </div>
              );
            })}
          </div>
          <Pager page={page} total={total} size={PAGE_SIZE} onPage={setPage} />
        </>
      )}
    </div>
  );
}

// ---- categories ----

function CategoriesTab() {
  const [cats, setCats] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [expanded, setExpanded] = useState<number | null>(null);
  const [creating, setCreating] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      setCats(await api<Category[]>("/categories"));
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "تعذّر تحميل الأقسام");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  async function remove(c: Category) {
    if (!confirm(`حذف قسم "${c.name_ar}"؟`)) return;
    try {
      await api(`/admin/categories/${c.id}`, { method: "DELETE", auth: true });
      load();
    } catch (e) {
      setError(
        e instanceof ApiError && e.status === 409
          ? "لا يمكن حذف قسم يحتوي على إعلانات أو أقسام فرعية."
          : "تعذّر حذف القسم"
      );
    }
  }

  if (loading) return <Spinner />;

  return (
    <div className="space-y-3">
      <div className="flex justify-end">
        <Button onClick={() => setCreating((v) => !v)}>{creating ? "إلغاء" : "+ قسم جديد"}</Button>
      </div>

      <ErrorBar msg={error} />

      {creating && (
        <CategoryForm
          onDone={() => {
            setCreating(false);
            load();
          }}
          onError={setError}
        />
      )}

      <div className="space-y-2">
        {cats.map((c) => (
          <div key={c.id} className="card p-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <span className="font-bold text-slate-800">
                  {c.icon} {c.name_ar}
                </span>
                <div className="text-xs text-slate-500">{c.slug}</div>
              </div>
              <div className="flex gap-2">
                <Button variant="outline" onClick={() => setExpanded(expanded === c.id ? null : c.id)}>
                  {expanded === c.id ? "إخفاء الحقول" : "الحقول المخصصة"}
                </Button>
                <Button variant="danger" onClick={() => remove(c)}>
                  حذف
                </Button>
              </div>
            </div>
            {expanded === c.id && <FieldsPanel categoryId={c.id} onError={setError} />}
          </div>
        ))}
      </div>
    </div>
  );
}

function CategoryForm({ onDone, onError }: { onDone: () => void; onError: (m: string) => void }) {
  const [f, setF] = useState({ slug: "", name_ar: "", name_en: "", icon: "", display_order: "0" });
  const [saving, setSaving] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    try {
      await api("/admin/categories", {
        method: "POST",
        auth: true,
        body: {
          slug: f.slug,
          name_ar: f.name_ar,
          name_en: f.name_en || null,
          icon: f.icon || null,
          display_order: Number(f.display_order) || 0,
          is_active: true,
        },
      });
      onDone();
    } catch (e) {
      onError(e instanceof ApiError && e.status === 409 ? "المعرّف (slug) مستخدم بالفعل." : "تعذّر إنشاء القسم");
    } finally {
      setSaving(false);
    }
  }

  return (
    <form onSubmit={submit} className="card grid grid-cols-2 gap-3 p-4">
      <div>
        <Label>الاسم بالعربية *</Label>
        <Input value={f.name_ar} onChange={(e) => setF({ ...f, name_ar: e.target.value })} required />
      </div>
      <div>
        <Label>المعرّف (slug) *</Label>
        <Input value={f.slug} onChange={(e) => setF({ ...f, slug: e.target.value })} placeholder="cars" required />
      </div>
      <div>
        <Label>الاسم بالإنجليزية</Label>
        <Input value={f.name_en} onChange={(e) => setF({ ...f, name_en: e.target.value })} />
      </div>
      <div>
        <Label>الأيقونة</Label>
        <Input value={f.icon} onChange={(e) => setF({ ...f, icon: e.target.value })} placeholder="🚗" />
      </div>
      <div>
        <Label>الترتيب</Label>
        <Input type="number" value={f.display_order} onChange={(e) => setF({ ...f, display_order: e.target.value })} />
      </div>
      <div className="col-span-2">
        <Button type="submit" disabled={saving}>
          {saving ? "جارٍ الحفظ..." : "إنشاء القسم"}
        </Button>
      </div>
    </form>
  );
}

function FieldsPanel({ categoryId, onError }: { categoryId: number; onError: (m: string) => void }) {
  const [fields, setFields] = useState<CategoryField[]>([]);
  const [loading, setLoading] = useState(true);
  const [adding, setAdding] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      setFields(await api<CategoryField[]>(`/categories/${categoryId}/fields`));
    } catch {
      setFields([]);
    } finally {
      setLoading(false);
    }
  }, [categoryId]);

  useEffect(() => {
    load();
  }, [load]);

  if (loading) return <Spinner />;

  return (
    <div className="mt-4 space-y-3 rounded-xl bg-slate-50 p-4">
      {fields.length === 0 ? (
        <p className="text-sm text-slate-500">لا توجد حقول مخصصة لهذا القسم.</p>
      ) : (
        <ul className="space-y-1 text-sm text-slate-700">
          {fields.map((f) => (
            <li key={f.id} className="flex items-center gap-2">
              <span className="font-semibold">{f.label_ar}</span>
              <span className="text-xs text-slate-500">
                {f.field_key} · {f.field_type}
                {f.is_required ? " · مطلوب" : ""}
              </span>
            </li>
          ))}
        </ul>
      )}
      {adding ? (
        <FieldForm
          categoryId={categoryId}
          onDone={() => {
            setAdding(false);
            load();
          }}
          onError={onError}
        />
      ) : (
        <Button variant="outline" onClick={() => setAdding(true)}>
          + حقل جديد
        </Button>
      )}
    </div>
  );
}

function FieldForm({
  categoryId,
  onDone,
  onError,
}: {
  categoryId: number;
  onDone: () => void;
  onError: (m: string) => void;
}) {
  const [f, setF] = useState({
    field_key: "",
    label_ar: "",
    field_type: "text",
    unit: "",
    options: "",
    is_required: false,
    is_filterable: false,
  });
  const [saving, setSaving] = useState(false);
  const needsOptions = f.field_type === "select" || f.field_type === "multiselect";

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    try {
      // Options are entered as a comma-separated list; the API stores the
      // {value, label_ar} shape the listing form renders.
      const options = needsOptions
        ? f.options
            .split(",")
            .map((s) => s.trim())
            .filter(Boolean)
            .map((s) => ({ value: s, label_ar: s }))
        : null;

      await api(`/admin/categories/${categoryId}/fields`, {
        method: "POST",
        auth: true,
        body: {
          field_key: f.field_key,
          label_ar: f.label_ar,
          field_type: f.field_type,
          unit: f.unit || null,
          options,
          is_required: f.is_required,
          is_filterable: f.is_filterable,
          display_order: 0,
        },
      });
      onDone();
    } catch (err) {
      onError(err instanceof ApiError ? err.message : "تعذّر إضافة الحقل");
    } finally {
      setSaving(false);
    }
  }

  return (
    <form onSubmit={submit} className="grid grid-cols-2 gap-3 rounded-lg bg-white p-3">
      <div>
        <Label>التسمية *</Label>
        <Input value={f.label_ar} onChange={(e) => setF({ ...f, label_ar: e.target.value })} required />
      </div>
      <div>
        <Label>المفتاح (key) *</Label>
        <Input
          value={f.field_key}
          onChange={(e) => setF({ ...f, field_key: e.target.value })}
          placeholder="mileage"
          required
        />
      </div>
      <div>
        <Label>النوع</Label>
        <Select value={f.field_type} onChange={(e) => setF({ ...f, field_type: e.target.value })}>
          <option value="text">نص</option>
          <option value="textarea">نص طويل</option>
          <option value="number">رقم</option>
          <option value="select">قائمة</option>
          <option value="multiselect">قائمة متعددة</option>
          <option value="boolean">نعم/لا</option>
          <option value="date">تاريخ</option>
          <option value="url">رابط</option>
        </Select>
      </div>
      <div>
        <Label>الوحدة</Label>
        <Input value={f.unit} onChange={(e) => setF({ ...f, unit: e.target.value })} placeholder="كم" />
      </div>
      {needsOptions && (
        <div className="col-span-2">
          <Label>الخيارات (افصل بفاصلة) *</Label>
          <Input
            value={f.options}
            onChange={(e) => setF({ ...f, options: e.target.value })}
            placeholder="أوتوماتيك, عادي"
            required
          />
        </div>
      )}
      <label className="flex items-center gap-2 text-sm text-slate-700">
        <input type="checkbox" checked={f.is_required} onChange={(e) => setF({ ...f, is_required: e.target.checked })} />
        حقل مطلوب
      </label>
      <label className="flex items-center gap-2 text-sm text-slate-700">
        <input
          type="checkbox"
          checked={f.is_filterable}
          onChange={(e) => setF({ ...f, is_filterable: e.target.checked })}
        />
        قابل للتصفية
      </label>
      <div className="col-span-2">
        <Button type="submit" disabled={saving}>
          {saving ? "جارٍ الحفظ..." : "إضافة الحقل"}
        </Button>
      </div>
    </form>
  );
}

// ---- messages (راسلنا inbox) ----

function MessagesTab() {
  const [items, setItems] = useState<ContactMessage[]>([]);
  const [total, setTotal] = useState(0);
  const [unread, setUnread] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const qs = new URLSearchParams({ page: String(page), page_size: String(PAGE_SIZE) });
      const res = await api<ContactListResponse>(`/admin/contact-messages?${qs}`, { auth: true });
      setItems(res.items || []);
      setTotal(res.total);
      setUnread(res.unread);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "تعذّر تحميل الرسائل");
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => {
    load();
  }, [load]);

  async function markRead(id: number) {
    try {
      await api(`/admin/contact-messages/${id}/read`, { method: "PATCH", auth: true });
      load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "تعذّر التحديث");
    }
  }

  async function remove(id: number) {
    if (!confirm("حذف هذه الرسالة؟")) return;
    try {
      await api(`/admin/contact-messages/${id}`, { method: "DELETE", auth: true });
      load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "تعذّر الحذف");
    }
  }

  if (loading) return <Spinner />;

  return (
    <div className="space-y-3">
      <ErrorBar msg={error} />
      <div className="text-sm text-slate-500">
        {total} رسالة
        {unread > 0 && <span className="mr-2 text-amber-600">· {unread} غير مقروءة</span>}
      </div>

      {items.length === 0 ? (
        <div className="card p-10 text-center text-slate-500">لا توجد رسائل.</div>
      ) : (
        <div className="space-y-2">
          {items.map((m) => (
            <div key={m.id} className={`card p-4 ${m.is_read ? "" : "border-amber-200 bg-amber-50/40"}`}>
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-bold text-slate-800">{m.name}</span>
                    {!m.is_read && <Badge className="bg-amber-100 text-amber-700">جديدة</Badge>}
                  </div>
                  <div className="text-xs text-slate-500">
                    {m.phone && <span>{m.phone}</span>}
                    {m.phone && m.email && <span> · </span>}
                    {m.email && <span>{m.email}</span>}
                  </div>
                  <p className="mt-2 whitespace-pre-wrap text-sm text-slate-700">{m.message}</p>
                </div>
                <div className="flex shrink-0 gap-2">
                  {m.phone && (
                    <a href={`https://wa.me/${m.phone.replace(/[^0-9]/g, "")}`} target="_blank" rel="noreferrer">
                      <Button variant="whatsapp">واتساب</Button>
                    </a>
                  )}
                  {!m.is_read && (
                    <Button variant="outline" onClick={() => markRead(m.id)}>
                      تحديد كمقروءة
                    </Button>
                  )}
                  <Button variant="danger" onClick={() => remove(m.id)}>
                    حذف
                  </Button>
                </div>
              </div>
            </div>
          ))}
          <Pager page={page} total={total} size={PAGE_SIZE} onPage={setPage} />
        </div>
      )}
    </div>
  );
}

// ---- settings (fees + duration) ----

const SETTING_LABELS: { key: string; label: string; hint: string; suffix: string }[] = [
  { key: "listing_fee_tier2", label: "رسوم الإعلان الثاني", hint: "يُدفع عند نشر الإعلان الثاني", suffix: "ر.ع / شهر" },
  { key: "listing_fee_tier3_plus", label: "رسوم الإعلان الثالث فأكثر", hint: "لكل إعلان بعد الثاني", suffix: "ر.ع / شهر" },
  { key: "listing_duration_days", label: "مدة بقاء الإعلان", hint: "المدة قبل انتهاء الإعلان تلقائياً", suffix: "يوم" },
];

function SettingsTab() {
  const [form, setForm] = useState<AppSettings>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    api<AppSettings>("/admin/settings", { auth: true })
      .then(setForm)
      .catch((e) => setError(e instanceof ApiError ? e.message : "تعذّر تحميل الإعدادات"))
      .finally(() => setLoading(false));
  }, []);

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    setSaved(false);
    try {
      const body: AppSettings = {};
      for (const s of SETTING_LABELS) body[s.key] = String(form[s.key] ?? "");
      const res = await api<AppSettings>("/admin/settings", { method: "PATCH", auth: true, body });
      setForm(res);
      setSaved(true);
    } catch (e) {
      setError(e instanceof ApiError ? "قيمة غير صالحة، تأكد من إدخال أرقام صحيحة." : "تعذّر الحفظ");
    } finally {
      setSaving(false);
    }
  }

  if (loading) return <Spinner />;

  return (
    <form onSubmit={save} className="card max-w-lg space-y-4 p-6">
      <div>
        <h2 className="font-bold text-slate-800">رسوم نشر الإعلانات</h2>
        <p className="text-sm text-slate-500">الإعلان الأول لكل مستخدم مجاني. هذه القيم تُطبّق على الإعلانات التالية.</p>
      </div>
      <ErrorBar msg={error} />
      {SETTING_LABELS.map((s) => (
        <div key={s.key}>
          <Label>{s.label}</Label>
          <div className="flex items-center gap-2">
            <Input
              type="number"
              step={s.key === "listing_duration_days" ? "1" : "0.001"}
              min="0"
              value={form[s.key] ?? ""}
              onChange={(e) => setForm({ ...form, [s.key]: e.target.value })}
              required
            />
            <span className="whitespace-nowrap text-sm text-slate-500">{s.suffix}</span>
          </div>
          <p className="mt-1 text-xs text-slate-400">{s.hint}</p>
        </div>
      ))}
      <div className="flex items-center gap-3">
        <Button type="submit" disabled={saving}>
          {saving ? "جارٍ الحفظ..." : "حفظ الإعدادات"}
        </Button>
        {saved && <span className="text-sm text-emerald-600">✓ تم الحفظ</span>}
      </div>
    </form>
  );
}

// ---- presence (who's online now) ----

function pathLabel(p: string): string {
  if (!p || p === "/") return "الصفحة الرئيسية";
  if (p.startsWith("/listings/")) return `يتصفّح إعلاناً (${p.replace("/listings/", "#")})`;
  if (p.startsWith("/post")) return "يضيف/يعدّل إعلاناً";
  if (p.startsWith("/dashboard")) return "لوحته الشخصية";
  if (p.startsWith("/admin")) return "لوحة الإدارة";
  if (p.startsWith("/contact")) return "صفحة راسلنا";
  if (p.startsWith("/pay")) return "صفحة الدفع";
  if (p.startsWith("/login") || p.startsWith("/register")) return "تسجيل الدخول";
  return p;
}

function agoLabel(s: number): string {
  if (s < 5) return "الآن";
  if (s < 60) return `قبل ${s} ث`;
  return `قبل ${Math.floor(s / 60)} د`;
}

function PresenceTab() {
  const [data, setData] = useState<PresenceResponse | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    try {
      setData(await api<PresenceResponse>("/admin/presence", { auth: true }));
      setError("");
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "تعذّر تحميل المتواجدين");
    } finally {
      setLoading(false);
    }
  }, []);

  // Live: refresh every 7 seconds.
  useEffect(() => {
    load();
    const iv = setInterval(load, 7000);
    return () => clearInterval(iv);
  }, [load]);

  if (loading) return <Spinner />;

  return (
    <div className="space-y-4">
      <ErrorBar msg={error} />

      <div className="flex flex-wrap gap-3">
        <Stat label="المتواجدون الآن" value={data?.count ?? 0} accent="brand" />
        <Stat label="مستخدمون" value={data?.users ?? 0} accent="emerald" />
        <Stat label="زوّار" value={data?.guests ?? 0} accent="slate" />
      </div>

      {!data || data.visitors.length === 0 ? (
        <div className="card p-10 text-center text-slate-500">لا يوجد متواجدون في الموقع الآن.</div>
      ) : (
        <div className="card divide-y divide-slate-100">
          {data.visitors.map((v, i) => (
            <div key={i} className="flex items-center justify-between gap-3 p-3">
              <div className="flex items-center gap-2">
                <span className="relative flex h-2.5 w-2.5">
                  <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75" />
                  <span className="relative inline-flex h-2.5 w-2.5 rounded-full bg-emerald-500" />
                </span>
                <span className="font-semibold text-slate-800">{v.name}</span>
                {v.is_user ? (
                  <Badge className="bg-emerald-100 text-emerald-700">مسجّل</Badge>
                ) : (
                  <Badge className="bg-slate-100 text-slate-500">زائر</Badge>
                )}
              </div>
              <div className="flex items-center gap-3 text-sm text-slate-500">
                <span>{pathLabel(v.path)}</span>
                <span className="text-xs text-slate-400">{agoLabel(v.seconds_ago)}</span>
              </div>
            </div>
          ))}
        </div>
      )}
      <p className="text-xs text-slate-400">يتحدّث تلقائياً كل ٧ ثوانٍ · يُعتبر الزائر متواجداً لمدة دقيقة من آخر نشاط.</p>
    </div>
  );
}

function Stat({ label, value, accent }: { label: string; value: number; accent: "brand" | "emerald" | "slate" }) {
  const colors = {
    brand: "bg-brand-50 text-brand-700 border-brand-100",
    emerald: "bg-emerald-50 text-emerald-700 border-emerald-100",
    slate: "bg-slate-50 text-slate-600 border-slate-200",
  }[accent];
  return (
    <div className={`flex-1 rounded-xl border p-4 text-center ${colors}`}>
      <div className="text-3xl font-black">{value}</div>
      <div className="text-sm">{label}</div>
    </div>
  );
}
