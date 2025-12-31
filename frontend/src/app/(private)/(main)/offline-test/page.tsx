import { OfflineDemo } from '@/components/offline/offline-demo';
import { OfflineBanner } from '@/components/offline/offline-indicator';
import { PageHeader } from '@/components/base/page-header';

export default function OfflineTestPage() {
  return (
    <div className="space-y-6">
      <OfflineBanner />
      
      <PageHeader
        title="Offline Functionality Test"
        description="Test and demonstrate the offline capabilities of the application"
      />
      
      <OfflineDemo />
    </div>
  );
}