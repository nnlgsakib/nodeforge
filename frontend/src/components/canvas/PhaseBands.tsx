import React from 'react';
import { useViewport } from '@xyflow/react';

interface PhaseBand {
  label: string;
  color: string;
}

const PHASE_BANDS: PhaseBand[] = [
  { label: 'Discovery', color: '#3B82F6' },
  { label: 'Execution', color: '#F97316' },
  { label: 'Recovery', color: '#EF4444' },
  { label: 'Completion', color: '#22C55E' },
];

// Phase band width in graph coordinates
const BAND_WIDTH = 400;
const BAND_HEIGHT = 32;
const GAP = 20;

// PhaseBands renders color-coded bands across the canvas top.
// It uses useViewport so the bands stay visible at all zoom levels.
export const PhaseBands: React.FC = () => {
  const { x, y, zoom } = useViewport();

  // Compute the visible horizontal range in graph coordinates
  const viewportWidth = window.innerWidth / Math.max(zoom, 0.1);
  const startX = Math.floor((x - 100) / (BAND_WIDTH + GAP)) * (BAND_WIDTH + GAP);
  const endX = x + viewportWidth + 200;

  const bands: React.ReactNode[] = [];
  for (let bx = startX; bx < endX; bx += BAND_WIDTH + GAP) {
    const bandIndex = Math.round(((bx - startX) / (BAND_WIDTH + GAP)) % PHASE_BANDS.length);
    const band = PHASE_BANDS[((bandIndex % PHASE_BANDS.length) + PHASE_BANDS.length) % PHASE_BANDS.length];

    bands.push(
      <g key={`phase-${bx}`}>
        <rect
          x={bx}
          y={y - BAND_HEIGHT - GAP}
          width={BAND_WIDTH}
          height={BAND_HEIGHT}
          fill={band.color}
          opacity={0.08}
          rx={4}
        />
        <text
          x={bx + BAND_WIDTH / 2}
          y={y - BAND_HEIGHT - GAP + BAND_HEIGHT / 2 + 4}
          textAnchor="middle"
          fill={band.color}
          opacity={0.6}
          fontSize={Math.max(10, 12 / Math.max(zoom, 0.1))}
          fontWeight={500}
          style={{ pointerEvents: 'none', userSelect: 'none' }}
        >
          {band.label}
        </text>
      </g>
    );
  }

  return <g className="phase-bands">{bands}</g>;
};
