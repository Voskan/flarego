import React, { useRef, useState } from "react";
import { FlameGraphCanvas } from "./FlameGraphCanvas";

function computeDiff(head: any, base: any): any {
  // Простой рекурсивный diff для flamegraph (JS, не учитывает все edge-cases)
  if (!head && !base) return null;
  if (!head) head = { name: base.name, value: 0, children: {} };
  if (!base) base = { name: head.name, value: 0, children: {} };
  const node: any = {
    name: head.name,
    value: (head.value || 0) - (base.value || 0),
    children: {},
  };
  const keys = new Set([
    ...Object.keys(head.children || {}),
    ...Object.keys(base.children || {}),
  ]);
  for (const k of keys) {
    const child = computeDiff(
      head.children ? head.children[k] : undefined,
      base.children ? base.children[k] : undefined
    );
    if (
      child &&
      (child.value !== 0 || Object.keys(child.children || {}).length > 0)
    ) {
      node.children[k] = child;
    }
  }
  if (node.value === 0 && Object.keys(node.children).length === 0) return null;
  return node;
}

export const DiffDrop: React.FC = () => {
  const [left, setLeft] = useState<any | null>(null);
  const [right, setRight] = useState<any | null>(null);
  const [diff, setDiff] = useState<any | null>(null);

  const handleFile = async (file: File, setter: (d: any) => void) => {
    const buf = await file.arrayBuffer();
    let data: any;
    try {
      // eslint-disable-next-line @typescript-eslint/no-var-requires
      const pako: any = require("pako");
      const gunzip = pako.ungzip;
      const json = new TextDecoder().decode(gunzip(new Uint8Array(buf)));
      data = JSON.parse(json);
    } catch {
      data = JSON.parse(new TextDecoder().decode(buf));
    }
    setter(data);
  };

  React.useEffect(() => {
    if (left && right) {
      setDiff(computeDiff(right, left));
    }
  }, [left, right]);

  return (
    <div className="flex flex-col space-y-4">
      <div className="flex space-x-4">
        <div className="flex-1">
          <input
            type="file"
            accept=".fgo,.json,.gz"
            onChange={(e) =>
              e.target.files && handleFile(e.target.files[0], setLeft)
            }
          />
          {left && <FlameGraphCanvas data={left} height={300} />}
        </div>
        <div className="flex-1">
          <input
            type="file"
            accept=".fgo,.json,.gz"
            onChange={(e) =>
              e.target.files && handleFile(e.target.files[0], setRight)
            }
          />
          {right && <FlameGraphCanvas data={right} height={300} />}
        </div>
      </div>
      {diff && (
        <div>
          <h3 className="text-sm font-bold mb-2">Diff (right - left)</h3>
          <FlameGraphCanvas data={diff} height={400} />
        </div>
      )}
    </div>
  );
};
