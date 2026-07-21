"use client";

import { useState } from "react";
import { api, ApiError } from "@/lib/api";
import { Button, Input, Textarea, Label } from "@/components/ui";

export default function ContactPage() {
  const [form, setForm] = useState({ name: "", phone: "", email: "", message: "" });
  const [sending, setSending] = useState(false);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");

  function upd(k: string, v: string) {
    setForm((f) => ({ ...f, [k]: v }));
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!form.name.trim() || !form.message.trim()) {
      setError("الاسم والرسالة مطلوبان");
      return;
    }
    setSending(true);
    try {
      await api("/contact", { method: "POST", body: form });
      setSent(true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "تعذّر إرسال الرسالة، حاول مرة أخرى");
    } finally {
      setSending(false);
    }
  }

  return (
    <div className="mx-auto max-w-xl">
      <h1 className="mb-1 text-2xl font-extrabold text-slate-800">راسلنا</h1>
      <p className="mb-4 text-sm text-slate-500">
        لديك سؤال أو اقتراح أو ترغب بإضافة قسم جديد؟ أرسل لنا رسالة وسيصلك ردّ إدارة الموقع.
      </p>

      {sent ? (
        <div className="card space-y-3 p-8 text-center">
          <div className="text-3xl">✓</div>
          <div className="text-lg font-bold text-slate-800">تم إرسال رسالتك</div>
          <p className="text-sm text-slate-500">شكراً لتواصلك معنا، سنقوم بمراجعة رسالتك في أقرب وقت.</p>
          <Button variant="outline" onClick={() => { setSent(false); setForm({ name: "", phone: "", email: "", message: "" }); }}>
            إرسال رسالة أخرى
          </Button>
        </div>
      ) : (
        <form onSubmit={onSubmit} className="card space-y-4 p-6">
          <div>
            <Label>الاسم *</Label>
            <Input value={form.name} onChange={(e) => upd("name", e.target.value)} required />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <Label>رقم الهاتف</Label>
              <Input value={form.phone} onChange={(e) => upd("phone", e.target.value)} placeholder="968..." />
            </div>
            <div>
              <Label>البريد الإلكتروني</Label>
              <Input type="email" value={form.email} onChange={(e) => upd("email", e.target.value)} />
            </div>
          </div>
          <div>
            <Label>الرسالة *</Label>
            <Textarea rows={5} value={form.message} onChange={(e) => upd("message", e.target.value)} required />
          </div>
          {error && <p className="text-sm text-red-600">{error}</p>}
          <Button type="submit" className="w-full" disabled={sending}>
            {sending ? "جارٍ الإرسال..." : "إرسال"}
          </Button>
        </form>
      )}
    </div>
  );
}
