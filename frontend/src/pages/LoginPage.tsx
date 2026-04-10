import { useNavigate } from "react-router-dom";
import { AuthFormCard } from "@/features/auth/AuthFormCard";
import { useAuth } from "@/features/auth/AuthContext";
import { type LoginFormValues, loginSchema } from "@/features/auth/authSchemas";

export function LoginPage() {
  const navigate = useNavigate();
  const { setSession } = useAuth();

  function handleLogin(values: LoginFormValues) {
    const email = values.email.trim().toLowerCase();

    setSession({
      token: `demo-token-${crypto.randomUUID()}`,
      user: {
        id: crypto.randomUUID(),
        name: email.split("@")[0] || "Demo User",
        email,
      },
    });

    navigate("/projects", { replace: true });
  }

  return (
    <AuthFormCard
      title="Welcome back"
      description="Sign in to continue into your task workspace."
      submitLabel="Login"
      schema={loginSchema}
      onSubmit={handleLogin}
      helperMessage="Frontend auth state is wired to localStorage. Login currently creates a local demo session until backend API integration is added."
      fields={[
        { name: "email", label: "Email", type: "email" },
        { name: "password", label: "Password", type: "password" },
      ]}
    />
  );
}
