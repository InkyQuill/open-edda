import { AlertCircle, Loader2, Sparkles } from "lucide-react";
import { useMemo } from "react";
import { useSelector } from "react-redux";

import type { RootState } from "../../app/store/store";
import { buttonVariants } from "../../shared/ui/button";

export function SkillChipsPanel() {
  const { selectedSkillIds, sessionSkillsStatus, skills, skillsStatus } = useSelector((state: RootState) => state.skills);
  const selectedSkillIdSet = useMemo(() => new Set(selectedSkillIds), [selectedSkillIds]);
  const selectedSkills = skills.filter((skill) => selectedSkillIdSet.has(skill.id));
  const hasDisabledScripts = selectedSkills.some((skill) => skill.scriptsDisabled);
  const isLoading = skillsStatus === "pending" || sessionSkillsStatus === "pending";

  return (
    <section className="flex flex-col gap-3" aria-labelledby="skill-chips-title">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h3 id="skill-chips-title" className="text-sm font-medium text-foreground">
            Skills
          </h3>
          <p className="text-xs text-muted-foreground">
            {skills.length === 1 ? "1 installed skill detected." : `${skills.length} installed skills detected.`}
          </p>
        </div>
        <Sparkles className="size-4 text-muted-foreground" aria-hidden="true" />
      </div>

      {isLoading ? (
        <p className="flex items-center gap-2 rounded-md border border-border p-3 text-sm text-muted-foreground">
          <Loader2 className="size-4 animate-spin" aria-hidden="true" />
          Loading skills...
        </p>
      ) : null}

      {!isLoading && skills.length === 0 ? (
        <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
          No skills installed for this project.
        </p>
      ) : null}

      {!isLoading && skills.length > 0 && selectedSkills.length === 0 ? (
        <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
          No skills selected for this chat.
        </p>
      ) : null}

      {selectedSkills.length > 0 ? (
        <div className="flex flex-wrap gap-2">
          {selectedSkills.map((skill) => (
            <span key={skill.id} className={buttonVariants({ variant: "secondary", size: "xs" })}>
              {skill.displayName}
            </span>
          ))}
        </div>
      ) : null}

      {hasDisabledScripts ? (
        <p className="flex items-start gap-2 rounded-md border border-dashed border-border p-3 text-xs text-muted-foreground">
          <AlertCircle className="mt-0.5 size-4" aria-hidden="true" />
          Scripts are disabled for one or more selected skills.
        </p>
      ) : null}
    </section>
  );
}
