// web/src/hooks/useTraceStream.ts
// React hook that consumes FlareGoClient.streamFlamegraphs() and exposes the
// latest decoded flamegraph JSON frame.  It handles automatic reconnect with
// exponential back‑off and exposes connection status to the caller.

import { useEffect, useRef, useState } from "react";
import { FlareGoClient } from "../api/client";

export interface UseTraceStreamOptions {
  gatewayURL: string;
  authToken?: string;
  reconnect?: boolean; // default true
}

export type ConnectionState =
  | "connecting"
  | "connected"
  | "disconnected"
  | "error";

export function useTraceStream(opts: UseTraceStreamOptions) {
  const { gatewayURL, authToken, reconnect = true } = opts;
  const [frame, setFrame] = useState<any | null>(null);
  const [state, setState] = useState<ConnectionState>("connecting");
  const backoffRef = useRef<number>(1000); // ms
  const abortCtlRef = useRef<AbortController | null>(null);

  useEffect(() => {
    abortCtlRef.current?.abort();
    const abortCtl = new AbortController();
    abortCtlRef.current = abortCtl;

    const client = new FlareGoClient({ gatewayURL, authToken });

    async function start() {
      setState("connecting");
      try {
        for await (const chunk of client.streamFlamegraphs()) {
          if (abortCtl.signal.aborted) return;
          const json = JSON.parse(new TextDecoder().decode(chunk));
          setFrame(json);
          setState("connected");
          backoffRef.current = 1000; // reset on success
        }
        // Stream ended gracefully
        setState("disconnected");
      } catch (err) {
        console.error("trace stream error", err);
        setState("error");
        if (!reconnect) return;
        // exponential backoff capped at 30s
        const wait = backoffRef.current;
        backoffRef.current = Math.min(backoffRef.current * 2, 30000);
        await new Promise((res) => setTimeout(res, wait));
        if (!abortCtl.signal.aborted) start();
      }
    }
    start();
    return () => abortCtl.abort();
  }, [gatewayURL, authToken, reconnect]);

  return { frame, state };
}
