"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth";
import { Button, Input, Label } from "@/components/ui";

export default function LoginPage() {
  const { login } = useAuth();
  const router = useRouter();
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await login(identifier, password);
      router.push("/dashboard");
    } catch (err) {
      setError("بيانات الدخول غير صحيحة");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="mx-auto max-w-md">
      <div className="card p-6">
        <h1 className="mb-4 text-xl font-extrabold text-slate-800">تسجيل الدخول</h1>
        <form onSubmit={onSubmit} className="space-y-4">
          <div>
            <Label>البريد الإلكتروني أو اسم المستخدم</Label>
            <Input value={identifier} onChange={(e) => setIdentifier(e.target.value)} required />
          </div>
          <div>
            <Label>كلمة المرور</Label>
            <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
          </div>
          {error && <p className="text-sm text-red-600">{error}</p>}
          <Button type="submit" className="w-full" disabled={loading}>
            {loading ? "جارٍ الدخول..." : "دخول"}
          </Button>
        </form>
        <p className="mt-4 text-center text-sm text-slate-500">
          ليس لديك حساب؟{" "}
          <Link href="/register" className="font-semibold text-brand-700">
            إنشاء حساب
          </Link>
        </p>
      </div>
    </div>
  );
}
