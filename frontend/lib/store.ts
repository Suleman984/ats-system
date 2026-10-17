import { create } from "zustand";
import {
  ADMIN_USER_KEY,
  SUPER_ADMIN_USER_KEY,
  authAPI,
  superAdminAPI,
  type Admin,
  type SuperAdmin,
} from "./api";

// Global because the signed-in user is read by every dashboard layout and page.
// The session itself is an HttpOnly cookie; only display data is kept in localStorage.
interface SessionState<U> {
  user: U | null;
  isAuthenticated: boolean;
  initialized: boolean;
  login: (user: U) => void;
  logout: () => Promise<void>;
  checkAuth: () => void;
}

const readStored = <U,>(key: string): U | null => {
  if (typeof window === "undefined") return null;
  try {
    const stored = localStorage.getItem(key);
    return stored ? (JSON.parse(stored) as U) : null;
  } catch {
    return null;
  }
};

const createSessionStore = <U,>(
  storageKey: string,
  legacyTokenKey: string,
  serverLogout: () => Promise<unknown>
) =>
  create<SessionState<U>>((set) => ({
    user: null,
    isAuthenticated: false,
    initialized: false,
    login: (user) => {
      localStorage.setItem(storageKey, JSON.stringify(user));
      set({ user, isAuthenticated: true, initialized: true });
    },
    logout: async () => {
      // Clear locally even if the request fails; the cookie expires on its own.
      await serverLogout().catch(() => undefined);
      localStorage.removeItem(storageKey);
      set({ user: null, isAuthenticated: false, initialized: true });
    },
    // Optimistic: trusts the stored user until an API call says otherwise (the 401 handler signs out).
    checkAuth: () => {
      // Tokens used to be stored here; drop any leftovers so scripts can't read them.
      localStorage.removeItem(legacyTokenKey);
      const user = readStored<U>(storageKey);
      set({ user, isAuthenticated: Boolean(user), initialized: true });
    },
  }));

export const useAuthStore = createSessionStore<Admin>(ADMIN_USER_KEY, "token", authAPI.logout);
export const useSuperAdminStore = createSessionStore<SuperAdmin>(
  SUPER_ADMIN_USER_KEY,
  "super_admin_token",
  superAdminAPI.logout
);

// Pulls the latest name/role from the server, e.g. after an owner changes someone's role.
export async function refreshAdmin() {
  const { data } = await authAPI.me();
  useAuthStore.getState().login(data.admin);
}
