import { useNavigate } from "react-router-dom";
import { AuthFormCard } from "@/features/auth/AuthFormCard";
import { useAuth } from "@/features/auth/AuthContext";
import { type RegisterFormValues, registerSchema } from "@/features/auth/authSchemas";

export function RegisterPage() {
  const navigate = useNavigate();
  const { setSession } = useAuth();

  function handleRegister(values: RegisterFormValues) {
    setSession({
      token: `demo-token-${crypto.randomUUID()}`,
      user: {
        id: crypto.randomUUID(),
        name: values.name.trim(),
        email: values.email.trim().toLowerCase(),
      },
    });

    navigate("/projects", { replace: true });
  }

  return (
    <AuthFormCard
      title="Create your account"
      description="Register a new user to start organizing projects and tasks."
      submitLabel="Register"
      schema={registerSchema}
      onSubmit={handleRegister}
      helperMessage="Register currently creates a local demo session and redirects into the protected app shell. API integration can replace this without changing the auth store."
      fields={[
        { name: "name", label: "Full name" },
        { name: "email", label: "Email", type: "email" },
        { name: "password", label: "Password", type: "password" },
      ]}
    />
  );
}
