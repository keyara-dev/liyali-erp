import Link from "next/link";
import { ShieldX } from "lucide-react";
import { Button } from "@/components/ui/button";
import { logoutAndRedirect } from "@/app/_actions/auth";

export default function UnauthorizedPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="flex flex-col items-center gap-6 text-center max-w-sm px-4">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
          <ShieldX className="h-8 w-8 text-destructive" />
        </div>
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold tracking-tight">Access Denied</h1>
          <p className="text-sm text-muted-foreground">
            This console is restricted to super administrators only. Your account does not have the required privileges.
          </p>
        </div>
        <form action={logoutAndRedirect}>
          <Button type="submit" variant="outline">
            Back to Login
          </Button>
        </form>
      </div>
    </div>
  );
}
