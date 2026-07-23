"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { Document, Page, pdfjs } from "react-pdf";
import "react-pdf/dist/Page/AnnotationLayer.css";
import "react-pdf/dist/Page/TextLayer.css";
import { ResponsiveSheet } from "@/components/ui/responsive-sheet";
import { Button } from "@/components/ui/button";
import {
  ChevronLeft,
  ChevronRight,
  Download,
  X,
  ZoomIn,
  ZoomOut,
} from "lucide-react";
import Loader from "../ui/loader";

// Configure PDF.js worker
pdfjs.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjs.version}/build/pdf.worker.min.mjs`;

interface PDFPreviewDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  pdfBlob: Blob;
  fileName: string;
  onDownload: () => void;
}

export function PDFPreviewDialog({
  open,
  onOpenChange,
  pdfBlob,
  fileName,
  onDownload,
}: PDFPreviewDialogProps) {
  const [pdfUrl, setPdfUrl] = useState<string | null>(null);
  const [numPages, setNumPages] = useState<number>(0);
  const [pageNumber, setPageNumber] = useState<number>(1);
  const [scale, setScale] = useState<number>(1.5);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Use a ref to track the blob URL for reliable cleanup (avoids stale closures)
  const pdfUrlRef = useRef<string | null>(null);

  // Load PDF function - wrapped in useCallback for stable reference
  const loadPDF = useCallback(() => {
    if (!pdfBlob) {
      setLoadError("No PDF data available");
      return;
    }

    // Revoke any previous blob URL before creating a new one
    if (pdfUrlRef.current) {
      URL.revokeObjectURL(pdfUrlRef.current);
      pdfUrlRef.current = null;
      setPdfUrl(null);
    }

    setIsLoading(true);
    setLoadError(null);

    try {
      console.log("Creating PDF URL, blob size:", pdfBlob.size);

      if (pdfBlob.size === 0) {
        throw new Error("PDF blob is empty");
      }

      // Create blob URL for react-pdf
      const url = URL.createObjectURL(pdfBlob);
      console.log("PDF URL created:", url);
      pdfUrlRef.current = url;
      setPdfUrl(url);
    } catch (err) {
      console.error("PDF preview error:", err);
      setLoadError(
        err instanceof Error ? err.message : "Failed to load PDF preview",
      );
    } finally {
      setIsLoading(false);
    }
  }, [pdfBlob]);

  // Generate blob URL when dialog opens; clean up on close/unmount
  useEffect(() => {
    if (open && pdfBlob) {
      loadPDF();
    }

    return () => {
      // Revoke via ref to avoid stale closure over pdfUrl state
      if (pdfUrlRef.current) {
        URL.revokeObjectURL(pdfUrlRef.current);
        pdfUrlRef.current = null;
      }
      setPdfUrl(null);
      setLoadError(null);
      setNumPages(0);
      setPageNumber(1);
    };
  }, [open, pdfBlob, loadPDF]);

  const onDocumentLoadSuccess = ({ numPages }: { numPages: number }) => {
    setNumPages(numPages);
    setPageNumber(1);
    setLoadError(null);
    console.log(`PDF loaded successfully with ${numPages} pages`);
  };

  const onDocumentLoadError = (error: Error) => {
    console.error("PDF load error:", error);
    setLoadError(`Failed to load PDF: ${error.message}`);
  };

  const handlePrevPage = () => {
    setPageNumber((prev) => Math.max(1, prev - 1));
  };

  const handleNextPage = () => {
    setPageNumber((prev) => Math.min(numPages, prev + 1));
  };

  const zoomIn = () => {
    setScale((prev) => Math.min(prev + 0.2, 2.0));
  };

  const zoomOut = () => {
    setScale((prev) => Math.max(prev - 0.2, 0.5));
  };

  const toolbar = pdfUrl && numPages > 0 ? (
    <div className="flex flex-wrap items-center justify-between gap-2 w-full">
      <div className="flex items-center gap-2">
        {/* Zoom Controls */}
        <div className="flex items-center gap-1 rounded-lg border bg-background">
          <Button
            variant="ghost"
            size="icon"
            onClick={zoomOut}
            disabled={scale <= 0.5}
            className="h-8 w-8 rounded-r-none"
          >
            <ZoomOut className="h-4 w-4" />
          </Button>
          <span className="px-2 text-sm text-muted-foreground">
            {Math.round(scale * 100)}%
          </span>
          <Button
            variant="ghost"
            size="icon"
            onClick={zoomIn}
            disabled={scale >= 2.0}
            className="h-8 w-8 rounded-l-none"
          >
            <ZoomIn className="h-4 w-4" />
          </Button>
        </div>

        {/* Page Navigation */}
        <div className="flex items-center gap-1 rounded-lg border">
          <Button
            variant="ghost"
            size="icon"
            onClick={handlePrevPage}
            disabled={pageNumber <= 1}
            className="h-8 w-8 rounded-r-none"
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="px-2 text-sm text-muted-foreground">
            {pageNumber} / {numPages}
          </span>
          <Button
            variant="ghost"
            size="icon"
            onClick={handleNextPage}
            disabled={pageNumber >= numPages}
            className="h-8 w-8 rounded-l-none"
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>

        {/* Download Button */}
        <Button
          size="sm"
          className="gap-2"
          onClick={() => {
            onDownload();
            onOpenChange(false);
          }}
        >
          <Download className="h-4 w-4" />
          Download
        </Button>
      </div>
    </div>
  ) : null;

  return (
    <ResponsiveSheet
      open={open}
      onOpenChange={onOpenChange}
      fullScreen
      title={fileName}
      footer={toolbar}
    >
      {/* PDF viewport */}
      <div className="h-full overflow-auto bg-muted -mx-6 -mb-4 px-0">
        {isLoading && (
          <div className="flex h-full items-center justify-center">
            <div className="text-muted-foreground">Loading PDF...</div>
          </div>
        )}

        {loadError && !isLoading && (
          <div className="flex h-full items-center justify-center">
            <div className="max-w-md rounded-lg bg-destructive/10 p-6 text-center">
              <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-destructive/20">
                <X className="h-6 w-6 text-destructive" />
              </div>
              <h3 className="mb-2 text-lg font-semibold text-destructive">
                Preview Failed
              </h3>
              <p className="text-sm text-destructive/80">{loadError}</p>
              <Button
                onClick={loadPDF}
                variant="destructive"
                size="sm"
                className="mt-4"
              >
                Try Again
              </Button>
            </div>
          </div>
        )}

        {pdfUrl && !isLoading && !loadError && (
          <div className="flex justify-center p-4">
            <Document
              file={pdfUrl}
              onLoadSuccess={onDocumentLoadSuccess}
              onLoadError={onDocumentLoadError}
              loading={
                <div className="flex h-96 items-center justify-center">
                  <Loader loadingText="Loading Preview..." />
                </div>
              }
            >
              <Page
                pageNumber={pageNumber}
                scale={scale}
                renderTextLayer={true}
                renderAnnotationLayer={true}
                className="shadow-lg"
              />
            </Document>
          </div>
        )}
      </div>
    </ResponsiveSheet>
  );
}
