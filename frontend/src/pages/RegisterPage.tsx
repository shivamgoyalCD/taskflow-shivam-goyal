import { AuthFormCard } from "@/features/auth/AuthFormCard";
import { registerSchema } from "@/features/auth/authSchemas";

export function RegisterPage() {
  return (
    <AuthFormCard
      title="Create your account"
      description="Register a new user to start organizing projects and tasks."
      submitLabel="Register"
      schema={registerSchema}
      fields={[
        { name: "name", label: "Full name" },
        { name: "email", label: "Email", type: "email" },
        { name: "password", label: "Password", type: "password" },
      ]}
    />
  );
}
