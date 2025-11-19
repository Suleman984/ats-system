import type { Metadata } from "next";
import "./globals.css";
import ModeIndicator from "@/components/ModeIndicator";
import ToastContainer from "@/components/Toast";

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
      <body>
        <ModeIndicator />
        <ToastContainer />
        <div className="pt-7">{children}</div>
      </body>
    </html>
  );
}
