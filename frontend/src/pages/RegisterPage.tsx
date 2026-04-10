import { Navigate, useNavigate } from "react-router-dom";
import { useState } from "react";
import { AuthFormCard } from "@/features/auth/AuthFormCard";
import { useAuth } from "@/features/auth/AuthContext";
import { type RegisterFormValues, registerSchema } from "@/features/auth/authSchemas";
import { ApiError } from "@/api/http";
import { useRegisterMutation } from "@/features/auth/useAuthMutations";

export function RegisterPage() {
  const navigate = useNavigate();
  const { isAuthenticated, setSession } = useAuth();
  const registerMutation = useRegisterMutation();
  const [apiError, setApiError] = useState<string | null>(null);
  const [serverFieldErrors, setServerFieldErrors] = useState<
    Partial<Record<keyof RegisterFormValues, string>>
  >({});

  if (isAuthenticated) {
    return <Navigate to="/projects" replace />;
  }

  async function handleRegister(values: RegisterFormValues) {
    setApiError(null);
    setServerFieldErrors({});

    try {
      const session = await registerMutation.mutateAsync({
        name: values.name.trim(),
        email: values.email.trim().toLowerCase(),
        password: values.password,
      });

      setSession(session);
    } catch (error) {
      if (error instanceof ApiError) {
        setApiError(error.message);
        setServerFieldErrors(error.fields as Partial<Record<keyof RegisterFormValues, string>>);
        return;
      }

      setApiError("Unable to create your account right now. Please try again.");
      return;
    }

    navigate("/projects", { replace: true });
  }

  return (
    <AuthFormCard
      title="Create your account"
      description="Register a new user to start organizing projects and tasks."
      submitLabel="Register"
      schema={registerSchema}
      onSubmit={handleRegister}
      helperMessage="Create an account against the backend API. On success, the session is stored locally and you are redirected into the app."
      apiError={apiError}
      serverFieldErrors={serverFieldErrors}
      submitInProgress={registerMutation.isPending}
      fields={[
        { name: "name", label: "Full name" },
        { name: "email", label: "Email", type: "email" },
        { name: "password", label: "Password", type: "password" },
      ]}
    />
  );
}
