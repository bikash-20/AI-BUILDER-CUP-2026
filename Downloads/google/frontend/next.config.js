/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: "standalone",
  experimental: {
    serverActions: { allowedOrigins: ["*"] },
  },
  async rewrites() {
    const fraudUrl = process.env.FRAUD_SERVICE_URL || "http://fraud-service:8080";
    const riskUrl = process.env.RISK_SERVICE_URL || "http://risk-service:8080";
    const chatUrl = process.env.CHAT_SERVICE_URL || "http://chat-bff:8080";
    return [
      { source: "/api/score", destination: `${fraudUrl}/v1/score` },
      { source: "/api/risk/:path*", destination: `${riskUrl}/v1/risk/:path*` },
      { source: "/api/chat", destination: `${chatUrl}/v1/chat` },
    ];
  },
};

module.exports = nextConfig;
