import { Sparkles } from "lucide-react";

import { Button } from "../../shared/ui/button";

const placeholderSkills = ["Continuity", "Style", "Revision"];

export function SkillChipsPanel() {
  return (
    <section className="flex flex-col gap-3" aria-labelledby="skill-chips-title">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h3 id="skill-chips-title" className="text-sm font-medium text-foreground">
            Skills
          </h3>
          <p className="text-xs text-muted-foreground">3 installed skills detected.</p>
        </div>
        <Sparkles className="size-4 text-muted-foreground" aria-hidden="true" />
      </div>

      <div className="flex flex-wrap gap-2">
        {placeholderSkills.map((skill) => (
          <Button key={skill} type="button" variant="secondary" size="xs">
            {skill}
          </Button>
        ))}
      </div>
    </section>
  );
}
