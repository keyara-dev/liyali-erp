import type { DashboardData } from "@/types/dashboard";
import {
  fetchSignupAnalytics,
  fetchSignupSettings,
} from "./_actions/dashboard";
import { fetchEventHosts } from "./_actions/event-hosts";
import { DashboardClient } from "./_components/DashboardClient";

export default async function AdminOverviewPage() {
  // Server-side parallel data fetching
  const [
    eventHostsResult,
    // reservedUsernamesResult,
    signupSettingsResult,
    signupAnalyticsResult,
  ] = await Promise.all([
    fetchEventHosts(),
    // fetchReservedUsernames(),
    fetchSignupSettings(),
    fetchSignupAnalytics(),
  ]);

  // Extract data with error handling and null safety
  const eventHosts = eventHostsResult.success
    ? eventHostsResult.data?.data || []
    : [];
  // const reservedUsernames = reservedUsernamesResult.success
  //   ? reservedUsernamesResult.data?.reservedUsernames || []
  //   : [];
  const signupSettings = signupSettingsResult.success
    ? signupSettingsResult.data
    : { enabled: true };
  const signupAnalytics = signupAnalyticsResult.success
    ? signupAnalyticsResult.data
    : { signupData: [], topReferrers: [] };

  // Handle authentication errors
  if (
    !eventHostsResult.success &&
    (eventHostsResult.status === 401 || eventHostsResult.status === 403)
  ) {
    return (
      <>
        <div className="space-y-6">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-red-100 flex items-center justify-center">
              <span className="text-red-600 text-sm">!</span>
            </div>
            <div>
              <h1 className="text-3xl font-bold text-red-600">
                Authentication Required
              </h1>
              <p className="text-gray-600">
                {eventHostsResult.message ||
                  "Please ensure you are logged in as an admin to access this page."}
              </p>
            </div>
          </div>
        </div>
      </>
    );
  }

  const dashboardData: DashboardData = {
    sellers: eventHosts,
    // reservedUsernames,
    signupSettings,
    signupAnalytics,
  };

  return (
    <>
      {/* Page Header - Server-rendered */}
      <div className="mb-8">
        <div className="flex items-center space-x-2 text-sm text-gray-400 mb-2">
          <span>Home</span>
        </div>
        <h1 className="text-3xl font-semibold text-gray-100 mb-2">
          Dashboard Overview
        </h1>
        <p className="text-gray-400">
          Monitor your event ticketing platform performance
        </p>
      </div>

      {/* Interactive Dashboard Content - Client-side */}
      <DashboardClient initialData={dashboardData} />
    </>
  );
}
