import React, { useRef } from "react";

export const ReplayDrop: React.FC<{ onLoad: (data: any) => void }> = ({
  onLoad,
}) => {
  const inputRef = useRef<HTMLInputElement>(null);

  const handleFile = async (file: File) => {
    const buf = await file.arrayBuffer();
    let data: any;
    try {
      // Try gzip
      // eslint-disable-next-line @typescript-eslint/no-var-requires
      const pako: any = require("pako");
      const gunzip = pako.ungzip;
      const json = new TextDecoder().decode(gunzip(new Uint8Array(buf)));
      data = JSON.parse(json);
    } catch {
      // Fallback: plain JSON
      data = JSON.parse(new TextDecoder().decode(buf));
    }
    onLoad(data);
  };

  return (
    <div
      className="border-2 border-dashed rounded-xl p-8 text-center cursor-pointer"
      onClick={() => inputRef.current?.click()}
      onDrop={(e) => {
        e.preventDefault();
        if (e.dataTransfer.files.length) handleFile(e.dataTransfer.files[0]);
      }}
      onDragOver={(e) => e.preventDefault()}
    >
      <input
        type="file"
        accept=".fgo,.json,.gz"
        ref={inputRef}
        style={{ display: "none" }}
        onChange={(e) => {
          if (e.target.files?.[0]) handleFile(e.target.files[0]);
        }}
      />
      <span>Drag & drop .fgo file or click to select</span>
    </div>
  );
};
