"use client";

import { useState } from "react";
import Link from "next/link";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useLoginMutation } from "@/hooks/use-auth-mutations";
import { AlertCircle } from "lucide-react";

export function LoginForm() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [rememberMe, setRememberMe] = useState(false);
  const [error, setError] = useState("");
  const { login, isPending } = useLoginMutation();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    try {
      const result = await login({ email, password });

      if (!result.success) {
        setError(result.message || "Login failed");
      }
    } catch (err: any) {
      setError(err.message || "An error occurred");
    }
  };


  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* Error Message */}
      {error && (
        <div className="flex items-center gap-2 p-3 bg-destructive/10 border border-destructive/30 rounded-lg">
          <AlertCircle className="h-4 w-4 text-destructive shrink-0" />
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      <Input
        id="email"
        type="email"
        label="Email Address"
        placeholder="bob.mwale@mail.com"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        disabled={isPending}
        required
        className="bg-muted"
      />

      {/* Password Input */}
      <div className="space-y-1">
        <Input
          id="password"
          type="password"
          placeholder="••••••••"
          label="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          disabled={isPending}
          required
          className="bg-muted"
        />
        
        {/* Forgot Password Link */}
        <div className="text-left">
          <Link 
            href="/forgot-password" 
            className="text-sm text-primary hover:text-primary/80 transition-colors font-medium"
          >
            Forgot password?
          </Link>
        </div>
      </div>

      {/* Remember Me Toggle */}
      {/* <div className="flex items-center justify-between">
        <span className="text-sm text-slate-600">Remember sign in details</span>
        <button
          type="button"
          onClick={() => setRememberMe(!rememberMe)}
          className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 ${
            rememberMe ? 'bg-primary' : 'bg-slate-200'
          }`}
        >
          <span
            className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
              rememberMe ? 'translate-x-6' : 'translate-x-1'
            }`}
          />
        </button>
      </div> */}

      {/* Submit Button */}
      <Button
        type="submit"
        className="w-full bg-primary hover:bg-primary/90 text-primary-foreground"
        disabled={isPending}
        isLoading={isPending}
        loadingText="Logging in..."
      >
        Log in
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

      {/* Google Login Button */}
      <button
        type="button"
        disabled
        className="w-full flex items-center justify-center gap-3 px-4 py-3 border border-slate-200 rounded-lg bg-card text-slate-400 cursor-not-allowed transition-colors"
      >
        <svg className="w-5 h-5" viewBox="0 0 24 24">
          <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
          <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
          <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
          <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
        </svg>
        <span>Continue with Google (Coming soon)</span>
      </button>

      {/* Sign Up Link */}
      <div className="mt-6 text-center">
        <p className="text-slate-600">
          Don't have an account?{" "}
          <Link
            href="/register"
            className="text-primary hover:text-primary/80 font-medium transition-colors"
          >
            Sign up
          </Link>
        </p>
      </div>
    </form>
  );
}
