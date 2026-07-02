import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError, apiError, createContent, listRevisions, restoreRevision, type CreateContentInput } from "./api";

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

  it("includes server failure details when listing revisions fails", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response("content missing", { status: 404 }));

    await expect(listRevisions("project-1", "content-1")).rejects.toMatchObject({
      name: "ApiError",
      status: 404,
      code: "HTTP_404",
      message: "list revisions failed: 404: content missing",
    } satisfies Partial<ApiError>);
  });

  it("includes server failure details when restoring a revision fails", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ error: "expected revision conflict" }), { status: 409 }));

    await expect(restoreRevision("project-1", "content-1", 2, { expectedRevision: 4, reason: "restore" })).rejects.toMatchObject({
      name: "ApiError",
      status: 409,
      code: "HTTP_409",
      message: "restore revision failed: 409: expected revision conflict",
    } satisfies Partial<ApiError>);
  });

  it("parses JSON error bodies in apiError", async () => {
    const error = await apiError("restore revision", new Response(JSON.stringify({ error: "conflict" }), { status: 409 }));

    expect(error).toMatchObject({
      code: "HTTP_409",
      message: "restore revision failed: 409: conflict",
    } satisfies Partial<ApiError>);
  });

  it("falls back to raw JSON bodies without an error field", async () => {
    const error = await apiError("list revisions", new Response(JSON.stringify({ message: "missing" }), { status: 404 }));

    expect(error.message).toBe('list revisions failed: 404: {"message":"missing"}');
  });

  it("uses plain text error bodies in apiError", async () => {
    const error = await apiError("list revisions", new Response("plain failure", { status: 500 }));

    expect(error.message).toBe("list revisions failed: 500: plain failure");
  });

  it("truncates long plain text error bodies in apiError", async () => {
    const error = await apiError("list revisions", new Response("x".repeat(320), { status: 500 }));

    expect(error.message).toBe(`list revisions failed: 500: ${"x".repeat(300)}...`);
  });

  it("stores status code and response body on direct ApiError instances", () => {
    const error = new ApiError("restore revision", 409, JSON.stringify({ error: "conflict" }));

    expect(error).toMatchObject({
      name: "ApiError",
      status: 409,
      code: "HTTP_409",
      body: JSON.stringify({ error: "conflict" }),
      message: "restore revision failed: 409: conflict",
    } satisfies Partial<ApiError>);
  });

  it("omits details for empty error bodies in apiError", async () => {
    const error = await apiError("list revisions", new Response("", { status: 502 }));

    expect(error.message).toBe("list revisions failed: 502");
  });
});
