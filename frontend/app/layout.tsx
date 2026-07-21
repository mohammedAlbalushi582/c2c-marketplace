import type { Metadata } from "next";
import "./globals.css";
import { AuthProvider } from "@/lib/auth";
import { Header } from "@/components/Header";
import { PresenceTracker } from "@/components/PresenceTracker";
import { SITE_URL } from "@/lib/server-api";

const SITE_NAME = "الأمجاد للأعمال والعقارات";
const SITE_DESC =
  "سوق إلكتروني في سلطنة عُمان لبيع وشراء الأراضي والعقارات والمنازل وقوالب ومواقع الويب الجاهزة. أضف إعلانك مجاناً وابحث عن أفضل العروض.";

export const metadata: Metadata = {
  metadataBase: new URL(SITE_URL),
  title: {
    default: `${SITE_NAME} — أراضٍ وعقارات وقوالب مواقع في عُمان`,
    template: `%s | ${SITE_NAME}`,
  },
  description: SITE_DESC,
  keywords: [
    "عقارات عمان",
    "أراضي للبيع",
    "منازل للبيع",
    "قوالب مواقع",
    "قوالب جاهزة",
    "مواقع للبيع",
    "سوق عمان",
    "إعلانات مبوبة",
    "OpenSooq",
  ],
  alternates: { canonical: "/" },
  openGraph: {
    type: "website",
    locale: "ar_OM",
    siteName: SITE_NAME,
    title: SITE_NAME,
    description: SITE_DESC,
    url: SITE_URL,
  },
  twitter: { card: "summary_large_image", title: SITE_NAME, description: SITE_DESC },
  robots: { index: true, follow: true },
  // Paste the token from Google Search Console (or set GOOGLE_SITE_VERIFICATION).
  verification: process.env.NEXT_PUBLIC_GOOGLE_SITE_VERIFICATION
    ? { google: process.env.NEXT_PUBLIC_GOOGLE_SITE_VERIFICATION }
    : undefined,
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ar" dir="rtl">
      <head>
        {/* Runtime font load; falls back to system fonts if offline. */}
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link
          href="https://fonts.googleapis.com/css2?family=Cairo:wght@400;600;700;800&display=swap"
          rel="stylesheet"
        />
      </head>
      <body className="font-sans min-h-screen bg-slate-50 text-slate-900 antialiased">
        <AuthProvider>
          <PresenceTracker />
          <Header />
          <main className="container py-6">{children}</main>
          <footer className="border-t border-slate-200 bg-white">
            <div className="container py-6 text-center text-sm text-slate-500">
              الأمجاد للأعمال والعقارات — جميع الحقوق محفوظة © 2026
            </div>
          </footer>
        </AuthProvider>
      </body>
    </html>
  );
}
