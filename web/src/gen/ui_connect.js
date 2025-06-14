// internal/proto/ui.proto
// This schema defines the gRPC contract between the FlareGo gateway and the UI.
// The protocol is designed for real-time streaming of flamegraph data to the UI.

// @generated by protoc-gen-connect-es v0.13.0 with parameter "import_extension=ts"
// @generated from file ui.proto (package agentpb, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { Empty, MethodKind } from "@bufbuild/protobuf";
import { FlamegraphChunk } from "./common_pb";

/**
 * UIService is implemented by the gateway; the UI connects to stream
 * flamegraph data in real-time.
 *
 * @generated from service agentpb.UIService
 */
export const UIService = {
  typeName: "agentpb.UIService",
  methods: {
    /**
     * StreamFlamegraphs streams flamegraph data to the UI.
     *
     * @generated from rpc agentpb.UIService.StreamFlamegraphs
     */
    streamFlamegraphs: {
      name: "StreamFlamegraphs",
      I: Empty,
      O: FlamegraphChunk,
      kind: MethodKind.ServerStreaming,
    },
  },
};
