// web/src/components/FlameGraphCanvas.tsx
// React component that renders a flame graph using d3-flame-graph.  The
// component accepts a raw JSON object (the root Frame) and re-renders whenever
// a new snapshot arrives.  It is kept deliberately stateless so the parent can
// control zoom, palette and timeline.
//
// Dependencies (package.json):
//   "d3": "^7.9.0",
//   "d3-flame-graph": "^4.1.5",
//   "@types/d3": "^7.4.0"
//
// TailwindCSS is used for layout; no inline styles besides sizing attrs.

import React, { useEffect, useRef, useState } from "react";
import * as d3 from "d3";
import { flamegraph } from "d3-flame-graph";
import "d3-flame-graph/dist/d3-flamegraph.css";

export interface FlameGraphCanvasProps {
  data: any; // root node in d3 flamegraph JSON format
  height?: number;
  onNodeClick?: (name: string, value: number) => void;
  filter?: string;
  onFilterChange?: (val: string) => void;
}

function filterTree(node: any, filter: string): any {
  if (!filter) return node;
  const match =
    node.name && node.name.toLowerCase().includes(filter.toLowerCase());
  const children = node.children
    ? Object.fromEntries(
        Object.entries(node.children)
          .map(([k, v]) => [k, filterTree(v, filter)])
          .filter(([, v]) => v)
      )
    : {};
  if (match || Object.keys(children).length > 0) {
    return { ...node, children, _matched: match };
  }
  return null;
}

/**
 * FlameGraphCanvas renders an interactive flame graph inside a responsive
 * container.  It auto-resizes on window change and exposes the clicked node to
 * the parent via onNodeClick.
 */
export const FlameGraphCanvas: React.FC<FlameGraphCanvasProps> = ({
  data,
  height = 400,
  onNodeClick,
  filter = "",
  onFilterChange,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const graphRef = useRef<any>(null);
  const [search, setSearch] = useState(filter);

  // Initialise once.
  useEffect(() => {
    if (!containerRef.current) return;
    const fg = flamegraph();
    fg.onClick((d: any) => {
      if (onNodeClick) onNodeClick(d.data.name, d.data.value);
    });
    d3.select(containerRef.current).append(() => fg as any);
    graphRef.current = fg;
    return () => {
      if (containerRef.current) {
        containerRef.current.innerHTML = "";
      }
    };
  }, []);

  // Update graph when data or filter changes.
  useEffect(() => {
    if (graphRef.current && data) {
      const filtered = filterTree(data, search);
      if (filtered) {
        graphRef.current.update(filtered);
        // Подсветка найденных узлов
        d3.select(containerRef.current)
          .selectAll(".d3-flame-graph-label")
          .style("font-weight", (d: any) => (d.data._matched ? "bold" : ""))
          .style("fill", (d: any) => (d.data._matched ? "#ff6b6b" : ""));
      } else {
        graphRef.current.update({ name: "No match", value: 1, children: {} });
      }
    }
  }, [data, search]);

  return (
    <div className="w-full">
      {onFilterChange && (
        <input
          className="mb-2 p-1 border rounded w-full text-xs"
          placeholder="Search function/package..."
          value={search}
          onChange={(e) => {
            setSearch(e.target.value);
            onFilterChange(e.target.value);
          }}
        />
      )}
      <div
        ref={containerRef}
        className="w-full overflow-x-auto rounded-2xl shadow-inner bg-gray-900/5"
        style={{ height }}
      />
    </div>
  );
};

export default FlameGraphCanvas;
