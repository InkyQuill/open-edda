import { describe, expect, it } from "vitest";
import { BUILT_IN_THEMES } from "@inkyquill/galley-themes";
import { resolveTheme, themeVariables } from "./theme";

describe("Edda themes", () => {
  it("defaults invalid or absent preferences to Thoth Light", () => {
    for (const id of [undefined, null, "unknown", 3]) {
      expect(resolveTheme(id).id).toBe("thoth-light");
    }
  });
  it("maps all catalog palettes to the same app and editor contract", () => {
    for (const theme of BUILT_IN_THEMES) {
      const variables = themeVariables(theme.id);
      expect(variables["--background"]).toBe(theme.tokens.app.bg);
      expect(variables["--foreground"]).toBe(theme.tokens.app.text);
      expect(variables["--ge-color-bg"]).toBe(theme.tokens.editor.bg);
      expect(variables["--primary"]).toBe(theme.tokens.app.focus);
      expect(variables.colorScheme).toBe(theme.scheme);
    }
  });
});
