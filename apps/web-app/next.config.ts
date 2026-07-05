import path from "node:path";

import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  outputFileTracingRoot: path.join(__dirname, "../.."),
  transpilePackages: [
    "@code-cleaner/ui-components",
    "@code-cleaner/utils",
  ],
};

export default nextConfig;
