import { BookOpen, FileText, StickyNote } from "lucide-react";

import { Button } from "../../shared/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../shared/ui/tabs";
import type { ContentItem, ContentKind } from "../../types";

type ContextDrawerProps = {
  activeTab: "contents" | "world" | "notes";
  contentItems: ContentItem[];
  contentLoading: boolean;
  contentError: string | null;
  activeContentKind: ContentKind;
  selectedContentId: string | null;
  onSelectContent: (item: ContentItem) => void;
  onContentKindChange: (kind: ContentKind) => void;
  onTabChange: (tab: "contents" | "world" | "notes") => void;
};

function isContextTab(value: string): value is ContextDrawerProps["activeTab"] {
  return value === "contents" || value === "world" || value === "notes";
}

const contentKindOptions: Array<{ kind: ContentKind; label: string }> = [
  { kind: "chapter", label: "Chapters" },
  { kind: "story_bible_entry", label: "World" },
  { kind: "writing_brief", label: "Briefs" },
  { kind: "project_note", label: "Notes" },
];

export function ContextDrawer({
  activeTab,
  contentItems,
  contentLoading,
  contentError,
  activeContentKind,
  selectedContentId,
  onSelectContent,
  onContentKindChange,
  onTabChange,
}: ContextDrawerProps) {
  return (
    <aside className="flex h-full flex-col gap-4" aria-label="Workspace context">
      <Tabs
        value={activeTab}
        onValueChange={(value) => {
          if (isContextTab(value)) onTabChange(value);
        }}
        className="min-h-0 flex-1"
      >
        <TabsList className="w-full">
          <TabsTrigger value="contents">Contents</TabsTrigger>
          <TabsTrigger value="world">World</TabsTrigger>
          <TabsTrigger value="notes">Notes</TabsTrigger>
        </TabsList>

        <TabsContent value="contents" className="flex min-h-0 flex-col gap-3 overflow-auto">
          <header className="flex items-center gap-2 text-sm font-medium text-foreground">
            <FileText className="size-4" aria-hidden="true" />
            Content
          </header>
          <div className="grid grid-cols-2 gap-2" aria-label="Content kind">
            {contentKindOptions.map(({ kind, label }) => (
              <Button
                key={kind}
                type="button"
                variant={activeContentKind === kind ? "secondary" : "outline"}
                size="xs"
                onClick={() => onContentKindChange(kind)}
              >
                {label}
              </Button>
            ))}
          </div>
          {contentLoading ? (
            <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
              Loading content...
            </p>
          ) : contentError ? (
            <p className="rounded-md border border-dashed border-border p-3 text-sm text-destructive" role="alert">
              Could not load content: {contentError}
            </p>
          ) : contentItems.length === 0 ? (
            <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
              No content found for this kind.
            </p>
          ) : (
            <nav className="flex flex-col gap-2" aria-label="Content items">
              {contentItems.map((item) => (
                <Button
                  key={item.id}
                  type="button"
                  variant={item.id === selectedContentId ? "secondary" : "ghost"}
                  className="h-auto w-full justify-start whitespace-normal px-3 py-2 text-left"
                  onClick={() => onSelectContent(item)}
                >
                  <span className="flex min-w-0 flex-col items-start gap-0.5">
                    <span className="line-clamp-2 text-sm font-medium">{item.title}</span>
                    <span className="text-xs text-muted-foreground">Revision {item.currentRevision}</span>
                  </span>
                </Button>
              ))}
            </nav>
          )}
        </TabsContent>

        <TabsContent value="world" className="flex flex-col gap-3">
          <header className="flex items-center gap-2 text-sm font-medium text-foreground">
            <BookOpen className="size-4" aria-hidden="true" />
            Story bible
          </header>
          <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
            World lookup will show linked story bible entries here.
          </p>
        </TabsContent>

        <TabsContent value="notes" className="flex flex-col gap-3">
          <header className="flex items-center gap-2 text-sm font-medium text-foreground">
            <StickyNote className="size-4" aria-hidden="true" />
            Notes
          </header>
          <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
            Attached project notes will appear here.
          </p>
        </TabsContent>
      </Tabs>
    </aside>
  );
}
