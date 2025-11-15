import { create } from "zustand";
import { Admin, SuperAdmin } from "./api";

interface AuthState {
  user: Admin | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (userData: Admin, token: string) => void;
  logout: () => void;
  checkAuth: () => void;
  initialized: boolean;
}

interface SuperAdminAuthState {
  user: SuperAdmin | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (userData: SuperAdmin, token: string) => void;
  logout: () => void;
  checkAuth: () => void;
  initialized: boolean;
}

// Helper to get stored user data
const getStoredUser = (): Admin | null => {
  if (typeof window === "undefined") return null;
  const stored = localStorage.getItem("user");
  if (stored) {
    try {
      return JSON.parse(stored);
    } catch {
      return null;
    }
  }
  return null;
};

export const useAuthStore = create<AuthState>((set) => {
  return {
    user: null,
    token: null,
    isAuthenticated: false,
    initialized: false,
    login: (userData, token) => {
      if (typeof window !== "undefined") {
        localStorage.setItem("token", token);
        localStorage.setItem("user", JSON.stringify(userData));
      }
      set({ user: userData, token, isAuthenticated: true, initialized: true });
    },
    logout: () => {
      if (typeof window !== "undefined") {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
      }
      set({
        user: null,
        token: null,
        isAuthenticated: false,
        initialized: true,
      });
    },
    checkAuth: () => {
      if (typeof window !== "undefined") {
        const token = localStorage.getItem("token");
        const user = getStoredUser();
        if (token) {
          set({ token, user, isAuthenticated: true, initialized: true });
        } else {
          set({
            token: null,
            user: null,
            isAuthenticated: false,
            initialized: true,
          });
        }
      } else {
        set({ initialized: true });
      }
    },
  };
});

// Helper to get stored super admin data
const getStoredSuperAdmin = (): SuperAdmin | null => {
  if (typeof window === "undefined") return null;
  const stored = localStorage.getItem("super_admin");
  if (stored) {
    try {
      return JSON.parse(stored);
    } catch {
      return null;
    }
  }
  return null;
};

export const useSuperAdminStore = create<SuperAdminAuthState>((set) => {
  return {
    user: null,
    token: null,
    isAuthenticated: false,
    initialized: false,
    login: (userData, token) => {
      if (typeof window !== "undefined") {
        localStorage.setItem("super_admin_token", token);
        localStorage.setItem("super_admin", JSON.stringify(userData));
      }
      set({
        user: userData,
        token,
        isAuthenticated: true,
        initialized: true,
      });
    },
    logout: () => {
      if (typeof window !== "undefined") {
        localStorage.removeItem("super_admin_token");
        localStorage.removeItem("super_admin");
      }
      set({
        user: null,
        token: null,
        isAuthenticated: false,
        initialized: true,
      });
    },
    checkAuth: () => {
      if (typeof window !== "undefined") {
        const token = localStorage.getItem("super_admin_token");
        const user = getStoredSuperAdmin();
        if (token) {
          set({ token, user, isAuthenticated: true, initialized: true });
        } else {
          set({
            token: null,
            user: null,
            isAuthenticated: false,
            initialized: true,
          });
        }
      } else {
        set({ initialized: true });
      }
    },
  };
});
