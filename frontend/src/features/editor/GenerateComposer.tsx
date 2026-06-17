import { Sparkles } from "lucide-react";
import { useDispatch, useSelector } from "react-redux";

import { Button } from "../../shared/ui/button";
import { Textarea } from "../../shared/ui/textarea";
import { editorActions, type EditorState } from "./editorSlice";

type GenerateComposerRootState = {
  editor: EditorState;
};

type GenerateComposerProps = {
  disabled: boolean;
  helperText: string;
  onGenerate: () => void;
};

export function GenerateComposer({ disabled, helperText, onGenerate }: GenerateComposerProps) {
  const dispatch = useDispatch();
  const instructions = useSelector(
    (state: GenerateComposerRootState) => state.editor.generateInstructions,
  );

  return (
    <section className="flex flex-col gap-3 border-t border-border pt-5" aria-label="Generate">
      <Textarea
        id="generate-instructions"
        name="generate-instructions"
        value={instructions}
        onChange={(event) => dispatch(editorActions.setGenerateInstructions(event.target.value))}
        aria-label="Generate instructions"
        placeholder="Describe what to generate next..."
        disabled={disabled}
        className="min-h-24 resize-y bg-background"
      />
      <div className="flex items-center justify-between gap-3">
        <p className="text-xs text-muted-foreground">{helperText}</p>
        <Button type="button" disabled={disabled} onClick={onGenerate}>
          <Sparkles data-icon="inline-start" aria-hidden="true" />
          Generate
        </Button>
      </div>
    </section>
  );
}
