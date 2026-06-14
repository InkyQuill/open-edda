import { WandSparkles } from "lucide-react";

import { Button } from "../../shared/ui/button";
import { ModelStatus } from "../model-settings/ModelStatus";
import { SkillChipsPanel } from "../skills/SkillChipsPanel";

export function AssistantDrawer() {
  return (
    <aside className="flex h-full flex-col gap-5" aria-label="Assistant">
      <header className="flex flex-col gap-1">
        <h2 className="text-base font-semibold text-foreground">Assistant</h2>
        <p className="text-sm text-muted-foreground">Draft-aware help will appear here.</p>
      </header>

      <section className="flex flex-col gap-3" aria-labelledby="assistant-transcript-title">
        <h3 id="assistant-transcript-title" className="text-sm font-medium text-foreground">
          Transcript
        </h3>
        <div className="rounded-md border border-border bg-background p-3 text-sm text-muted-foreground">
          No assistant messages yet.
        </div>
      </section>

      <section className="flex flex-col gap-3" aria-labelledby="quick-actions-title">
        <div className="flex items-center justify-between gap-3">
          <h3 id="quick-actions-title" className="text-sm font-medium text-foreground">
            Quick actions
          </h3>
          <Button type="button" variant="outline" size="xs">
            <WandSparkles />
            Ready
          </Button>
        </div>
        <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
          Selection actions will attach to the editor in the next task.
        </p>
      </section>

      <SkillChipsPanel />
      <ModelStatus />
    </aside>
  );
}
