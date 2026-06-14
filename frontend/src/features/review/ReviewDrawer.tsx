import { Activity, AlertCircle, CheckCircle2, FileText, MessageSquareText } from "lucide-react";
import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";

import type { AppDispatch, RootState } from "../../app/store/store";
import type { ActivityEvent, PromptRecord } from "../../agentTypes";
import { Button } from "../../shared/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../shared/ui/tabs";
import { reviewActions } from "./reviewSlice";
import { loadPromptRecords, loadReviewActivity } from "./reviewThunks";

type ReviewDrawerProps = {
  projectId: string;
};

function formatDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value || "Unknown time";
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

export function ReviewDrawer({ projectId }: ReviewDrawerProps) {
  const dispatch = useDispatch<AppDispatch>();
  const {
    activityEvents,
    activityStatus,
    error,
    projectId: reviewProjectId,
    promptRecords,
    promptRecordsStatus,
    selectedPromptRecordId,
  } = useSelector((state: RootState) => state.review);
  const isCurrentProject = reviewProjectId === projectId;
  const visibleError = isCurrentProject ? error : null;
  const visibleActivityEvents = isCurrentProject ? activityEvents : [];
  const visiblePromptRecords = isCurrentProject ? promptRecords : [];
  const loading = activityStatus === "pending" || promptRecordsStatus === "pending";
  const hasActivity = visibleActivityEvents.length > 0;
  const hasPromptRecords = visiblePromptRecords.length > 0;
  const isEmpty = !visibleError && !loading && !hasActivity && !hasPromptRecords;

  useEffect(() => {
    dispatch(reviewActions.resetForProject());
    void dispatch(loadReviewActivity({ projectId }));
    void dispatch(loadPromptRecords({ projectId }));
  }, [dispatch, projectId]);

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
          <TabsTrigger value="revisions">Revisions</TabsTrigger>
          <TabsTrigger value="activity">Activity</TabsTrigger>
        </TabsList>

        <TabsContent value="reports" className="flex flex-col gap-3">
          <Button type="button" variant="outline" className="w-full justify-start">
            <FileText />
            Read report
          </Button>
          <Button type="button" variant="outline" className="w-full justify-start">
            <CheckCircle2 />
            Check draft
          </Button>
          <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
            Report output will appear here.
          </p>
        </TabsContent>

        <TabsContent value="revisions">
          <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
            Revision suggestions will appear here.
          </p>
        </TabsContent>

        <TabsContent value="activity" className="min-h-0">
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
          </div>
        </TabsContent>
      </Tabs>
    </aside>
  );
}
