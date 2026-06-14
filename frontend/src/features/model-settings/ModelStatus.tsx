import { AlertCircle, Settings2 } from "lucide-react";

import { Button } from "../../shared/ui/button";

export function ModelStatus() {
  return (
    <section className="flex flex-col gap-3" aria-labelledby="model-status-title">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h3 id="model-status-title" className="text-sm font-medium text-foreground">
            Model
          </h3>
          <p className="text-xs text-muted-foreground">Selection pending settings sync.</p>
        </div>
        <Button type="button" variant="outline" size="icon-sm" aria-label="Open model settings">
          <Settings2 />
        </Button>
      </div>

      <div className="flex items-start gap-2 rounded-md border border-border bg-muted/40 p-3 text-sm">
        <AlertCircle className="mt-0.5 size-4 text-muted-foreground" aria-hidden="true" />
        <div>
          <p className="font-medium text-foreground">Selected model unavailable</p>
          <p className="text-xs text-muted-foreground">
            Choose a model after settings wiring lands.
          </p>
        </div>
      </div>
    </section>
  );
}
