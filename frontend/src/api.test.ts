import { beforeEach, describe, expect, it, vi } from "vitest";

import { createContent, listRevisions, restoreRevision, type CreateContentInput } from "./api";

class MemoryStorage implements Storage {
  readonly #items = new Map<string, string>();

  get length(): number {
    return this.#items.size;
  }

  clear(): void {
    this.#items.clear();
  }

  getItem(key: string): string | null {
    return this.#items.get(key) ?? null;
  }

  key(index: number): string | null {
    return Array.from(this.#items.keys())[index] ?? null;
  }

  removeItem(key: string): void {
    this.#items.delete(key);
  }

  setItem(key: string, value: string): void {
    this.#items.set(key, value);
  }
}

function mockCreateContentResponse(input: CreateContentInput): Response {
  return {
    ok: true,
    json: async () => ({
      id: "content-1",
      projectId: "project/with space",
      slug: "content-1",
      currentRevision: 1,
      ...input,
    }),
  } as Response;
}

describe("api createContent", () => {
  beforeEach(() => {
    Object.defineProperty(globalThis, "localStorage", {
      configurable: true,
      value: new MemoryStorage(),
    });
    vi.restoreAllMocks();
  });

  it("posts chapter content to the project content endpoint", async () => {
    const input: CreateContentInput = {
      kind: "chapter",
      title: "Chapter 1",
      bodyMarkdown: "",
      metadataJson: "{}",
      sortOrder: 1,
      reason: "created from workspace",
    };
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(mockCreateContentResponse(input));

    await createContent("project/with space", input);

    expect(fetchMock).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/projects/project%2Fwith%20space/content",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify(input),
      }),
    );
    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect((init.headers as Headers).get("Content-Type")).toBe("application/json");
  });

  it("posts story bible entry content to the project content endpoint", async () => {
    const input: CreateContentInput = {
      kind: "story_bible_entry",
      title: "Story bible entry 1",
      bodyMarkdown: "## Overview\n\n",
      metadataJson: "{}",
      sortOrder: 2,
      reason: "created from workspace",
    };
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(mockCreateContentResponse(input));

    await createContent("project-1", input);

    expect(fetchMock).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/projects/project-1/content",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify(input),
      }),
    );
  });

  it("lists revisions from the encoded content revisions endpoint", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => [],
    } as Response);

    await listRevisions("project/with space", "content/with space");

    expect(fetchMock).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/projects/project%2Fwith%20space/content/content%2Fwith%20space/revisions",
      expect.any(Object),
    );
  });

  it("posts restore requests to the encoded revision restore endpoint", async () => {
    const input = {
      expectedRevision: 4,
      reason: "restore checkpoint 2",
    };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => ({
        id: "content/with space",
        projectId: "project/with space",
        kind: "chapter",
        title: "Opening",
        slug: "opening",
        bodyMarkdown: "Earlier body",
        metadataJson: "{}",
        sortOrder: 1,
        currentRevision: 5,
      }),
    } as Response);

    await restoreRevision("project/with space", "content/with space", 2, input);

    expect(fetchMock).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/projects/project%2Fwith%20space/content/content%2Fwith%20space/revisions/2/restore",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify(input),
      }),
    );
    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect((init.headers as Headers).get("Content-Type")).toBe("application/json");
  });
});
