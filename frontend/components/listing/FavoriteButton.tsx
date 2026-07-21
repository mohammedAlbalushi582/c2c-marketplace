"use client";

import { useState } from "react";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { Button } from "@/components/ui";

export function FavoriteButton({ id }: { id: number }) {
  const { user } = useAuth();
  const [fav, setFav] = useState(false);

  async function toggle() {
    if (!user) {
      window.location.href = "/login";
      return;
    }
    try {
      await api(`/listings/${id}/favorite`, { method: fav ? "DELETE" : "POST", auth: true });
      setFav(!fav);
    } catch {
      /* ignore */
    }
  }

  return (
    <Button variant="ghost" onClick={toggle} className="w-full">
      {fav ? "★ في المفضلة" : "☆ أضف للمفضلة"}
    </Button>
  );
}
