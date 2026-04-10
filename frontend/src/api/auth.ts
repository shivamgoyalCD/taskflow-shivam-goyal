import { apiClient } from "@/api/client";
import type { AuthSession } from "@/types/auth";

export type LoginPayload = {
  email: string;
  password: string;
};

export type RegisterPayload = {
  name: string;
  email: string;
  password: string;
};

export function login(payload: LoginPayload) {
  return apiClient.post<AuthSession>("/auth/login", payload);
}

export function register(payload: RegisterPayload) {
  return apiClient.post<AuthSession>("/auth/register", payload);
}
