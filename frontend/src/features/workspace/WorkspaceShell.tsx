import {
  Bot,
  ClipboardCheck,
  FolderOpen,
  Library,
  MessageSquare,
  PenLine,
  Settings2,
} from "lucide-react";
import { useDispatch, useSelector } from "react-redux";

import { AssistantDrawer } from "../assistant/AssistantDrawer";
import { EditorFrame } from "../editor/EditorFrame";
import { ModelSettingsPanel } from "../model-settings/ModelSettingsPanel";
import { ModelStatus } from "../model-settings/ModelStatus";
import { ContextDrawer } from "../notes/ContextDrawer";
import { ReviewDrawer } from "../review/ReviewDrawer";
import { Button } from "../../shared/ui/button";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "../../shared/ui/sheet";
import type { ContentItem, ContentKind } from "../../types";
import type { DrawerTab, MobileSheet, WorkspaceMode, WorkspaceState } from "./workspaceSlice";
import { workspaceActions } from "./workspaceSlice";

type WorkspaceRootState = {
  workspace: WorkspaceState;
};

type ContextTab = "contents" | "world" | "notes";

type WorkspaceShellProps = {
  projectId: string;
  projectTitle: string;
  contentItems: ContentItem[];
  contentLoading: boolean;
  contentError: string | null;
  activeContentKind: ContentKind;
  selectedContent: ContentItem | null;
  onSelectContent: (item: ContentItem) => void;
  onContentKindChange: (kind: ContentKind) => void;
};

const modeButtons: Array<{
  mode: WorkspaceMode;
  label: string;
  icon: typeof PenLine;
}> = [
  { mode: "draft", label: "Draft", icon: PenLine },
  { mode: "assistant", label: "Assistant", icon: Bot },
  { mode: "review", label: "Review", icon: ClipboardCheck },
];

const mobileButtons: Array<{
  sheet: NonNullable<MobileSheet>;
  label: string;
  icon: typeof FolderOpen;
}> = [
  { sheet: "contents", label: "Files", icon: FolderOpen },
  { sheet: "assistant", label: "Assistant", icon: MessageSquare },
  { sheet: "review", label: "Review", icon: ClipboardCheck },
  { sheet: "world-notes", label: "World/Notes", icon: Library },
  { sheet: "model", label: "Model", icon: Settings2 },
];

function toContextTab(tab: DrawerTab): ContextTab {
  return tab === "world" || tab === "notes" ? tab : "contents";
}

function mobileSheetTitle(sheet: NonNullable<MobileSheet>): string {
  switch (sheet) {
    case "contents":
      return "Files";
    case "assistant":
      return "Assistant";
    case "review":
      return "Review";
    case "world-notes":
      return "World and notes";
    case "model":
      return "Model";
  }
}

