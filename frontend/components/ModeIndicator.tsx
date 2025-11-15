"use client";

export default function ModeIndicator() {
  const appMode = process.env.NEXT_PUBLIC_APP_MODE || "development";
  const isProduction = appMode === "production";

  if (isProduction) {
    return (
      <div className="fixed top-0 left-0 right-0 bg-green-600 text-white text-center py-1.5 text-xs font-semibold z-50 shadow-md">
        🚀 PRODUCTION MODE
      </div>
    );
  }

  return (
    <div className="fixed top-0 left-0 right-0 bg-yellow-500 text-black text-center py-1.5 text-xs font-semibold z-50 shadow-md">
      🔧 DEVELOPMENT MODE
    </div>
  );
}
