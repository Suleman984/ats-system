"use client";

// Shown only outside production so testers know they're on a dev build.
export default function ModeIndicator() {
  if (process.env.NEXT_PUBLIC_APP_MODE === "production") return null;

  return (
    <div className="fixed top-0 left-0 right-0 bg-yellow-500 text-black text-center py-1.5 text-xs font-semibold z-50 shadow-md">
      DEVELOPMENT MODE
    </div>
  );
}