export function WorkspaceShell({
  projectId,
  projectTitle,
  contentItems,
  contentLoading,
  contentError,
  activeContentKind,
  selectedContent,
  onSelectContent,
  onContentKindChange,
}: WorkspaceShellProps) {
  const dispatch = useDispatch();
  const workspace = useSelector((state: WorkspaceRootState) => state.workspace);
  const activeLeftTab = toContextTab(workspace.activeLeftTab);
  const rightDrawer =
    workspace.activeRightTab === "model" ? (
      <ModelSettingsPanel />
    ) : workspace.mode === "review" || workspace.activeRightTab === "tools" || workspace.activeRightTab === "revisions" ? (
      <ReviewDrawer projectId={projectId} />
    ) : (
      <AssistantDrawer projectId={projectId} />
    );

  const contextDrawer = (
    <ContextDrawer
      activeTab={activeLeftTab}
      contentItems={contentItems}
      contentLoading={contentLoading}
      contentError={contentError}
      activeContentKind={activeContentKind}
      selectedContentId={selectedContent?.id ?? null}
      onSelectContent={onSelectContent}
      onContentKindChange={onContentKindChange}
      onTabChange={(tab) => dispatch(workspaceActions.setActiveLeftTab(tab))}
    />
  );

  function renderMobileSheet(sheet: NonNullable<MobileSheet>) {
    if (sheet === "assistant") return <AssistantDrawer projectId={projectId} />;
    if (sheet === "review") return <ReviewDrawer projectId={projectId} />;
    if (sheet === "model") return <ModelSettingsPanel />;
    return (
      <ContextDrawer
        activeTab={sheet === "world-notes" ? "world" : "contents"}
        contentItems={contentItems}
        contentLoading={contentLoading}
        contentError={contentError}
        activeContentKind={activeContentKind}
        selectedContentId={selectedContent?.id ?? null}
        onSelectContent={onSelectContent}
        onContentKindChange={onContentKindChange}
        onTabChange={(tab) => dispatch(workspaceActions.setActiveLeftTab(tab))}
      />
    );
  }

  return (
    <main className="flex h-dvh min-h-0 flex-col bg-background text-foreground">
      <header className="flex items-center justify-between gap-3 border-b border-border px-4 py-3">
        <div className="min-w-0">
          <p className="text-xs text-muted-foreground">Workspace</p>
          <h1 className="truncate text-base font-semibold">{projectTitle}</h1>
        </div>
        <div className="hidden items-center gap-2 md:flex">
          {modeButtons.map(({ mode, label, icon: Icon }) => (
            <Button
              key={mode}
              type="button"
              variant={workspace.mode === mode ? "secondary" : "ghost"}
              onClick={() => dispatch(workspaceActions.setMode(mode))}
            >
              <Icon data-icon="inline-start" aria-hidden="true" />
              {label}
            </Button>
          ))}
        </div>
      </header>

      <div className="flex min-h-0 flex-1">
        <nav className="hidden w-16 shrink-0 flex-col items-center gap-2 border-r border-border p-2 md:flex" aria-label="Workspace mode">
          {modeButtons.map(({ mode, label, icon: Icon }) => (
            <Button
              key={mode}
              type="button"
              variant={workspace.mode === mode ? "secondary" : "ghost"}
              size="icon"
              aria-label={label}
              onClick={() => dispatch(workspaceActions.setMode(mode))}
            >
              <Icon data-icon="inline-start" aria-hidden="true" />
            </Button>
          ))}
        </nav>

        {workspace.leftDrawerOpen ? (
          <aside className="hidden min-h-0 shrink-0 border-r border-border p-4 md:block" style={{ width: workspace.leftDrawerWidth }}>
            {contextDrawer}
          </aside>
        ) : null}

        <section className="workspace-editor-stage flex min-w-0 flex-1 justify-center overflow-auto bg-muted/30 px-4 py-6">
          <EditorFrame
            content={selectedContent}
            mode={workspace.mode}
            contentLoading={contentLoading}
            contentError={contentError}
          />
        </section>

        {workspace.rightDrawerOpen ? (
          <aside className="hidden min-h-0 shrink-0 border-l border-border p-4 md:block" style={{ width: workspace.rightDrawerWidth }}>
            {rightDrawer}
          </aside>
        ) : null}
      </div>

      <nav className="grid grid-cols-5 border-t border-border bg-background p-2 md:hidden" aria-label="Workspace panels">
        {mobileButtons.map(({ sheet, label, icon: Icon }) => (
          <Button
            key={sheet}
            type="button"
            variant={workspace.mobileSheet === sheet ? "secondary" : "ghost"}
            className="h-auto flex-col gap-1 px-1 py-2 text-[0.7rem]"
            onClick={() => dispatch(workspaceActions.setMobileSheet(sheet))}
          >
            <Icon data-icon="inline-start" aria-hidden="true" />
            <span>{label}</span>
          </Button>
        ))}
      </nav>

      <Sheet
        open={workspace.mobileSheet !== null}
        onOpenChange={(open) => {
          if (!open) dispatch(workspaceActions.setMobileSheet(null));
        }}
      >
        {workspace.mobileSheet ? (
          <SheetContent side="bottom" className="max-h-[85dvh] overflow-auto">
            <SheetHeader>
              <SheetTitle>{mobileSheetTitle(workspace.mobileSheet)}</SheetTitle>
            </SheetHeader>
            <div className="min-h-0 px-4 pb-4">{renderMobileSheet(workspace.mobileSheet)}</div>
          </SheetContent>
        ) : null}
      </Sheet>
    </main>
  );
}
