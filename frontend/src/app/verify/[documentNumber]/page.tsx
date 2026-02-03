import { Metadata } from "next";
import { verifyDocument } from "@/app/_actions/verification";
import { VerificationResult } from "./_components/verification-result";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Document Verification | Liyali",
  description: "Verify the authenticity of documents using QR code verification",
};

interface VerifyPageProps {
  params: Promise<{
    documentNumber: string;
  }>;
}

export default async function VerifyPage({ params }: VerifyPageProps) {
  const { documentNumber } = await params;
  const decodedDocumentNumber = decodeURIComponent(documentNumber);

  // Fetch verification result on the server
  const result = await verifyDocument(decodedDocumentNumber);

  return (
    <div className="min-h-screen bg-gradient-to-b from-background to-muted/20">
      <VerificationResult
        documentNumber={decodedDocumentNumber}
        result={result}
      />
    </div>
  );
}
