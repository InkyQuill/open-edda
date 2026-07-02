import { expect, test, type Page } from "@playwright/test";

const consoleErrorsByPage = new WeakMap<Page, string[]>();

const project = {
  id: "project-1",
  title: "Alchemy Draft",
  slug: "alchemy-draft",
  language: "en",
};

const chapter = {
  id: "content-1",
  projectId: "project-1",
  kind: "chapter",
  title: "Opening",
  slug: "opening",
  bodyMarkdown: "The lantern burned blue.",
  metadataJson: "{}",
  sortOrder: 1,
  currentRevision: 1,
};

const revision = {
  id: "revision-1",
  contentItemId: "content-1",
  revisionNumber: 1,
  bodyMarkdown: "The lantern burned blue.",
  metadataJson: "{}",
  reason: "initial content",
  createdBy: "author",
  createdAt: "2026-07-01T00:00:00.000Z",
};

async function mockOpenEddaApi(page: Page): Promise<void> {
  await page.addInitScript(() => {
    window.localStorage.setItem("open_edda_token", "smoke-token");
  });

  await page.route("**/api/**", async (route) => {
    const url = new URL(route.request().url());
    const path = url.pathname;
    let body: unknown = [];

    if (path === "/api/projects") {
      body = [project];
    } else if (path === "/api/provider-configs") {
      body = [];
    } else if (path === "/api/projects/project-1/content") {
      body = url.searchParams.get("kind") === "chapter" ? [chapter] : [];
    } else if (path === "/api/projects/project-1/agent/sessions") {
      body = [];
    } else if (path === "/api/projects/project-1/agent/activity") {
      body = [];
    } else if (path === "/api/projects/project-1/agent/prompt-records") {
      body = [];
    } else if (path === "/api/projects/project-1/content/content-1/revisions") {
      body = [revision];
    } else {
      await route.fulfill({ status: 404, body: `unmocked route: ${path}` });
      return;
    }

    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(body),
    });
  });
}

test.describe("workspace smoke", () => {
  test.beforeEach(async ({ page }) => {
    const consoleErrors: string[] = [];
    consoleErrorsByPage.set(page, consoleErrors);
    page.on("console", (message) => {
      if (message.type() === "error") {
        consoleErrors.push(message.text());
      }
    });
    page.on("pageerror", (error) => {
      consoleErrors.push(error.message);
    });
    await mockOpenEddaApi(page);
    await page.goto("/projects/project-1/content/chapter/content-1");
    await expect(page.getByRole("heading", { name: "Alchemy Draft" })).toBeVisible();
  });

  test("desktop mode switching keeps the workspace usable", async ({ page }, testInfo) => {
    test.skip(testInfo.project.name !== "desktop", "desktop-only smoke");

    await page.getByRole("button", { name: "Draft" }).first().click();
    await expect(page.getByText("Opening").first()).toBeVisible();
    await expect(page.getByText("Select a model before starting assistant chat.")).toBeHidden();

    await page.getByRole("button", { name: "Review" }).first().click();
    await expect(page.getByRole("heading", { name: "Review" })).toBeVisible();
    await expect(page.getByText("Current checkpoint 1")).toBeVisible();

    await page.getByRole("button", { name: "Assistant" }).first().click();
    await expect(page.getByRole("heading", { name: "Assistant" })).toBeVisible();
    await expect(page.getByText("Select a model before starting assistant chat.")).toBeVisible();
  });

  test("mobile sheets open and close over the editor", async ({ page }, testInfo) => {
    test.skip(testInfo.project.name !== "mobile", "mobile-only smoke");

    for (const [buttonName, title] of [
      ["Files", "Files"],
      ["Assistant", "Assistant"],
      ["Review", "Review"],
      ["World/Notes", "World and notes"],
    ] as const) {
      await page.getByRole("button", { name: buttonName }).click();
      await expect(page.getByRole("dialog")).toBeVisible();
      await expect(page.getByText(title).last()).toBeVisible();
      await page.getByRole("button", { name: "Close" }).click();
      await expect(page.getByRole("dialog")).toBeHidden();
    }
  });

  test.afterEach(async ({ page }) => {
    const errors = consoleErrorsByPage.get(page) ?? [];
    expect(errors).toEqual([]);
  });
});
