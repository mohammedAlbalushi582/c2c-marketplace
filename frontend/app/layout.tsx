import type { Metadata } from "next";
import "./globals.css";
import { AuthProvider } from "@/lib/auth";
import { Header } from "@/components/Header";

export const metadata: Metadata = {
  title: "الأمجاد للأعمال والعقارات",
  description: "سوق إلكتروني لبيع الأراضي والعقارات وقوالب المواقع في عُمان",
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
