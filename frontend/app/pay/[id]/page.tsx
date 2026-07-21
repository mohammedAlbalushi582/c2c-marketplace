"use client";

import { Suspense, useCallback, useEffect, useState } from "react";
import { useParams, useSearchParams, useRouter } from "next/navigation";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { PaymentInfo } from "@/lib/types";
import { formatOMR } from "@/lib/format";
import { Button, Spinner } from "@/components/ui";

// Single page for both flows: the dev stub redirects here to "pay" with a
// button, and a real gateway (Thawani) redirects back here (?status=success)
// after its own hosted checkout — in which case we auto-confirm.
export default function PayPage() {
  return (
    <Suspense fallback={<Spinner />}>
      <Pay />
    </Suspense>
  );
}

type View = "loading" | "confirm" | "verifying" | "paid" | "canceled" | "error";

function Pay() {
  const { id } = useParams<{ id: string }>();
  const params = useSearchParams();
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();

  const [payment, setPayment] = useState<PaymentInfo | null>(null);
  const [view, setView] = useState<View>("loading");
  const [error, setError] = useState("");

  const returnedStatus = params.get("status"); // set by a real gateway redirect

  const verify = useCallback(async () => {
    setView("verifying");
    setError("");
    try {
      const res = await api<{ paid: boolean; status: string }>(`/payments/${id}/verify`, {
        method: "POST",
        auth: true,
      });
      setView(res.paid ? "paid" : "confirm");
      if (!res.paid) setError("لم يتم تأكيد الدفع بعد. إذا كنت قد أكملت الدفع فانتظر لحظات وأعد المحاولة.");
    } catch (err) {
      setView("error");
      setError(err instanceof ApiError ? err.message : "تعذّر تأكيد الدفع");
    }
  }, [id]);

  useEffect(() => {
    if (authLoading) return;
    if (!user) {
      router.replace("/login");
      return;
    }
    api<PaymentInfo>(`/payments/${id}`, { auth: true })
      .then((p) => {
        setPayment(p);
        if (p.status === "paid") setView("paid");
        else if (returnedStatus === "cancel") setView("canceled");
        else if (returnedStatus === "success") verify(); // gateway return → confirm
        else setView("confirm"); // stub / manual
      })
      .catch((err) => {
        setView("error");
        setError(err instanceof ApiError ? err.message : "تعذّر تحميل عملية الدفع");
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [authLoading, user, id]);

  if (authLoading || view === "loading") return <Spinner />;

  const amount = payment ? formatOMR(payment.amount) : "";
  const purposeLabel = payment?.purpose === "renewal" ? "تجديد إعلان" : "رسوم نشر إعلان";

  return (
    <div className="mx-auto max-w-md">
      <div className="card space-y-4 p-8 text-center">
        {view === "paid" ? (
          <>
            <div className="text-4xl">✓</div>
            <h1 className="text-xl font-extrabold text-slate-800">تم الدفع بنجاح</h1>
            <p className="text-sm text-slate-500">
              تم استلام إعلانك وهو الآن قيد المراجعة من قبل الإدارة قبل نشره لمدة شهر.
            </p>
            <Link href="/dashboard">
              <Button className="w-full">الذهاب إلى لوحتي</Button>
            </Link>
          </>
        ) : view === "canceled" ? (
          <>
            <div className="text-4xl">✕</div>
            <h1 className="text-xl font-extrabold text-slate-800">تم إلغاء الدفع</h1>
            <p className="text-sm text-slate-500">لم تكتمل عملية الدفع. يمكنك المحاولة مرة أخرى من لوحتك.</p>
            <Button className="w-full" onClick={() => setView("confirm")}>
              إعادة المحاولة
            </Button>
          </>
        ) : (
          <>
            <h1 className="text-xl font-extrabold text-slate-800">إتمام الدفع</h1>
            <div className="rounded-xl bg-slate-50 p-4">
              <div className="text-sm text-slate-500">{purposeLabel}</div>
              <div className="mt-1 text-3xl font-black text-brand-700">{amount}</div>
              <div className="text-xs text-slate-400">لمدة شهر واحد</div>
            </div>
            {error && <p className="text-sm text-red-600">{error}</p>}
            <Button className="w-full" onClick={verify} disabled={view === "verifying"}>
              {view === "verifying" ? "جارٍ التأكيد..." : "ادفع الآن"}
            </Button>
            <p className="text-xs text-slate-400">
              الدفع الإلكتروني عبر بوابة الدفع. بعد إتمام الدفع يُرسل إعلانك للمراجعة تلقائياً.
            </p>
            <Link href="/dashboard" className="block text-xs text-slate-400 hover:text-slate-600">
              الدفع لاحقاً
            </Link>
          </>
        )}
      </div>
    </div>
  );
}
