"use client";

import { create } from "zustand";
import { authApi, tokenStore, User, AuthResponse } from "./api";

interface AuthState {
  user: User | null;
  loading: boolean;
  initialized: boolean;
  setUser: (user: User | null) => void;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  bootstrap: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  loading: false,
  initialized: false,

  setUser: (user) => set({ user }),

  login: async (email, password) => {
    set({ loading: true });
    try {
      const data: AuthResponse = await authApi.login(email, password);
      tokenStore.set(data.access_token, data.refresh_token);
      set({ user: data.user });
    } finally {
      set({ loading: false });
    }
  },

  register: async (email, password) => {
    set({ loading: true });
    try {
      const data: AuthResponse = await authApi.register(email, password);
      tokenStore.set(data.access_token, data.refresh_token);
      set({ user: data.user });
    } finally {
      set({ loading: false });
    }
  },

  logout: async () => {
    try {
      await authApi.logout();
    } catch {
      // ignore
    }
    tokenStore.clear();
    set({ user: null });
  },

  bootstrap: async () => {
    console.log("[bootstrap] START");
    console.log("[bootstrap] initialized?", get().initialized);

    if (get().initialized) {
      console.log("[bootstrap] already initialized");
      return;
    }

    const token = tokenStore.getAccess();
    console.log("[bootstrap] token:", token ? "EXISTS" : "NONE");

    if (!token) {
      set({ initialized: true });
      console.log("[bootstrap] no token — initialized=true");
      return;
    }

    try {
      console.log("[bootstrap] calling /auth/me...");
      await authApi.me();
      console.log("[bootstrap] /auth/me OK");

      const payload = JSON.parse(atob(token.split(".")[1]));
      set({
        user: {
          id: payload.uid,
          email: payload.email,
          role: payload.role,
          status: "active",
          created_at: "",
          updated_at: "",
        },
      });
      console.log("[bootstrap] user set");
    } catch (err: any) {
      console.error("[bootstrap] ERROR:", err);
      console.error("[bootstrap] response:", err?.response?.data);
      tokenStore.clear();
    } finally {
      set({ initialized: true });
      console.log("[bootstrap] DONE");
    }
  },
}));
