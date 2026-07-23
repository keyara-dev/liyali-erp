import { redirect } from "next/navigation";
import { LoginForm } from "./_components/login-form";
import Logo from "@/components/base/logo";
import { verifySession } from "@/lib/auth";

export const metadata = {
  title: "Login - Liyali Gateway",
  description: "Sign in to your account",
};

export default async function LoginPage() {
  // If already logged in, redirect to dashboard
  const { session } = await verifySession();

  if (session && session.user) {
    redirect("/welcome");
  }

  return (
    <div className="w-full max-w-md">
      {/* Card */}
      <div className="bg-card rounded-lg p-8 space-y-6">
        {/* Logo/Title */}
        <div className="text-left space-y-2 mb-8">
          <Logo isFull href="/" />
           
        </div>

        {/* Login Form */}
        <LoginForm />
      </div>

        {/* <div className="pt-4 border-t border-border">
          <p className="text-xs text-muted-foreground text-center">
            Need help? Contact support at support@liyali.com
          </p>
        </div> */}
    </div>
  );
}
