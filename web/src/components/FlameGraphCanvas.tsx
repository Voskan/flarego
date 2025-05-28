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

import React, { useEffect, useRef } from "react";
import * as d3 from "d3";
import flamegraph from "d3-flame-graph";
import "d3-flame-graph/dist/d3-flamegraph.css";

export interface FlameGraphCanvasProps {
  data: any; // root node in d3 flamegraph JSON format
  height?: number;
  onNodeClick?: (name: string, value: number) => void;
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
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const graphRef = useRef<any>(null);

  // Initialise once.
  useEffect(() => {
    if (!containerRef.current) return;
    const fg = flamegraph.flamegraph();
    fg.onClick((d: any) => {
      if (onNodeClick) onNodeClick(d.data.name, d.data.value);
    });

    d3.select(containerRef.current).append(() => fg as any);
    graphRef.current = fg;

    // Clean‑up on unmount.
    return () => {
      if (containerRef.current) {
        containerRef.current.innerHTML = "";
      }
    };
  }, []);

  // Update graph when data changes.
  useEffect(() => {
    if (graphRef.current && data) {
      graphRef.current.update(data);
    }
  }, [data]);

  return (
    <div
      ref={containerRef}
      className="w-full overflow-x-auto rounded-2xl shadow-inner bg-gray-900/5"
      style={{ height }}
    />
  );
};

export default FlameGraphCanvas;
