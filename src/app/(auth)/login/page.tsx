import { redirect } from "next/navigation";
import { getCurrentUser } from "@/auth";
import { LoginForm } from "./_components/login-form";
import Logo from "@/components/base/logo";

export const metadata = {
  title: "Login - Liyali Gateway",
  description: "Sign in to your account",
};

export default async function LoginPage() {
  // If already logged in, redirect to dashboard
  const user = await getCurrentUser();
  if (user) {
    redirect("/workflows/dashboard");
  }

  return (
    <div className="w-full max-w-md">
      {/* Card */}
      <div className="bg-card rounded-lg p-8 space-y-6">
        {/* Logo/Title */}
        <div className="text-center space-y-2">
          <Logo isFull />
        </div>

        {/* Login Form */}
        <LoginForm />

        {/* Demo Info */}
        <div className="border-t pt-6">
          <h3 className="text-sm font-semibold text-foreground mb-3">
            Demo Accounts
          </h3>
          <div className="space-y-2 text-xs">
            <div className="flex justify-between items-start">
              <span className="text-muted-foreground">
                <strong>Requester:</strong> requester@liyali.com
              </span>
              <span className="text-primary font-medium">👤</span>
            </div>
            <div className="flex justify-between items-start">
              <span className="text-muted-foreground">
                <strong>Manager:</strong> manager@liyali.com
              </span>
              <span className="text-primary font-medium">👥</span>
            </div>
            <div className="flex justify-between items-start">
              <span className="text-muted-foreground">
                <strong>Finance:</strong> finance@liyali.com
              </span>
              <span className="text-primary font-medium">💼</span>
            </div>
            <div className="flex justify-between items-start">
              <span className="text-muted-foreground">
                <strong>Director:</strong> director@liyali.com
              </span>
              <span className="text-primary font-medium">👔</span>
            </div>
            <div className="flex justify-between items-start">
              <span className="text-muted-foreground">
                <strong>CFO:</strong> cfo@liyali.com
              </span>
              <span className="text-primary font-medium">💎</span>
            </div>
            <div className="flex justify-between items-start">
              <span className="text-muted-foreground">
                <strong>Compliance:</strong> compliance@liyali.com
              </span>
              <span className="text-primary font-medium">✅</span>
            </div>
            <div className="flex justify-between items-start">
              <span className="text-muted-foreground">
                <strong>Admin:</strong> admin@liyali.com
              </span>
              <span className="text-primary font-medium">⚙️</span>
            </div>
            <p className="text-muted-foreground italic pt-2">
              Password:{" "}
              <code className="bg-muted px-2 py-1 rounded text-xs">
                password123
              </code>
            </p>
          </div>

        {/* Footer */}
        <div className="text-center text-xs text-muted-foreground">
          <p>🔒 Simulated Authentication System</p>
          <p>For development and testing purposes only</p>
        </div>
        </div>
      </div>
    </div>
  );
}
