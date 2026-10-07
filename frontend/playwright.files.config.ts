import { defineConfig, devices } from "@playwright/test";
export default defineConfig({
  testDir: "./.playwright/files", outputDir: "./.playwright/test-results/files",
  timeout: 45_000, reporter: "list", workers: 1,
  use: { baseURL: process.env.EDDA_TEST_BASE_URL ?? "http://127.0.0.1:4187", trace: "retain-on-failure" },
  webServer: process.env.EDDA_TEST_BASE_URL ? undefined : { command: "python3 ../scripts/run-file-workspace-smoke.py", url: "http://127.0.0.1:4187/api/health", timeout: 60_000, reuseExistingServer: false },
  projects: [
    { name: "desktop", use: { ...devices["Desktop Chrome"] } },
    { name: "mobile", use: { ...devices["Pixel 5"] } },
  ],
});
