"use client";

import { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useSignupMutation } from "@/hooks/use-auth-mutations";
import { EyeIcon, EyeOffIcon, AlertCircle, Crown, Clock } from "lucide-react";
import Link from "next/link";
import Logo from "@/components/base/logo";
import { motion } from "framer-motion";

export default function Signup() {
  const searchParams = useSearchParams();
  const selectedPlan = searchParams.get("plan");
  const showTrial = searchParams.get("trial") === "true";

  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState({
    password: false,
    confirmPassword: false,
  });
  const [error, setError] = useState("");
  const { signup, isPending } = useSignupMutation();

  // Plan display names
  const planNames = {
    STARTER_PLAN: "Starter",
    PRO_PLAN: "Pro",
    ENTERPRISE: "Enterprise",
  };

  const validatePassword = (pwd: string): string[] => {
    const errors: string[] = [];
    if (pwd.length < 8) errors.push("At least 8 characters");
    if (!/[A-Z]/.test(pwd)) errors.push("At least 1 uppercase letter");
    if (!/[a-z]/.test(pwd)) errors.push("At least 1 lowercase letter");
    if (!/[0-9]/.test(pwd)) errors.push("At least 1 digit");
    return errors;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    // Validation
    if (!email || !name || !password || !confirmPassword) {
      setError("Please fill in all required fields");
      return;
    }

    if (password !== confirmPassword) {
      setError("Passwords do not match. Please try again.");
      return;
    }

    const passwordErrors = validatePassword(password);
    if (passwordErrors.length > 0) {
      setError(`Password requirements: ${passwordErrors.join(", ")}`);
      return;
    }

    try {
      const result = await signup({
        email,
        name,
        password,
        role: "admin", // Default to admin since they get their own organization
      });

      if (result.success && selectedPlan) {
        // Store selected plan in localStorage to redirect after login
        localStorage.setItem("selectedPlan", selectedPlan);
        localStorage.setItem("redirectToUpgrade", "true");
      }

      if (!result.success) {
        setError(result.message || "Registration failed");
      }
    } catch (err: any) {
      setError(err.message || "An error occurred");
    }
  };

  return (
    <div className="w-full max-w-md">
      {/* Logo/Title */}
      <div className="text-left space-y-2 mb-3">
        <Logo isFull href="/" />
      </div>

      {/* Selected Plan Banner */}
      {selectedPlan && (
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-6 p-4 bg-gradient-to-r from-blue-50 to-purple-50 border border-blue-200 rounded-lg"
        >
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-100 rounded-full">
              <Crown className="h-4 w-4 text-blue-600" />
            </div>
            <div>
              <h3 className="font-semibold text-slate-900">
                {planNames[selectedPlan as keyof typeof planNames]} Plan Selected
              </h3>
              <p className="text-sm text-slate-600">
                You'll be redirected to complete your subscription after registration
              </p>
            </div>
          </div>
        </motion.div>
      )}

      {/* Free Trial Banner */}
      {showTrial && (
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="mb-6 p-4 bg-gradient-to-r from-green-50 to-blue-50 border border-green-200 rounded-lg"
        >
          <div className="flex items-center gap-3">
            <div className="p-2 bg-green-100 rounded-full">
              <Clock className="h-4 w-4 text-green-600" />
            </div>
            <div>
              <h3 className="font-semibold text-slate-900">
                🎉 14-Day Free Trial
              </h3>
              <p className="text-sm text-slate-600">
                Try all features risk-free. No credit card required to start.
              </p>
            </div>
          </div>
        </motion.div>
      )}

      {/* Header */}
      <div className="text-left mb-6">
        <h1 className="text-3xl font-bold text-slate-900 mb-2">
          Create your account
        </h1>
        <p className="text-slate-600">
          Join Liyali and start streamlining your operations with ease.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-2">
        {/* Name Input */}
        <Input
          id="name"
          type="text"
          label="Full Name"
          placeholder="Bob Mwale"
          className="bg-muted"
          value={name}
          onChange={(e) => setName(e.target.value)}
          disabled={isPending}
          required
        />

        {/* Email Input */}
        <Input
          id="email"
          type="email"
          label="Email"
          placeholder="bob.mwale@mail.com"
          className="bg-muted"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          disabled={isPending}
          required
        />

        {/* Password Input */}
        <div className="relative">
          <Input
            id="password"
            type={showPassword.password ? "text" : "password"}
            label="Password"
            placeholder="••••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={isPending}
            required
            autoComplete="new-password"
            className="bg-muted"
            descriptionText="Must have: 8+ characters, 1 uppercase, 1 lowercase, 1 digit"
          />
          {password && password.length > 0 && (
            <button
              type="button"
              className="absolute right-3 top-9 text-slate-400 hover:text-slate-600"
              onClick={() =>
                setShowPassword((prev) => ({
                  ...prev,
                  password: !showPassword.password,
                }))
              }
              disabled={isPending}
            >
              {showPassword.password ? (
                <EyeOffIcon className="w-5 h-5" />
              ) : (
                <EyeIcon className="w-5 h-5" />
              )}
            </button>
          )}
        </div>

        {/* Confirm Password Input */}
        <div className="relative">
          <Input
            id="confirmPassword"
            type={showPassword.confirmPassword ? "text" : "password"}
            label="Confirm password"
            placeholder="••••••••••"
            className="bg-muted"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            disabled={isPending}
            required
            autoComplete="new-password"
          />
          {confirmPassword && confirmPassword.length > 0 && (
            <button
              type="button"
              className="absolute right-3 top-9 text-slate-400 hover:text-slate-600"
              onClick={() =>
                setShowPassword((prev) => ({
                  ...prev,
                  confirmPassword: !showPassword.confirmPassword,
                }))
              }
              disabled={isPending}
            >
              {showPassword.confirmPassword ? (
                <EyeOffIcon className="w-5 h-5" />
              ) : (
                <EyeIcon className="w-5 h-5" />
              )}
            </button>
          )}
        </div>

        {/* Error Message */}
        {error && (
          <div className="flex items-center gap-2 p-3 bg-red-50 border border-red-200 rounded-lg">
            <AlertCircle className="h-4 w-4 text-red-600 shrink-0" />
            <p className="text-sm text-red-600">{error}</p>
          </div>
        )}

        {/* Submit Button */}
        <Button
          type="submit"
          disabled={isPending}
          className="w-full bg-primary hover:bg-primary/90 text-primary-foreground mt-2"
          isLoading={isPending}
          loadingText="Creating account..."
        >
          Create account
        </Button>

        {/* Divider */}
        <div className="relative my-6">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-slate-200" />
          </div>
          <div className="relative flex justify-center text-sm">
            <span className="px-2 bg-card text-slate-500">OR</span>
          </div>
        </div>

        {/* Google Signup Button */}
        <button
          type="button"
          disabled
          className="w-full flex items-center justify-center gap-3 px-4 py-3 border border-slate-200 rounded-lg bg-card text-slate-400 cursor-not-allowed transition-colors"
        >
          <svg className="w-5 h-5" viewBox="0 0 24 24">
            <path
              fill="#4285F4"
              d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
            />
            <path
              fill="#34A853"
              d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
            />
            <path
              fill="#FBBC05"
              d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
            />
            <path
              fill="#EA4335"
              d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
            />
          </svg>
          <span>Sign up with Google (Coming soon)</span>
        </button>

        {/* Login Link */}
        <div className="mt-6 text-center">
          <p className="text-slate-600">
            Already have an account?{" "}
            <Link
              href="/login"
              className={`font-medium transition-colors ${
                isPending
                  ? "text-muted-foreground cursor-not-allowed pointer-events-none"
                  : "text-primary hover:text-primary/80"
              }`}
            >
              Log in
            </Link>
          </p>
        </div>
      </form>
    </div>
  );
}
