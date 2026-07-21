"use client";

import { useState } from "react";
import { ListingImage } from "@/lib/types";

// Client island for image switching. Server-renders the primary image into the
// initial HTML (good for SEO), then hydrates for thumbnail navigation.
export function Gallery({ images, title }: { images: ListingImage[]; title: string }) {
  const [active, setActive] = useState(0);
  return (
    <div className="card overflow-hidden">
      <div className="aspect-[16/10] bg-slate-100">
        {images.length > 0 ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={images[active].url} alt={title} className="h-full w-full object-cover" />
        ) : (
          <div className="grid h-full place-items-center text-slate-300">لا توجد صور</div>
        )}
      </div>
      {images.length > 1 && (
        <div className="flex gap-2 overflow-x-auto p-3">
          {images.map((img, i) => (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              key={img.id}
              src={img.url}
              alt=""
              onClick={() => setActive(i)}
              className={`h-16 w-20 flex-shrink-0 cursor-pointer rounded-lg object-cover ${
                i === active ? "ring-2 ring-brand-500" : ""
              }`}
            />
          ))}
        </div>
      )}
    </div>
  );
}
