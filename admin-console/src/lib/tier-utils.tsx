import { Badge } from "@/components/ui/badge";

export function getTierBadge(tier: string) {
  switch (tier) {
    case "enterprise":
    case "custom": // legacy slug — now renamed to enterprise
      return <Badge variant="default">Enterprise</Badge>;
    case "pro":
    case "professional": // legacy slug
      return <Badge className="bg-blue-100 text-blue-800">Professional</Badge>;
    case "starter":
    case "basic": // legacy slug
      return <Badge variant="secondary">Starter</Badge>;
    default:
      return <Badge variant="outline">{tier}</Badge>;
  }
}

export function getTrialStatusBadge(trialStatus: string) {
  switch (trialStatus) {
    case "trial":
      return <Badge variant="secondary">Trial</Badge>;
    case "subscribed":
      return <Badge variant="default">Subscribed</Badge>;
    case "expired":
      return <Badge variant="destructive">Expired</Badge>;
    default:
      return <Badge variant="outline">{trialStatus}</Badge>;
  }
}
