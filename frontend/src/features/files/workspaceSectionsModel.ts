export type WorkspaceSection = { id: string; label: string; folderId: string; purpose?: string; [key: string]: unknown };
export type WorkspaceConfig = { schemaVersion: 1; sections: WorkspaceSection[]; [key: string]: unknown };
export function decodeSections(body: string): WorkspaceConfig {
  if (new TextEncoder().encode(body).length > 65536) throw new Error('Настройки разделов превышают 64 КиБ.');
  const value: unknown = JSON.parse(body);
  if (!value || typeof value !== 'object' || !('schemaVersion' in value) || value.schemaVersion !== 1 || !('sections' in value) || !Array.isArray(value.sections) || value.sections.length > 50) throw new Error('Неизвестный формат настроек разделов.');
  const ids = new Set<string>();
  for (const row of value.sections) {
    if (!row || typeof row !== 'object' || typeof row.id !== 'string' || !row.id || ids.has(row.id) || typeof row.label !== 'string' || !row.label.trim() || row.label.length > 100 || typeof row.folderId !== 'string' || row.folderId === '-' || (row.purpose !== undefined && (typeof row.purpose !== 'string' || row.purpose.length > 100))) throw new Error('Проверьте названия и назначения разделов.');
    ids.add(row.id);
  }
  return value as WorkspaceConfig;
}
