import { Activity, CheckCircle2, FileText } from "lucide-react";

import { Button } from "../../shared/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../shared/ui/tabs";

export function ReviewDrawer() {
  return (
    <aside className="flex h-full flex-col gap-4" aria-label="Review">
      <header className="flex flex-col gap-1">
        <h2 className="text-base font-semibold text-foreground">Review</h2>
        <p className="text-sm text-muted-foreground">Checks and revisions are not connected yet.</p>
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

        <TabsContent value="activity">
          <div className="flex items-center gap-2 rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
            <Activity className="size-4" aria-hidden="true" />
            No review activity yet.
          </div>
        </TabsContent>
      </Tabs>
    </aside>
  );
}
