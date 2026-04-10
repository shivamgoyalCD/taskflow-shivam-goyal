import { apiRequest } from "@/api/http";
import type { AuthSession } from "@/types/auth";

type LoginPayload = {
  email: string;
  password: string;
};

type RegisterPayload = {
  name: string;
  email: string;
  password: string;
};

export function loginRequest(payload: LoginPayload) {
  return apiRequest<AuthSession>("/auth/login", {
    method: "POST",
    body: payload,
  });
}

export function registerRequest(payload: RegisterPayload) {
  return apiRequest<AuthSession>("/auth/register", {
    method: "POST",
    body: payload,
  });
}
