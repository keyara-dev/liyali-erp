import { ImageResponse } from "next/og";

// Image metadata
export const alt = "Liyali Suite - Modern Business Operations Platform";
export const size = {
  width: 1200,
  height: 630,
};

export const contentType = "image/png";

// Image generation
export default async function Image() {
  return new ImageResponse(
    <div
      style={{
        background: "linear-gradient(135deg, #0c54e7 0%, #0a3fb8 100%)",
        width: "100%",
        height: "100%",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        fontFamily: "system-ui, sans-serif",
        position: "relative",
      }}
    >
      {/* Background decoration */}
      <div
        style={{
          position: "absolute",
          top: "-50%",
          right: "-50%",
          width: "100%",
          height: "100%",
          background:
            "radial-gradient(circle, rgba(255,255,255,0.1) 0%, transparent 70%)",
        }}
      />

      {/* Content */}
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          zIndex: 1,
        }}
      >
        {/* Logo */}
        <div
          style={{
            width: 120,
            height: 120,
            background: "white",
            borderRadius: 24,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            fontSize: 48,
            fontWeight: "bold",
            color: "#0c54e7",
            marginBottom: 40,
            boxShadow: "0 20px 60px rgba(0,0,0,0.3)",
          }}
        >
          L
        </div>

        {/* Title */}
        <div
          style={{
            fontSize: 64,
            fontWeight: 800,
            color: "white",
            marginBottom: 24,
            textAlign: "center",
            textShadow: "0 4px 12px rgba(0,0,0,0.2)",
          }}
        >
          Liyali Suite
        </div>

        {/* Subtitle */}
        <div
          style={{
            fontSize: 32,
            fontWeight: 400,
            color: "rgba(255,255,255,0.95)",
            marginBottom: 40,
            textAlign: "center",
          }}
        >
          Modern Business Operations Platform
        </div>

        {/* Badge */}
        <div
          style={{
            background: "rgba(255,255,255,0.2)",
            padding: "12px 32px",
            borderRadius: 50,
            fontSize: 18,
            color: "white",
            fontWeight: 600,
            border: "2px solid rgba(255,255,255,0.3)",
          }}
        >
          Trusted by 500+ Organizations
        </div>
      </div>
    </div>,
    {
      ...size,
    },
  );
}
