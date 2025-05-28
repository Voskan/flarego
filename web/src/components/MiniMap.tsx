// web/src/components/MiniMap.tsx
// MiniMap renders a compact sparkline panel visualising recent runtime
// metrics (heap bytes, blocked goroutines, GC pauses) so users can correlate
// numeric trends with the flamegraph view.  It expects the parent to feed a
// rolling history array; rendering is cheap thanks to Recharts.
//
// Dependencies (package.json):
//   "recharts": "^2.6.2"

import React from "react";
import { LineChart, Line, ResponsiveContainer, Tooltip, YAxis } from "recharts";

export interface MiniMapPoint {
  ts: number; // unix ms
  heap: number; // bytes
  blocked: number; // count
  gc: number; // pause ns
}

interface MiniMapProps {
  data: MiniMapPoint[]; // chronological (oldest → newest)
}

/**
 * MiniMap renders three overlaid sparklines in a tiny foot‑print panel.  Lines
 * share the same x axis but independent Y domains so relative shape matters
 * more than absolute numbers.  Colours match the pseudo-stack palette.
 */
export const MiniMap: React.FC<MiniMapProps> = ({ data }) => {
  if (data.length === 0) return null;

  return (
    <div className="w-full h-24 bg-gray-50 rounded-xl shadow-inner p-2">
      <ResponsiveContainer>
        <LineChart
          data={data}
          margin={{ top: 4, right: 8, left: 8, bottom: 4 }}
        >
          <YAxis hide domain={["auto", "auto"]} />
          <Tooltip
            formatter={(val: any) =>
              typeof val === "number" ? val.toLocaleString() : val
            }
            labelFormatter={(label: any) =>
              new Date(label).toLocaleTimeString()
            }
          />
          <Line
            type="monotone"
            dataKey="heap"
            stroke="#80cbc4"
            dot={false}
            strokeWidth={2}
          />
          <Line
            type="monotone"
            dataKey="blocked"
            stroke="#ef9a9a"
            dot={false}
            strokeWidth={2}
          />
          <Line
            type="monotone"
            dataKey="gc"
            stroke="#b39ddb"
            dot={false}
            strokeWidth={2}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
};

export default MiniMap;
