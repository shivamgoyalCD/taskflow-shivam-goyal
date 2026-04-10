import { AuthFormCard } from "@/features/auth/AuthFormCard";
import { loginSchema } from "@/features/auth/authSchemas";

export function LoginPage() {
  return (
    <AuthFormCard
      title="Welcome back"
      description="Sign in to continue into your task workspace."
      submitLabel="Login"
      schema={loginSchema}
      fields={[
        { name: "email", label: "Email", type: "email" },
        { name: "password", label: "Password", type: "password" },
      ]}
    />
  );
}
