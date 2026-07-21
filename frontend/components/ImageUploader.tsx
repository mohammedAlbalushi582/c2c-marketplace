"use client";

import { useEffect, useRef, useState } from "react";

export interface ExistingImage {
  id: number;
  url: string;
}

// Dubizzle-style tiled image picker: a grid of square slots. Filled slots show
// a thumbnail (first = الغلاف/cover) with a remove button; the next empty slot is
// a red "+" add button; the rest are camera placeholders. Clicking any empty
// slot opens the file picker.
export function ImageUploader({
  files,
  onFilesChange,
  existing = [],
  onRemoveExisting,
  max = 12,
}: {
  files: File[];
  onFilesChange: (files: File[]) => void;
  existing?: ExistingImage[];
  onRemoveExisting?: (id: number) => void;
  max?: number;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [previews, setPreviews] = useState<string[]>([]);

  // Object URLs for the selected File previews; revoked when files change/unmount.
  useEffect(() => {
    const urls = files.map((f) => URL.createObjectURL(f));
    setPreviews(urls);
    return () => urls.forEach((u) => URL.revokeObjectURL(u));
  }, [files]);

  const filled = existing.length + files.length;
  const showAdd = filled < max;
  const placeholders = Math.max(0, max - filled - (showAdd ? 1 : 0));

  function openPicker() {
    inputRef.current?.click();
  }

  function onPick(e: React.ChangeEvent<HTMLInputElement>) {
    const picked = Array.from(e.target.files || []);
    if (picked.length) {
      const room = max - filled;
      onFilesChange([...files, ...picked.slice(0, room)]);
    }
    e.target.value = ""; // allow re-picking the same file
  }

  function removeFile(idx: number) {
    onFilesChange(files.filter((_, i) => i !== idx));
  }

  let slot = 0; // running index across existing + new, to mark the cover (slot 0)

  return (
    <div>
      <input ref={inputRef} type="file" accept="image/*" multiple onChange={onPick} className="hidden" />
      <div className="grid grid-cols-4 gap-2 sm:grid-cols-6 lg:grid-cols-8">
        {existing.map((img) => {
          const isCover = slot++ === 0;
          return (
            <Tile key={`e${img.id}`} isCover={isCover} onRemove={onRemoveExisting ? () => onRemoveExisting(img.id) : undefined}>
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src={img.url} alt="" className="h-full w-full object-cover" />
            </Tile>
          );
        })}

        {previews.map((src, i) => {
          const isCover = slot++ === 0;
          return (
            <Tile key={`f${i}`} isCover={isCover} onRemove={() => removeFile(i)}>
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src={src} alt="" className="h-full w-full object-cover" />
            </Tile>
          );
        })}

        {showAdd && (
          <button
            type="button"
            onClick={openPicker}
            aria-label="إضافة صور"
            className="flex aspect-square items-center justify-center rounded-xl border-2 border-dashed border-brand-300 bg-brand-50 text-brand-600 transition hover:bg-brand-100"
          >
            <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <path d="M12 5v14M5 12h14" />
            </svg>
          </button>
        )}

        {Array.from({ length: placeholders }).map((_, i) => (
          <button
            key={`p${i}`}
            type="button"
            onClick={openPicker}
            aria-label="إضافة صورة"
            className="flex aspect-square items-center justify-center rounded-xl border border-slate-200 bg-slate-50 text-slate-400 transition hover:bg-slate-100"
          >
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6">
              <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z" />
              <circle cx="12" cy="13" r="4" />
            </svg>
          </button>
        ))}
      </div>
    </div>
  );
}

function Tile({
  children,
  isCover,
  onRemove,
}: {
  children: React.ReactNode;
  isCover: boolean;
  onRemove?: () => void;
}) {
  return (
    <div className="group relative aspect-square overflow-hidden rounded-xl border border-slate-200 bg-slate-100">
      {children}
      {isCover && (
        <span className="absolute bottom-1 right-1 rounded bg-black/60 px-1.5 py-0.5 text-[10px] font-semibold text-white">
          الغلاف
        </span>
      )}
      {onRemove && (
        <button
          type="button"
          onClick={onRemove}
          aria-label="حذف الصورة"
          className="absolute left-1 top-1 grid h-6 w-6 place-items-center rounded-full bg-black/55 text-white transition hover:bg-red-600"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round">
            <path d="M18 6 6 18M6 6l12 12" />
          </svg>
        </button>
      )}
    </div>
  );
}
