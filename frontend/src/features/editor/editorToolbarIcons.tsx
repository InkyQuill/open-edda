import { Bold, Italic, Strikethrough, Code, List, ListOrdered, ListTodo, Link, Image, FileCode, Table, Minus, Undo2, Redo2, Eye } from "lucide-react";
import type { GalleyToolbarOptions } from "@inkyquill/galley-editor";

// Use Galley's public icon slots; the editor owns labels, commands and keyboard behavior.
export const editorToolbarIcons: GalleyToolbarOptions["icons"] = {
  bold: <Bold aria-hidden="true" />,
  italic: <Italic aria-hidden="true" />,
  strikethrough: <Strikethrough aria-hidden="true" />,
  inlineCode: <Code aria-hidden="true" />,
  bulletList: <List aria-hidden="true" />,
  orderedList: <ListOrdered aria-hidden="true" />,
  taskList: <ListTodo aria-hidden="true" />,
  link: <Link aria-hidden="true" />,
  image: <Image aria-hidden="true" />,
  codeBlock: <FileCode aria-hidden="true" />,
  table: <Table aria-hidden="true" />,
  divider: <Minus aria-hidden="true" />,
  undo: <Undo2 aria-hidden="true" />,
  redo: <Redo2 aria-hidden="true" />,
  mode: <Eye aria-hidden="true" />,
};
