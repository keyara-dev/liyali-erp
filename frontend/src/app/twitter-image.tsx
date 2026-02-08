import { ImageResponse } from "next/og";

// Image metadata
export const alt = "Liyali Suite - Modern Business Operations Platform";
export const size = {
  width: 1200,
  height: 675,
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
            width: 100,
            height: 100,
            background: "white",
            borderRadius: 20,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            fontSize: 40,
            fontWeight: "bold",
            color: "#0c54e7",
            marginBottom: 32,
            boxShadow: "0 20px 60px rgba(0,0,0,0.3)",
          }}
        >
          L
        </div>

        {/* Title */}
        <div
          style={{
            fontSize: 56,
            fontWeight: 800,
            color: "white",
            marginBottom: 20,
            textAlign: "center",
            textShadow: "0 4px 12px rgba(0,0,0,0.2)",
          }}
        >
          Liyali Suite
        </div>

        {/* Subtitle */}
        <div
          style={{
            fontSize: 28,
            fontWeight: 400,
            color: "rgba(255,255,255,0.95)",
            textAlign: "center",
          }}
        >
          Streamline Your Business Operations
        </div>
      </div>
    </div>,
    {
      ...size,
    },
  );
}
