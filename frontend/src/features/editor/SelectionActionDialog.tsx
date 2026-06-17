import { useDispatch, useSelector } from "react-redux";

import { Button } from "../../shared/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../../shared/ui/dialog";
import { Textarea } from "../../shared/ui/textarea";
import { editorActions, type EditorState } from "./editorSlice";

type SelectionActionRootState = {
  editor: EditorState;
};

const actionLabels = {
  rewrite: {
    title: "Rewrite selection",
    description: "Add guidance for how this passage should change.",
    placeholder: "Make it sharper, warmer, shorter...",
    submit: "Preview rewrite",
  },
  check: {
    title: "Check selection",
    description: "Describe what the review should focus on.",
    placeholder: "Check continuity, tone, grammar...",
    submit: "Run check",
  },
  note: {
    title: "Note on selection",
    description: "Capture note context for the selected passage.",
    placeholder: "Remember this detail for later...",
    submit: "Notes are next",
  },
};

export function SelectionActionDialog() {
  const dispatch = useDispatch();
  const { actionModal, selection } = useSelector((state: SelectionActionRootState) => state.editor);
  const action = actionModal ? actionLabels[actionModal.kind] : null;
  const submitDisabled = !selection || actionModal?.kind === "note";

  return (
    <Dialog
      open={actionModal !== null}
      onOpenChange={(open) => {
        if (!open) dispatch(editorActions.closeActionModal());
      }}
    >
      {actionModal && action ? (
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{action.title}</DialogTitle>
            <DialogDescription>{action.description}</DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-4">
            <div className="rounded-md border border-border bg-muted/40 p-3">
              <p className="mb-1 text-xs font-medium text-muted-foreground">Selected text</p>
              <p className="editor-action-preview line-clamp-4 overflow-hidden text-ellipsis text-sm leading-6 text-foreground">
                {selection?.previewText || "No text selected."}
              </p>
            </div>

            <Textarea
              id="selection-action-instructions"
              name="selection-action-instructions"
              value={actionModal.instructions}
              onChange={(event) => dispatch(editorActions.setActionInstructions(event.target.value))}
              aria-label={`${action.title} instructions`}
              placeholder={action.placeholder}
              className="min-h-28 resize-y bg-background"
            />
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => dispatch(editorActions.closeActionModal())}>
              Cancel
            </Button>
            <Button type="button" disabled={submitDisabled}>
              {selection ? action.submit : "Select text first"}
            </Button>
          </DialogFooter>
        </DialogContent>
      ) : null}
    </Dialog>
  );
}
