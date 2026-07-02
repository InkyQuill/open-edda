import { Activity, AlertCircle, FileText, GitCompareArrows, MessageSquareText, RotateCcw } from "lucide-react";
import { useEffect, useMemo } from "react";
import { useDispatch, useSelector } from "react-redux";

import type { AppDispatch, RootState } from "../../app/store/store";
import type { ActivityEvent, PromptRecord } from "../../agentTypes";
import type { ContentItem, Revision } from "../../types";
import { Button } from "../../shared/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../shared/ui/tabs";
import { reviewActions } from "./reviewSlice";
import { loadContentRevisions, loadPromptRecords, loadReviewActivity, restoreContentRevision } from "./reviewThunks";

type ReviewDrawerProps = {
  projectId: string;
  content: ContentItem | null;
  onContentSaved: (item: ContentItem) => void;
};

function formatDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Unknown time";
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function formatActionKind(value: string): string {
  return value
    .split("_")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function formatJSON(value: string): string {
  try {
    return JSON.stringify(JSON.parse(value), null, 2);
  } catch {
    return value || "{}";
  }
}

function restoreConflictMessage(error: string | null, code: string | null): string | null {
  if (!error) return null;
  if (code === "HTTP_409" || /\b409\b|conflict/i.test(error)) {
    return "Content changed before this checkpoint was restored. Review the latest draft, then try again.";
  }
  return error;
}

function diffLines(current: string, selected: string): Array<{ kind: "same" | "added" | "removed"; text: string }> {
  const currentLines = current.split("\n");
  const selectedLines = selected.split("\n");
  const lengths = Array.from({ length: currentLines.length + 1 }, () => Array<number>(selectedLines.length + 1).fill(0));

  for (let currentIndex = currentLines.length - 1; currentIndex >= 0; currentIndex -= 1) {
    for (let selectedIndex = selectedLines.length - 1; selectedIndex >= 0; selectedIndex -= 1) {
      lengths[currentIndex][selectedIndex] =
        currentLines[currentIndex] === selectedLines[selectedIndex]
          ? lengths[currentIndex + 1][selectedIndex + 1] + 1
          : Math.max(lengths[currentIndex + 1][selectedIndex], lengths[currentIndex][selectedIndex + 1]);
    }
  }

  const lines: Array<{ kind: "same" | "added" | "removed"; text: string }> = [];
  let currentIndex = 0;
  let selectedIndex = 0;

  while (currentIndex < currentLines.length && selectedIndex < selectedLines.length) {
    if (currentLines[currentIndex] === selectedLines[selectedIndex]) {
      lines.push({ kind: "same", text: currentLines[currentIndex] });
      currentIndex += 1;
      selectedIndex += 1;
      continue;
    }

    if (lengths[currentIndex + 1][selectedIndex] >= lengths[currentIndex][selectedIndex + 1]) {
      lines.push({ kind: "removed", text: currentLines[currentIndex] });
      currentIndex += 1;
    } else {
      lines.push({ kind: "added", text: selectedLines[selectedIndex] });
      selectedIndex += 1;
    }
  }
  while (currentIndex < currentLines.length) {
    lines.push({ kind: "removed", text: currentLines[currentIndex] });
    currentIndex += 1;
  }
  while (selectedIndex < selectedLines.length) {
    lines.push({ kind: "added", text: selectedLines[selectedIndex] });
    selectedIndex += 1;
  }

  return lines;
}

function ActivityEventItem({ event }: { event: ActivityEvent }) {
  return (
    <article className="flex flex-col gap-1 rounded-md border border-border bg-background p-3 text-sm">
      <div className="flex items-start justify-between gap-3">
        <p className="min-w-0 font-medium text-foreground">{event.summary}</p>
        <time dateTime={event.createdAt} className="shrink-0 text-xs text-muted-foreground">
          {formatDateTime(event.createdAt)}
        </time>
      </div>
      <p className="text-xs uppercase text-muted-foreground">{event.eventType}</p>
    </article>
  );
}

function PromptRecordItem({
  promptRecord,
  selected,
  onSelect,
}: {
  promptRecord: PromptRecord;
  selected: boolean;
  onSelect: () => void;
}) {
  return (
    <button
      type="button"
      className="flex w-full flex-col gap-2 rounded-md border border-border bg-background p-3 text-left text-sm hover:bg-muted/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring aria-pressed:border-primary"
      aria-pressed={selected}
      onClick={onSelect}
    >
      <span className="flex items-start justify-between gap-3">
        <span className="min-w-0 font-medium text-foreground">
          {formatActionKind(promptRecord.actionKind)} with {promptRecord.modelName}
        </span>
        <time dateTime={promptRecord.createdAt} className="shrink-0 text-xs text-muted-foreground">
          {formatDateTime(promptRecord.createdAt)}
        </time>
      </span>
      <span className="flex flex-wrap gap-2 text-xs text-muted-foreground">
        <span>{promptRecord.providerName}</span>
        <span>{promptRecord.usage.totalTokens.toLocaleString()} tokens</span>
        <span>${promptRecord.usage.totalCost.toFixed(4)}</span>
      </span>
    </button>
  );
}

function RevisionItem({
  revision,
  selected,
  currentRevision,
  onSelect,
}: {
  revision: Revision;
  selected: boolean;
  currentRevision: number;
  onSelect: () => void;
}) {
  return (
    <button
      type="button"
      className="flex w-full flex-col gap-1 rounded-md border border-border bg-background p-3 text-left text-sm hover:bg-muted/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring aria-pressed:border-primary"
      aria-pressed={selected}
      onClick={onSelect}
    >
      <span className="flex items-start justify-between gap-3">
        <span className="font-medium text-foreground">
          Checkpoint {revision.revisionNumber}
          {revision.revisionNumber === currentRevision ? " (current)" : ""}
        </span>
        <time dateTime={revision.createdAt} className="shrink-0 text-xs text-muted-foreground">
          {formatDateTime(revision.createdAt)}
        </time>
      </span>
      <span className="text-xs text-muted-foreground">
        {revision.reason || "No reason recorded"} · {revision.createdBy}
        {revision.actionKind ? ` · ${formatActionKind(revision.actionKind)}` : ""}
      </span>
    </button>
  );
}

export function ReviewDrawer({ projectId, content, onContentSaved }: ReviewDrawerProps) {
  const dispatch = useDispatch<AppDispatch>();
  const {
    activityEvents,
    activityStatus,
    contentId,
    error,
    projectId: reviewProjectId,
    promptRecords,
    promptRecordsStatus,
    revisionsError,
    restoreError,
    restoreErrorCode,
    restoreStatus,
    revisions,
    revisionsStatus,
    selectedPromptRecordId,
    selectedRevisionNumber,
  } = useSelector((state: RootState) => state.review);
  const checkResult = useSelector((state: RootState) => state.assistantActions.checkResult);
  const isCurrentProject = reviewProjectId === projectId;
  const isCurrentContent = isCurrentProject && contentId === content?.id;
  const visibleError = isCurrentProject ? error : null;
  const visibleActivityEvents = isCurrentProject ? activityEvents : [];
  const visiblePromptRecords = isCurrentProject ? promptRecords : [];
  const visibleRevisions = isCurrentContent ? revisions : [];
  const visibleRevisionsError = isCurrentContent ? revisionsError : null;
  const visibleRestoreError = isCurrentContent ? restoreConflictMessage(restoreError, restoreErrorCode) : null;
  const visibleCheckResult = checkResult?.note.contentItemId === content?.id ? checkResult : null;
  const selectedPromptRecord = visiblePromptRecords.find((record) => record.id === selectedPromptRecordId) ?? null;
  const selectedRevision =
    visibleRevisions.find((revision) => revision.revisionNumber === selectedRevisionNumber) ?? visibleRevisions[0] ?? null;
  const loading = activityStatus === "pending" || promptRecordsStatus === "pending";
  const hasActivity = visibleActivityEvents.length > 0;
  const hasPromptRecords = visiblePromptRecords.length > 0;
  const isEmpty = !visibleError && !loading && !hasActivity && !hasPromptRecords;
  const visibleRevisionsStatus = isCurrentContent ? revisionsStatus : "idle";
  const revisionsLoading = visibleRevisionsStatus === "pending";
  const hasRevisions = visibleRevisions.length > 0;
  const restorePending = restoreStatus === "pending";
  const loadedThroughRevision = Math.max(0, ...visibleRevisions.map((revision) => revision.revisionNumber));

  useEffect(() => {
    if (reviewProjectId !== projectId || activityStatus === "idle") {
      void dispatch(loadReviewActivity({ projectId }));
    }
  }, [activityStatus, dispatch, projectId, reviewProjectId]);

  useEffect(() => {
    if (reviewProjectId !== projectId || promptRecordsStatus === "idle") {
      void dispatch(loadPromptRecords({ projectId }));
    }
  }, [dispatch, projectId, promptRecordsStatus, reviewProjectId]);

  useEffect(() => {
    if (!content) return;
    if (
      reviewProjectId !== projectId ||
      contentId !== content.id ||
      revisionsStatus === "idle" ||
      (revisionsStatus === "succeeded" && loadedThroughRevision < content.currentRevision)
    ) {
      void dispatch(loadContentRevisions({ projectId, contentId: content.id }));
    }
  }, [content, contentId, dispatch, loadedThroughRevision, projectId, reviewProjectId, revisionsStatus]);

  function handleRestore(): void {
    if (!content || !selectedRevision || selectedRevision.revisionNumber === content.currentRevision || restorePending) {
      return;
    }

    void dispatch(
      restoreContentRevision({
        projectId,
        contentId: content.id,
        revisionNumber: selectedRevision.revisionNumber,
        expectedRevision: content.currentRevision,
      }),
    )
      .unwrap()
      .then((item) => {
        onContentSaved(item);
        void dispatch(loadContentRevisions({ projectId, contentId: item.id }));
      })
      .catch(() => undefined);
  }

  function handleRetryRevisions(): void {
    if (!content || revisionsLoading) return;
    void dispatch(loadContentRevisions({ projectId, contentId: content.id }));
  }

  const restoreDisabled =
    !content || !selectedRevision || selectedRevision.revisionNumber === content.currentRevision || restorePending;
  const diff = useMemo(
    () => (content && selectedRevision ? diffLines(content.bodyMarkdown, selectedRevision.bodyMarkdown) : []),
    [content, selectedRevision],
  );

  return (
    <aside className="flex h-full flex-col gap-4" aria-label="Review">
      <header className="flex flex-col gap-1">
        <h2 className="text-base font-semibold text-foreground">Review</h2>
        <p className="text-sm text-muted-foreground">
          {loading ? "Loading review history..." : "Review history and prompt records."}
        </p>
      </header>

      <Tabs defaultValue="reports" className="min-h-0 flex-1">
        <TabsList className="w-full">
          <TabsTrigger value="reports">Reports</TabsTrigger>
          <TabsTrigger value="revisions">File history</TabsTrigger>
          <TabsTrigger value="activity">Activity</TabsTrigger>
        </TabsList>

        <TabsContent value="reports" forceMount className="flex flex-col gap-3">
          <Button type="button" variant="outline" className="w-full justify-start" disabled>
            <FileText />
            Read report
          </Button>
          {visibleCheckResult ? (
            <article className="rounded-md border border-border bg-background p-3">
              <h3 className="text-sm font-medium text-foreground">{visibleCheckResult.note.title}</h3>
              <p className="mt-2 whitespace-pre-wrap text-sm leading-6 text-foreground">
                {visibleCheckResult.assistantMessage.bodyMarkdown}
              </p>
              <p className="mt-2 text-xs text-muted-foreground">
                Attached to bytes {visibleCheckResult.note.selectionStart}-{visibleCheckResult.note.selectionEnd}
              </p>
            </article>
          ) : (
            <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
              Run Check from the editor selection toolbar to create a review report for the current content.
            </p>
          )}
        </TabsContent>

        <TabsContent value="revisions" forceMount className="min-h-0">
          {!content ? (
            <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
              Select content to review checkpoints.
            </p>
          ) : (
            <div className="flex max-h-[calc(100dvh-13rem)] flex-col gap-4 overflow-auto pr-1">
              <section className="rounded-md border border-border bg-background p-3 text-sm">
                <p className="font-medium text-foreground">Current checkpoint {content.currentRevision}</p>
                <p className="mt-1 text-xs text-muted-foreground">{content.title}</p>
              </section>

              {revisionsLoading ? (
                <div className="flex items-center gap-2 rounded-md border border-border bg-muted/30 p-3 text-sm text-muted-foreground">
                  <Activity className="size-4 animate-pulse" aria-hidden="true" />
                  Loading checkpoints...
                </div>
              ) : null}

              {visibleRevisionsError || visibleRestoreError ? (
                <div
                  role="alert"
                  className="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive"
                >
                  <div className="flex items-start gap-2">
                    <AlertCircle className="size-4 shrink-0" aria-hidden="true" />
                    <span>{visibleRestoreError ?? visibleRevisionsError}</span>
                  </div>
                  {visibleRevisionsError ? (
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="mt-3"
                      disabled={revisionsLoading}
                      onClick={handleRetryRevisions}
                    >
                      Retry
                    </Button>
                  ) : null}
                </div>
              ) : null}

              {!revisionsLoading && !hasRevisions ? (
                <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
                  No checkpoints found for this content.
                </p>
              ) : null}

              {hasRevisions ? (
                <section className="flex flex-col gap-2" aria-label="Checkpoints">
                  {visibleRevisions.map((revision) => (
                    <RevisionItem
                      key={revision.id}
                      revision={revision}
                      currentRevision={content.currentRevision}
                      selected={revision.revisionNumber === selectedRevision?.revisionNumber}
                      onSelect={() => dispatch(reviewActions.setSelectedRevisionNumber(revision.revisionNumber))}
                    />
                  ))}
                </section>
              ) : null}

              {selectedRevision ? (
                <section className="flex flex-col gap-3 rounded-md border border-border bg-background p-3">
                  <div className="flex items-start justify-between gap-3">
                    <h3 className="flex items-center gap-2 text-sm font-medium text-foreground">
                      <GitCompareArrows className="size-4" aria-hidden="true" />
                      Checkpoint {selectedRevision.revisionNumber} diff
                    </h3>
                    <Button type="button" size="sm" disabled={restoreDisabled} onClick={handleRestore}>
                      <RotateCcw data-icon="inline-start" aria-hidden="true" />
                      {restorePending ? "Restoring..." : "Restore"}
                    </Button>
                  </div>
                  <pre className="max-h-80 overflow-auto rounded-md bg-muted/40 p-3 text-xs leading-5 text-foreground">
                    {diff.map((line, index) => {
                      const prefix = line.kind === "added" ? "+ " : line.kind === "removed" ? "- " : "  ";
                      return (
                        <code
                          key={`${index}-${line.kind}`}
                          className={
                            line.kind === "added"
                              ? "block text-emerald-700"
                              : line.kind === "removed"
                                ? "block text-destructive"
                                : "block text-muted-foreground"
                          }
                        >
                          {prefix}
                          {line.text || " "}
                        </code>
                      );
                    })}
                  </pre>
                </section>
              ) : null}
            </div>
          )}
        </TabsContent>

        <TabsContent value="activity" forceMount className="min-h-0">
          <div className="flex max-h-[calc(100dvh-13rem)] flex-col gap-4 overflow-auto pr-1">
            {loading ? (
              <div className="flex items-center gap-2 rounded-md border border-border bg-muted/30 p-3 text-sm text-muted-foreground">
                <Activity className="size-4 animate-pulse" aria-hidden="true" />
                Loading activity...
              </div>
            ) : null}

            {visibleError ? (
              <p
                role="alert"
                className="flex items-start gap-2 rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive"
              >
                <AlertCircle className="size-4 shrink-0" aria-hidden="true" />
                {visibleError}
              </p>
            ) : null}

            {isEmpty ? (
              <div className="flex items-center gap-2 rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
                <Activity className="size-4" aria-hidden="true" />
                No review activity yet.
              </div>
            ) : null}

            {hasActivity ? (
              <section className="flex flex-col gap-2" aria-labelledby="review-activity-title">
                <h3 id="review-activity-title" className="text-sm font-medium text-foreground">
                  Activity
                </h3>
                <div className="flex flex-col gap-2">
                  {visibleActivityEvents.map((event) => (
                    <ActivityEventItem key={event.id} event={event} />
                  ))}
                </div>
              </section>
            ) : null}

            {hasPromptRecords ? (
              <section className="flex flex-col gap-2" aria-labelledby="review-prompts-title">
                <h3 id="review-prompts-title" className="flex items-center gap-2 text-sm font-medium text-foreground">
                  <MessageSquareText className="size-4" aria-hidden="true" />
                  Prompt records
                </h3>
                <div className="flex flex-col gap-2">
                  {visiblePromptRecords.map((promptRecord) => (
                    <PromptRecordItem
                      key={promptRecord.id}
                      promptRecord={promptRecord}
                      selected={promptRecord.id === selectedPromptRecordId}
                      onSelect={() => dispatch(reviewActions.setSelectedPromptRecordId(promptRecord.id))}
                    />
                  ))}
                </div>
              </section>
            ) : null}

            {selectedPromptRecord ? (
              <section className="flex flex-col gap-2 rounded-md border border-border bg-background p-3" aria-label="Prompt record details">
                <h3 className="text-sm font-medium text-foreground">Prompt record details</h3>
                <div className="flex flex-col gap-2">
                  <details open>
                    <summary className="cursor-pointer text-xs font-medium text-muted-foreground">Request</summary>
                    <pre className="mt-2 max-h-52 overflow-auto rounded-md bg-muted/40 p-2 text-xs text-foreground">
                      {formatJSON(selectedPromptRecord.requestJson)}
                    </pre>
                  </details>
                  <details>
                    <summary className="cursor-pointer text-xs font-medium text-muted-foreground">Response</summary>
                    <pre className="mt-2 max-h-52 overflow-auto rounded-md bg-muted/40 p-2 text-xs text-foreground">
                      {formatJSON(selectedPromptRecord.responseJson)}
                    </pre>
                  </details>
                </div>
              </section>
            ) : null}
          </div>
        </TabsContent>
      </Tabs>
    </aside>
  );
}
