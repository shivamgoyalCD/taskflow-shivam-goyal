import { useMutation } from "@tanstack/react-query";
import { loginRequest, registerRequest } from "@/features/auth/authApi";

export function useLoginMutation() {
  return useMutation({
    mutationFn: loginRequest,
  });
}

export function useRegisterMutation() {
  return useMutation({
    mutationFn: registerRequest,
  });
}
