import type { Metadata } from "next";
import "./globals.css";
import ModeIndicator from "@/components/ModeIndicator";
import ToastContainer from "@/components/Toast";

export const metadata: Metadata = {
  metadataBase: new URL(process.env.NEXT_PUBLIC_FRONTEND_URL || "http://localhost:3000"),
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
      <body>
        <ModeIndicator />
        <ToastContainer />
        <div className={process.env.NEXT_PUBLIC_APP_MODE === "production" ? "" : "pt-7"}>{children}</div>
      </body>
    </html>
  );
}
