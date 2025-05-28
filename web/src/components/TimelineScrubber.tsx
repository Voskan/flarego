// web/src/components/TimelineScrubber.tsx
// Interactive timeline scrubber that lets the user seek through a ring buffer
// of historical flamegraph snapshots.  The parent component provides the list
// of timestamps; the scrubber emits an index via onChange so the parent can
// render the corresponding frame.
//
// For simplicity the component uses an HTML <input type="range"> styled with
// Tailwind and shows the selected timestamp in relative human‐friendly words
// ("5 s ago").  When the list prop updates, the scrubber keeps the cursor at
// the latest item unless the user has actively scrubbed elsewhere (sticky).
import React, { useEffect, useState } from "react";
import { formatDistanceToNowStrict } from "date-fns";

export interface TimelineScrubberProps {
  times: number[]; // unix milliseconds sorted ascending (oldest → newest)
  onChange: (index: number) => void;
}

export const TimelineScrubber: React.FC<TimelineScrubberProps> = ({
  times,
  onChange,
}) => {
  const [idx, setIdx] = useState(times.length - 1);
  const [isUserScrubbing, setIsUserScrubbing] = useState(false);

  // Keep pointer at tail when new snapshot arrives and user not scrubbing.
  useEffect(() => {
    if (!isUserScrubbing) {
      setIdx(times.length - 1);
      onChange(times.length - 1);
    }
  }, [times]);

  // Call onChange when idx changes.
  useEffect(() => {
    if (idx >= 0 && idx < times.length) {
      onChange(idx);
    }
  }, [idx]);

  if (times.length === 0) return null;

  const handleInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    setIsUserScrubbing(true);
    setIdx(parseInt(e.target.value, 10));
  };

  const handleMouseUp = () => {
    // After user interaction, keep manual position until next drag.
    setTimeout(() => setIsUserScrubbing(false), 2000);
  };

  const rel = formatDistanceToNowStrict(new Date(times[idx]), {
    addSuffix: true,
  });

  return (
    <div className="flex flex-col space-y-2 w-full">
      <input
        type="range"
        min={0}
        max={times.length - 1}
        step={1}
        value={idx}
        onChange={handleInput}
        onMouseUp={handleMouseUp}
        className="w-full accent-blue-500 cursor-pointer"
      />
      <div className="text-xs text-gray-500 text-center">
        {idx + 1}/{times.length} – {rel}
      </div>
    </div>
  );
};

export default TimelineScrubber;
