// web/src/pages/Dashboard.tsx
// High‑level page that wires together the trace stream hook, the flame‑graph
// canvas and basic UI chrome (status indicator, connect form, simple node
// details side‑panel).  It is meant to be the landing route for the standalone
// web dashboard and the IDE WebView.

import React, { useState } from "react";
import { useTraceStream, ConnectionState } from "../hooks/useTraceStream";
import { FlameGraphCanvas } from "../components/FlameGraphCanvas";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

interface DashboardProps {
  defaultGatewayURL?: string;
  defaultAuthToken?: string;
}

export const Dashboard: React.FC<DashboardProps> = ({
  defaultGatewayURL = "http://localhost:4317",
  defaultAuthToken = "",
}) => {
  const [gatewayURL, setGatewayURL] = useState(defaultGatewayURL);
  const [token, setToken] = useState(defaultAuthToken);
  const { frame, state } = useTraceStream({ gatewayURL, authToken: token });
  const [selectedNode, setSelectedNode] = useState<{
    name: string;
    value: number;
  } | null>(null);

  const statusColor = (s: ConnectionState) => {
    switch (s) {
      case "connected":
        return "bg-green-500";
      case "connecting":
        return "bg-yellow-400";
      case "error":
        return "bg-red-500";
      default:
        return "bg-gray-400";
    }
  };

  return (
    <div className="p-4 grid grid-cols-12 gap-4 h-screen">
      {/* Sidebar */}
      <aside className="col-span-3 flex flex-col space-y-4 overflow-y-auto">
        <Card>
          <CardContent className="p-4 flex items-center space-x-2">
            <span
              className={`w-3 h-3 rounded-full ${statusColor(state)}`}
            ></span>
            <span className="text-sm font-medium capitalize">{state}</span>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4 space-y-2">
            <label className="block text-xs font-semibold">Gateway URL</label>
            <input
              className="w-full p-2 bg-gray-100 rounded-md text-sm"
              value={gatewayURL}
              onChange={(e) => setGatewayURL(e.target.value)}
            />
            <label className="block text-xs font-semibold">Auth Token</label>
            <input
              className="w-full p-2 bg-gray-100 rounded-md text-sm"
              value={token}
              onChange={(e) => setToken(e.target.value)}
            />
            <Button
              className="mt-2 w-full"
              onClick={() => window.location.reload()}
            >
              Reconnect
            </Button>
          </CardContent>
        </Card>

        {selectedNode && (
          <Card>
            <CardContent className="p-4 space-y-1">
              <h3 className="text-sm font-bold">Selected Node</h3>
              <p className="text-xs text-gray-700 break-all">
                {selectedNode.name}
              </p>
              <p className="text-xs text-gray-500">
                {(selectedNode.value / 1e6).toFixed(2)} ms
              </p>
            </CardContent>
          </Card>
        )}
      </aside>

      {/* Main flamegraph area */}
      <main className="col-span-9 flex flex-col">
        {frame ? (
          <FlameGraphCanvas
            data={frame}
            height={600}
            onNodeClick={(name, value) => setSelectedNode({ name, value })}
          />
        ) : (
          <div className="flex flex-1 items-center justify-center text-gray-400 text-sm">
            Waiting for data…
          </div>
        )}
      </main>
    </div>
  );
};

export default Dashboard;
