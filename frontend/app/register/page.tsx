"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth";
import { ApiError } from "@/lib/api";
import { Button, Input, Label } from "@/components/ui";

export default function RegisterPage() {
  const { register } = useAuth();
  const router = useRouter();
  const [form, setForm] = useState({
    full_name: "",
    email: "",
    password: "",
    phone: "",
    whatsapp_number: "",
  });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  function upd(k: string, v: string) {
    setForm((f) => ({ ...f, [k]: v }));
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await register({
        full_name: form.full_name,
        email: form.email,
        password: form.password,
        phone: form.phone || undefined,
        whatsapp_number: form.whatsapp_number || undefined,
      });
      router.push("/dashboard");
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) setError("البريد الإلكتروني مستخدم مسبقاً");
      else setError("تعذّر إنشاء الحساب، تحقق من البيانات");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="mx-auto max-w-md">
      <div className="card p-6">
        <h1 className="mb-4 text-xl font-extrabold text-slate-800">إنشاء حساب جديد</h1>
        <form onSubmit={onSubmit} className="space-y-4">
          <div>
            <Label>الاسم الكامل</Label>
            <Input value={form.full_name} onChange={(e) => upd("full_name", e.target.value)} required />
          </div>
          <div>
            <Label>البريد الإلكتروني</Label>
            <Input type="email" value={form.email} onChange={(e) => upd("email", e.target.value)} required />
          </div>
          <div>
            <Label>كلمة المرور (6 أحرف على الأقل)</Label>
            <Input
              type="password"
              value={form.password}
              onChange={(e) => upd("password", e.target.value)}
              minLength={6}
              required
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <Label>الهاتف</Label>
              <Input value={form.phone} onChange={(e) => upd("phone", e.target.value)} placeholder="968..." />
            </div>
            <div>
              <Label>واتساب</Label>
              <Input
                value={form.whatsapp_number}
                onChange={(e) => upd("whatsapp_number", e.target.value)}
                placeholder="968..."
              />
            </div>
          </div>
          {error && <p className="text-sm text-red-600">{error}</p>}
          <Button type="submit" className="w-full" disabled={loading}>
            {loading ? "جارٍ الإنشاء..." : "إنشاء الحساب"}
          </Button>
        </form>
        <p className="mt-4 text-center text-sm text-slate-500">
          لديك حساب بالفعل؟{" "}
          <Link href="/login" className="font-semibold text-brand-700">
            تسجيل الدخول
          </Link>
        </p>
      </div>
    </div>
  );
}
