import { Navigate, useNavigate } from "react-router-dom";
import { useState } from "react";
import { AuthFormCard } from "@/features/auth/AuthFormCard";
import { useAuth } from "@/features/auth/AuthContext";
import { type LoginFormValues, loginSchema } from "@/features/auth/authSchemas";
import { ApiError } from "@/api/http";
import { useLoginMutation } from "@/features/auth/useAuthMutations";

export function LoginPage() {
  const navigate = useNavigate();
  const { isAuthenticated, setSession } = useAuth();
  const loginMutation = useLoginMutation();
  const [apiError, setApiError] = useState<string | null>(null);
  const [serverFieldErrors, setServerFieldErrors] = useState<
    Partial<Record<keyof LoginFormValues, string>>
  >({});

  if (isAuthenticated) {
    return <Navigate to="/projects" replace />;
  }

  async function handleLogin(values: LoginFormValues) {
    setApiError(null);
    setServerFieldErrors({});

    try {
      const session = await loginMutation.mutateAsync({
        email: values.email.trim().toLowerCase(),
        password: values.password,
      });

      setSession(session);
    } catch (error) {
      if (error instanceof ApiError) {
        setApiError(error.message);
        setServerFieldErrors(error.fields as Partial<Record<keyof LoginFormValues, string>>);
        return;
      }

      setApiError("Unable to sign in right now. Please try again.");
      return;
    }

    navigate("/projects", { replace: true });
  }

  return (
    <AuthFormCard
      title="Welcome back"
      description="Sign in to continue into your task workspace."
      submitLabel="Login"
      schema={loginSchema}
      onSubmit={handleLogin}
      helperMessage="Sign in with your backend account. Successful login stores the token and user session locally."
      apiError={apiError}
      serverFieldErrors={serverFieldErrors}
      submitInProgress={loginMutation.isPending}
      fields={[
        { name: "email", label: "Email", type: "email" },
        { name: "password", label: "Password", type: "password" },
      ]}
    />
  );
}
