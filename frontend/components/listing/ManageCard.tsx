"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { Button, Badge } from "@/components/ui";
import { statusLabel, statusColor } from "@/lib/format";

// Owner/admin controls for a listing. Renders nothing for guests or unrelated
// users, so it stays out of the public (indexable) view.
export function ManageCard({
  id,
  userId,
  initialStatus,
}: {
  id: number;
  userId: number;
  initialStatus: string;
}) {
  const { user } = useAuth();
  const router = useRouter();
  const [status, setStatus] = useState(initialStatus);
  const [error, setError] = useState("");

  if (!user) return null;
  const isAdmin = user.role === "admin";
  const isOwner = user.id === userId;
  if (!isAdmin && !isOwner) return null;

  async function setListingStatus(next: string) {
    setError("");
    try {
      await api(`/admin/listings/${id}/status`, { method: "PATCH", auth: true, body: { status: next } });
      setStatus(next);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "تعذّر تحديث الحالة");
    }
  }

  async function remove() {
    if (!confirm("حذف هذا الإعلان؟ لن يظهر بعد الآن في الموقع.")) return;
    setError("");
    try {
      await api(`/listings/${id}`, { method: "DELETE", auth: true });
      router.push(isAdmin ? "/admin" : "/dashboard");
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "تعذّر حذف الإعلان");
    }
  }

  return (
    <div className="card space-y-3 p-5">
      <div className="flex items-center justify-between">
        <h2 className="font-bold text-slate-800">{isAdmin ? "إدارة الإعلان" : "إعلانك"}</h2>
        <Badge className={statusColor(status)}>{statusLabel(status)}</Badge>
      </div>

      {isAdmin && (
        <div className="flex gap-2">
          {status !== "active" && (
            <Button className="flex-1" onClick={() => setListingStatus("active")}>
              موافقة
            </Button>
          )}
          {status !== "rejected" && (
            <Button variant="outline" className="flex-1" onClick={() => setListingStatus("rejected")}>
              رفض
            </Button>
          )}
        </div>
      )}

      <div className="flex gap-2">
        <Link href={`/post?edit=${id}&from=/listings/${id}`} className="flex-1">
          <Button variant="outline" className="w-full">
            تعديل
          </Button>
        </Link>
        <Button variant="danger" className="flex-1" onClick={remove}>
          حذف
        </Button>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}
    </div>
  );
}
