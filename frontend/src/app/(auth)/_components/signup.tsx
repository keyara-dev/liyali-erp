"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components";
import { useSignupMutation } from "@/hooks/use-auth-mutations";
import { EyeIcon, EyeOffIcon, AlertCircle } from "lucide-react";
import Link from "next/link";

export default function Signup() {
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
        role: "requester",
      });

      if (!result.success) {
        setError(result.message || "Registration failed");
      }
    } catch (err: any) {
      setError(err.message || "An error occurred");
    }
  };

  return (
    <div className="w-full max-w-md">
      {/* Header */}
      <div className="text-center mb-12">
        <h1 className="text-4xl font-light text-black mb-4">
          Create your <span className="font-bold">account</span>
        </h1>
        <p className="text-gray-600 leading-relaxed">
          Join us and start using Liyali Gateway in minutes.
        </p>
      </div>

      {/* Error Message */}
      {error && (
        <div className="flex items-center gap-2 p-3 mb-6 bg-red-50 border border-red-300 rounded-lg">
          <AlertCircle className="h-4 w-4 text-red-600 shrink-0" />
          <p className="text-sm text-red-600">{error}</p>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Email Input */}
        <div>
          <label
            htmlFor="email"
            className="block text-sm font-medium text-black mb-2"
          >
            Email Address
          </label>
          <input
            type="email"
            id="email"
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent transition-colors placeholder:text-gray-400 text-black"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            disabled={isPending}
          />
        </div>

        {/* Name Input */}
        <div>
          <label
            htmlFor="name"
            className="block text-sm font-medium text-black mb-2"
          >
            Full Name
          </label>
          <input
            type="text"
            id="name"
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent transition-colors placeholder:text-gray-400 text-black"
            placeholder="John Doe"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            disabled={isPending}
          />
        </div>

        {/* Password Input */}
        <div className="relative">
          <label
            htmlFor="password"
            className="block text-sm font-medium text-black mb-2"
          >
            Password
          </label>
          <input
            type={showPassword.password ? "text" : "password"}
            id="password"
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent transition-colors placeholder:text-gray-400 text-black"
            placeholder="Enter a strong password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            disabled={isPending}
            autoComplete="new-password"
          />
          {password && password.length > 0 && (
            <button
              type="button"
              className="absolute cursor-pointer right-3 top-[55%] transform -translate-y-1/2 text-gray-400"
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
          <p className="text-xs text-gray-500 mt-1">
            Must have: 8+ characters, 1 uppercase, 1 lowercase, 1 digit
          </p>
        </div>

        {/* Confirm Password Input */}
        <div className="relative">
          <label
            htmlFor="confirmPassword"
            className="block text-sm font-medium text-black mb-2"
          >
            Confirm Password
          </label>
          <input
            type={showPassword.confirmPassword ? "text" : "password"}
            id="confirmPassword"
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent transition-colors placeholder:text-gray-400 text-black"
            placeholder="Confirm your password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            disabled={isPending}
            autoComplete="new-password"
          />
          {confirmPassword && confirmPassword.length > 0 && (
            <button
              type="button"
              className="absolute cursor-pointer right-3 top-[55%] transform -translate-y-1/2 text-gray-400"
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

        {/* Submit Button */}
        <Button
          type="submit"
          disabled={isPending}
          className="w-full bg-black hover:bg-black/90 text-white"
          isLoading={isPending}
          loadingText="Creating account..."
        >
          Sign Up
        </Button>
      </form>

      {/* Login Link */}
      <div className="mt-8 text-center">
        <p className="text-gray-600">
          Already have an account?{" "}
          <Link
            href="/login"
            className="text-black font-medium hover:underline"
          >
            Login
          </Link>
        </p>
      </div>
    </div>
  );
}
