import { useRef, useState } from "react";
import { Button } from "../../shared/ui/button";
import { loadFile, loadHistory, loadVersion, publishTree, restoreVersion, uploadBlob, type ProjectVersion, type TreeEntry, type VersionPage } from "./fileApi";

const input = "min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-2 focus-visible:outline-ring";
const excluded = (path: string) => path.split("/").some(part => [".git", ".edda", "node_modules", "__pycache__", ".DS_Store", ".env"].includes(part) || part.startsWith(".env."));
type ImportFile = { file: File; path: string; id: string };
type Props = { version: ProjectVersion; disabled: boolean; action: (work: () => Promise<void>) => Promise<void>; onVersion: (version: ProjectVersion) => void };
export function FileActions({ version, disabled, action, onVersion }: Props) {
  const [selected, setSelected] = useState("");
  const [destination, setDestination] = useState("");
  const [deleting, setDeleting] = useState(false);
  const [history, setHistory] = useState<VersionPage | null>(null);
  const [past, setPast] = useState<ProjectVersion | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [files, setFiles] = useState<ImportFile[]>([]);
  const [omitted, setOmitted] = useState<string[]>([]);
  const fileInput = useRef<HTMLInputElement>(null);
  const folderInput = useRef<HTMLInputElement>(null);
  const pending = useRef<{ signature: string; operation: string } | null>(null);
  const importEntries = useRef<{ signature: string; entries: TreeEntry[] } | null>(null);
  function operation(signature: string) {
    if (pending.current?.signature !== signature) pending.current = { signature, operation: crypto.randomUUID() };
    return pending.current.operation;
  }
  async function publish(entries: TreeEntry[]) {
    const saved = await publishTree(version, entries, operation(JSON.stringify([version.id, entries])));
    onVersion(saved); setSelected(""); setDestination(""); setDeleting(false); setHistory(null); setPast(null); setPreview(null);
  }
  function chooseFiles(list: FileList | null) {
    if (!list) return;
    const chosen: ImportFile[] = [], skipped: string[] = [];
    for (const file of Array.from(list)) {
      // The chosen folder itself is the project root, not an extra nesting level.
      const path = file.webkitRelativePath ? file.webkitRelativePath.split("/").slice(1).join("/") : file.name;
      if (excluded(path)) skipped.push(path); else chosen.push({ file, path, id: crypto.randomUUID() });
    }
    setFiles(chosen); setOmitted(skipped); importEntries.current = null;
  }
  async function upload() {
    const signature = JSON.stringify([version.id, files.map(f => [f.path, f.id])]);
    if (importEntries.current?.signature !== signature) {
      const entries = [...version.entries]; const paths = new Map(entries.map(e => [e.path, e]));
      const added = new Set<string>();
      for (const item of files) {
        if (paths.has(item.path) || added.has(item.path)) throw new Error(`A file already exists at ${item.path}. Rename it or choose another folder.`);
        added.add(item.path);
        const parts = item.path.split("/"); parts.pop();
        for (let i = 1; i <= parts.length; i++) {
          const path = parts.slice(0, i).join("/");
          if (paths.get(path)?.kind === "file") throw new Error(`A file blocks the folder ${path}.`);
          if (!paths.has(path)) { const entry: TreeEntry = { id: crypto.randomUUID(), path, kind: "directory", bytes: 0 }; paths.set(path, entry); entries.push(entry); }
        }
        entries.push({ id: item.id, path: item.path, kind: "file", ...await uploadBlob(version.projectId, item.file) });
      }
      importEntries.current = { signature, entries };
    }
    await publish(importEntries.current.entries); setFiles([]); setOmitted([]); importEntries.current = null;
  }
  return <div className="mt-6 space-y-4 border-t border-border pt-4">
    <details>
      <summary className="cursor-pointer py-2 text-sm font-medium">Organize files</summary>
      <div className="mt-3 grid gap-3">
        <label className="grid gap-1 text-sm">File or folder<select className={input} value={selected} disabled={disabled} onChange={e => { setSelected(e.target.value); setDeleting(false); }}><option value="">Choose…</option>{version.entries.map(e => <option key={e.id} value={e.path}>{e.path}</option>)}</select></label>
        <form className="grid gap-2" onSubmit={event => { event.preventDefault(); if (!selected || !destination.trim()) return; void action(async () => {
          const target = destination.trim(); if (excluded(target)) throw new Error("This path is reserved for local settings or caches.");
          await publish(version.entries.map(entry => entry.path === selected || entry.path.startsWith(`${selected}/`) ? { ...entry, path: target + entry.path.slice(selected.length) } : entry));
        }); }}>
          <label className="grid gap-1 text-sm">New path<input className={input} disabled={disabled} value={destination} onChange={e => setDestination(e.target.value)} placeholder="chapters/chapter-01.md" /></label>
          <Button type="submit" variant="outline" disabled={disabled || !selected || !destination.trim()}>Move or rename</Button>
        </form>
        {deleting ? <div className="grid gap-2 text-sm"><p className="break-words">Remove {selected} and everything inside it? Previous versions remain available.</p><Button variant="destructive" disabled={disabled} onClick={() => void action(() => publish(version.entries.filter(e => e.path !== selected && !e.path.startsWith(`${selected}/`))))}>Confirm removal</Button><Button variant="outline" disabled={disabled} onClick={() => setDeleting(false)}>Cancel removal</Button></div> : <Button variant="outline" disabled={disabled || !selected} onClick={() => setDeleting(true)}>Remove…</Button>}
      </div>
    </details>
    <details>
      <summary className="cursor-pointer py-2 text-sm font-medium">Bring local files</summary>
      <div className="mt-3 grid gap-3 text-sm">
        <p className="text-muted-foreground">Choose files or a folder. Paths and file contents are kept. Existing files are never replaced.</p>
        <input className="hidden" type="file" multiple ref={fileInput} onChange={e => chooseFiles(e.target.files)} />
        <input className="hidden" type="file" multiple ref={node => { folderInput.current = node; node?.setAttribute("webkitdirectory", ""); }} onChange={e => chooseFiles(e.target.files)} />
        <Button variant="outline" disabled={disabled} onClick={() => fileInput.current?.click()}>Choose files</Button>
        <Button variant="outline" disabled={disabled} onClick={() => folderInput.current?.click()}>Choose folder</Button>
        {!!omitted.length && <details><summary>{omitted.length} local settings/cache files excluded</summary><ul className="max-h-40 overflow-auto">{omitted.map(p => <li className="break-all" key={p}>{p}</li>)}</ul></details>}
        {!!files.length && <><p>{files.length} files ready to add.</p><ul className="max-h-40 overflow-auto text-muted-foreground">{files.map(f => <li className="break-all" key={f.id}>{f.path}</li>)}</ul><Button disabled={disabled} onClick={() => void action(upload)}>Add selected files</Button><Button variant="outline" disabled={disabled} onClick={() => { setFiles([]); setOmitted([]); }}>Cancel import</Button></>}
        <p className="text-xs text-muted-foreground">For ongoing local work and empty folders, use edda attach and edda send.</p>
      </div>
    </details>
    <details onToggle={event => { if (event.currentTarget.open && !history && !disabled) void action(async () => setHistory(await loadHistory(version.projectId))); }}>
      <summary className="cursor-pointer py-2 text-sm font-medium">Version history</summary>
      <div className="mt-3 grid gap-3 text-sm">
        <p className="text-muted-foreground">Every save keeps the previous files. Save or download your draft before restoring.</p>
        <Button variant="outline" disabled={disabled} onClick={() => void action(async () => setHistory(await loadHistory(version.projectId)))}>Refresh history</Button>
        {history && <label className="grid gap-1">Saved version<select className={input} value={past?.id ?? ""} disabled={disabled} onChange={e => { const id = e.target.value; if (id) void action(async () => { setPast(await loadVersion(version.projectId, id)); setPreview(null); }); }}><option value="">Choose…</option>{history.versions.map((v, index) => <option value={v.id} key={v.id}>{index === 0 ? "Latest · " : ""}{new Date(v.createdAt).toLocaleString()} · {v.id.slice(-8)}</option>)}</select></label>}
        {history?.next && <Button variant="outline" disabled={disabled} onClick={() => void action(async () => { const next = await loadHistory(version.projectId, history.next); setHistory({ versions: [...history.versions, ...next.versions], next: next.next }); })}>Older versions</Button>}
        {past && <>
          <p>{past.entries.length} files and folders.</p>
          <ul className="max-h-48 overflow-auto">{past.entries.map(e => { const now = version.entries.find(n => n.id === e.id); return <li key={e.id} className="break-all">{e.path}{!now ? " · removed" : now.path !== e.path ? " · moved" : now.sha256 !== e.sha256 ? " · changed" : ""}</li>; })}</ul>
          <label className="grid gap-1">Preview saved file<select className={input} value="" disabled={disabled} onChange={e => { const entry = past.entries.find(f => f.id === e.target.value); if (entry) void action(async () => {
            if (entry.bytes > 1024 * 1024) throw new Error("Preview is limited to 1 MiB. Download this version with the CLI to inspect larger files.");
            const text = new TextDecoder("utf-8", { fatal: true }).decode(await (await loadFile(past, entry)).arrayBuffer()); if (text.includes("\0")) throw new Error("This file is binary; use a local copy to inspect it."); setPreview(text);
          }); }}><option value="">Choose…</option>{past.entries.filter(e => e.kind === "file").map(e => <option value={e.id} key={e.id}>{e.path}</option>)}</select></label>
          {preview !== null && <pre className="max-h-64 overflow-auto whitespace-pre-wrap break-words rounded border border-border p-2 text-xs">{preview}</pre>}
          <p className="text-muted-foreground">Restoring replaces the current tree with this saved version. The current tree stays in history.</p>
          <Button variant="outline" disabled={disabled || past.id === version.id} onClick={() => void action(async () => { const saved = await restoreVersion(version, past.id, operation(`${version.id}:restore:${past.id}`)); onVersion(saved); setHistory(null); setPast(null); setPreview(null); })}>Restore selected version</Button>
        </>}
      </div>
    </details>
  </div>;
}
