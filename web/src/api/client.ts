// web/src/api/client.ts
// Thin wrapper around gRPC‑Web transport (Connect-Web) to stream Flamegraph
// chunks from the FlareGo gateway.  The client exposes a single async
// generator `streamFlamegraphs()` which yields Uint8Array payloads that the UI
// component consumes and parses to JSON.
//
// Dependencies (already in package.json):
//   "@bufbuild/connect-web": "^1.5.0"
//   "protobufjs": "^7.2.0" (indirect via generated TS code)
//
// Code assumes ts-proto‑generated types for `internal/proto/gateway.proto` have
// been emitted under `src/gen/agentpb/*.ts` via buf or protoc‑gen‑ts.

import { createConnectTransport } from "@bufbuild/connect-web";
import { createPromiseClient } from "@bufbuild/connect";
import { GatewayService } from "../gen/agentpb/gateway_connect";
import { FlamegraphChunk } from "../gen/agentpb/gateway_pb";

interface ClientOptions {
  gatewayURL: string; // e.g., "https://localhost:4317"
  authToken?: string;
}

export class FlareGoClient {
  private client: ReturnType<typeof createPromiseClient<typeof GatewayService>>;

  constructor(private opts: ClientOptions) {
    const transport = createConnectTransport({
      baseUrl: opts.gatewayURL,
      useBinaryFormat: true,
      interceptors: [
        (next) => async (req) => {
          if (opts.authToken) {
            req.header.set("Authorization", `Bearer ${opts.authToken}`);
          }
          return next(req);
        },
      ],
    });
    this.client = createPromiseClient(GatewayService, transport);
  }

  /**
   * streamFlamegraphs yields Uint8Array payloads as they arrive from the
   * gateway.  The caller converts each to string/JSON as needed.
   */
  async *streamFlamegraphs(): AsyncGenerator<Uint8Array> {
    const stream = this.client.stream({});
    for await (const chunk of stream) {
      if (chunk instanceof FlamegraphChunk) {
        yield chunk.payload;
      }
    }
  }
}
