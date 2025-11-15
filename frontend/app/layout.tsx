import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "ATS Platform - Applicant Tracking System",
  description: "Simplify your hiring process with our ATS platform",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
