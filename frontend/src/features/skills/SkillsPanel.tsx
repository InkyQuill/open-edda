import { CheckCircle2, Loader2 } from "lucide-react";
import { useMemo } from "react";
import { useDispatch, useSelector } from "react-redux";

import type { AppDispatch, RootState } from "../../app/store/store";
import { Button } from "../../shared/ui/button";
import { skillsActions } from "./skillsSlice";
import { saveSessionSkills } from "./skillsThunks";

type SkillsPanelProps = {
  projectId: string;
  sessionId: string | null;
};

export function SkillsPanel({ projectId, sessionId }: SkillsPanelProps) {
  const dispatch = useDispatch<AppDispatch>();
  const { saveStatus, selectedSkillIds, sessionSkillsStatus, skills, skillsStatus } = useSelector(
    (state: RootState) => state.skills,
  );
  const selectedSkillIdSet = useMemo(() => new Set(selectedSkillIds), [selectedSkillIds]);
  const isLoading = skillsStatus === "pending" || sessionSkillsStatus === "pending";
  const isSaving = saveStatus === "pending";
  const hasSession = Boolean(sessionId);

  function handleToggle(skillId: string): void {
    if (!sessionId) return;
    const nextSkillIds = selectedSkillIdSet.has(skillId)
      ? selectedSkillIds.filter((selectedSkillId) => selectedSkillId !== skillId)
      : [...selectedSkillIds, skillId];

    dispatch(skillsActions.toggleSelectedSkillId(skillId));
    void dispatch(saveSessionSkills({ projectId, sessionId, skillIds: nextSkillIds }));
  }

  return (
    <section className="flex flex-col gap-3" aria-labelledby="skills-panel-title">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h3 id="skills-panel-title" className="text-sm font-medium text-foreground">
            Session skills
          </h3>
          <p className="text-xs text-muted-foreground">
            {hasSession ? "Choose skills for this assistant chat." : "Start a chat before selecting skills."}
          </p>
        </div>
        {isSaving ? <Loader2 className="size-4 animate-spin text-muted-foreground" aria-hidden="true" /> : null}
      </div>

      {isLoading ? (
        <p className="flex items-center gap-2 rounded-md border border-border p-3 text-sm text-muted-foreground">
          <Loader2 className="size-4 animate-spin" aria-hidden="true" />
          Loading available skills...
        </p>
      ) : null}

      {!isLoading && skills.length === 0 ? (
        <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
          No skills installed for this project.
        </p>
      ) : null}

      {skills.length > 0 ? (
        <div className="flex max-h-64 flex-col gap-2 overflow-auto pr-1">
          {skills.map((skill) => {
            const isSelected = selectedSkillIdSet.has(skill.id);

            return (
              <Button
                key={skill.id}
                type="button"
                variant={isSelected ? "secondary" : "outline"}
                className="h-auto justify-start px-3 py-2 text-left"
                aria-pressed={isSelected}
                disabled={!hasSession || isLoading || isSaving}
                onClick={() => handleToggle(skill.id)}
              >
                {isSelected ? (
                  <CheckCircle2 data-icon="inline-start" aria-hidden="true" />
                ) : (
                  <span data-icon="inline-start" className="size-4" aria-hidden="true" />
                )}
                <span className="min-w-0 flex-1">
                  <span className="block truncate font-medium">{skill.displayName}</span>
                  <span className="block truncate text-xs text-muted-foreground">{skill.description}</span>
                </span>
              </Button>
            );
          })}
        </div>
      ) : null}
    </section>
  );
}
