"use client";

import { createContext, useContext, useEffect, useState, ReactNode } from "react";
import { api, setTokens, clearTokens, getToken } from "./api";
import { AuthResponse, User } from "./types";

interface AuthContextValue {
  user: User | null;
  loading: boolean;
  login: (identifier: string, password: string) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => void;
}

export interface RegisterData {
  email: string;
  password: string;
  full_name: string;
  phone?: string;
  whatsapp_number?: string;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!getToken()) {
      setLoading(false);
      return;
    }
    api<User>("/auth/me", { auth: true })
      .then(setUser)
      .catch(() => clearTokens())
      .finally(() => setLoading(false));
  }, []);

  async function login(identifier: string, password: string) {
    const res = await api<AuthResponse>("/auth/login", {
      method: "POST",
      body: { identifier, password },
    });
    setTokens(res.access_token, res.refresh_token);
    setUser(res.user);
  }

  async function register(data: RegisterData) {
    const res = await api<AuthResponse>("/auth/register", { method: "POST", body: data });
    setTokens(res.access_token, res.refresh_token);
    setUser(res.user);
  }

  function logout() {
    clearTokens();
    setUser(null);
  }

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
