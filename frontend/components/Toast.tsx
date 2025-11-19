"use client";

import { useEffect, useState } from "react";

export type ToastType = "success" | "error" | "info" | "warning";

interface Toast {
  id: string;
  message: string;
  type: ToastType;
}

let toastId = 0;
const listeners: Array<(toasts: Toast[]) => void> = [];
let toasts: Toast[] = [];

const toast = {
  success: (message: string) => {
    const id = `toast-${toastId++}`;
    toasts = [...toasts, { id, message, type: "success" }];
    listeners.forEach((listener) => listener(toasts));
    setTimeout(() => {
      toasts = toasts.filter((t) => t.id !== id);
      listeners.forEach((listener) => listener(toasts));
    }, 4000);
  },
  error: (message: string) => {
    const id = `toast-${toastId++}`;
    toasts = [...toasts, { id, message, type: "error" }];
    listeners.forEach((listener) => listener(toasts));
    setTimeout(() => {
      toasts = toasts.filter((t) => t.id !== id);
      listeners.forEach((listener) => listener(toasts));
    }, 5000);
  },
  info: (message: string) => {
    const id = `toast-${toastId++}`;
    toasts = [...toasts, { id, message, type: "info" }];
    listeners.forEach((listener) => listener(toasts));
    setTimeout(() => {
      toasts = toasts.filter((t) => t.id !== id);
      listeners.forEach((listener) => listener(toasts));
    }, 4000);
  },
  warning: (message: string) => {
    const id = `toast-${toastId++}`;
    toasts = [...toasts, { id, message, type: "warning" }];
    listeners.forEach((listener) => listener(toasts));
    setTimeout(() => {
      toasts = toasts.filter((t) => t.id !== id);
      listeners.forEach((listener) => listener(toasts));
    }, 4000);
  },
};

export { toast };

export default function ToastContainer() {
  const [currentToasts, setCurrentToasts] = useState<Toast[]>([]);

  useEffect(() => {
    const listener = (newToasts: Toast[]) => {
      setCurrentToasts(newToasts);
    };
    listeners.push(listener);
    setCurrentToasts(toasts);

    return () => {
      const index = listeners.indexOf(listener);
      if (index > -1) {
        listeners.splice(index, 1);
      }
    };
  }, []);

  const getToastStyles = (type: ToastType) => {
    switch (type) {
      case "success":
        return "bg-green-50 border-green-200 text-green-800";
      case "error":
        return "bg-red-50 border-red-200 text-red-800";
      case "info":
        return "bg-blue-50 border-blue-200 text-blue-800";
      case "warning":
        return "bg-yellow-50 border-yellow-200 text-yellow-800";
    }
  };

  const getIcon = (type: ToastType) => {
    switch (type) {
      case "success":
        return "✓";
      case "error":
        return "✗";
      case "info":
        return "ℹ";
      case "warning":
        return "⚠";
    }
  };

  return (
    <div className="fixed top-4 right-4 z-50 space-y-2">
      {currentToasts.map((toast) => (
        <div
          key={toast.id}
          className={`min-w-[300px] max-w-md px-4 py-3 rounded-lg shadow-lg border flex items-start gap-3 animate-slide-in ${getToastStyles(
            toast.type
          )}`}
        >
          <span className="text-lg font-bold flex-shrink-0">
            {getIcon(toast.type)}
          </span>
          <p className="text-sm font-medium flex-1">{toast.message}</p>
          <button
            onClick={() => {
              toasts = toasts.filter((t) => t.id !== toast.id);
              listeners.forEach((listener) => listener(toasts));
            }}
            className="text-gray-500 hover:text-gray-700 flex-shrink-0"
          >
            ×
          </button>
        </div>
      ))}
      <style jsx>{`
        @keyframes slide-in {
          from {
            transform: translateX(100%);
            opacity: 0;
          }
          to {
            transform: translateX(0);
            opacity: 1;
          }
        }
        .animate-slide-in {
          animation: slide-in 0.3s ease-out;
        }
      `}</style>
    </div>
  );
}
