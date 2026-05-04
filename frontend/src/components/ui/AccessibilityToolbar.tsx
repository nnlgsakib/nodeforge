import React, { useState, useEffect, useCallback } from 'react';
import { Switch } from './Switch';

const FONT_MIN = 12;
const FONT_MAX = 24;
const FONT_DEFAULT = 14;

interface AccessibilityToolbarProps {
  visible: boolean;
  onToggle: () => void;
}

export const AccessibilityToolbar: React.FC<AccessibilityToolbarProps> = ({ visible, onToggle }) => {
  const [highContrast, setHighContrast] = useState(false);
  const [rtlMode, setRtlMode] = useState(false);
  const [fontSize, setFontSize] = useState(() => {
    const saved = sessionStorage.getItem('accessibility-font-size');
    const parsed = saved ? parseInt(saved, 10) : FONT_DEFAULT;
    return Number.isNaN(parsed) || parsed < FONT_MIN || parsed > FONT_MAX ? FONT_DEFAULT : parsed;
  });

  // Load preferences from session storage on mount
  useEffect(() => {
    const savedContrast = sessionStorage.getItem('accessibility-high-contrast');
    const savedRtl = sessionStorage.getItem('accessibility-rtl');
    if (savedContrast === 'true') setHighContrast(true);
    if (savedRtl === 'true') setRtlMode(true);
  }, []);

  // Persist preferences and apply to DOM
  useEffect(() => {
    try {
      sessionStorage.setItem('accessibility-high-contrast', String(highContrast));
      sessionStorage.setItem('accessibility-rtl', String(rtlMode));
      sessionStorage.setItem('accessibility-font-size', String(fontSize));
    } catch {
      // sessionStorage may be unavailable in private browsing — skip persistence
    }

    // Apply high contrast to root element
    document.documentElement.classList.toggle('high-contrast', highContrast);

    // Apply RTL direction
    document.documentElement.dir = rtlMode ? 'rtl' : 'ltr';

    // Apply font size to root element
    document.documentElement.style.fontSize = `${fontSize}px`;
  }, [highContrast, rtlMode, fontSize]);

  const handleFontSizeChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const parsed = parseInt(e.target.value, 10);
    if (!Number.isNaN(parsed)) {
      setFontSize(Math.max(FONT_MIN, Math.min(FONT_MAX, parsed)));
    }
  }, []);

  if (!visible) return null;

  return (
    <div className="accessibility-toolbar" role="toolbar" aria-label="Accessibility settings">
      <button
        className="accessibility-toolbar-toggle cursor-pointer"
        onClick={onToggle}
        aria-label="Toggle accessibility toolbar"
        title="Accessibility"
      >
        &#9881;
      </button>

      <div className="accessibility-toolbar-controls">
        {/* High Contrast Toggle */}
        <div className="accessibility-control">
          <label htmlFor="high-contrast-toggle" className="accessibility-label">
            High Contrast
          </label>
          <Switch
            id="high-contrast-toggle"
            checked={highContrast}
            onCheckedChange={setHighContrast}
          />
        </div>

        {/* RTL Toggle */}
        <div className="accessibility-control">
          <label htmlFor="rtl-toggle" className="accessibility-label">
            RTL Mode
          </label>
          <Switch
            id="rtl-toggle"
            checked={rtlMode}
            onCheckedChange={setRtlMode}
          />
        </div>

        {/* Font Size Slider */}
        <div className="accessibility-control">
          <label htmlFor="font-size-slider" className="accessibility-label">
            Font Size: {fontSize}px
          </label>
          <input
            id="font-size-slider"
            type="range"
            min={FONT_MIN}
            max={FONT_MAX}
            value={fontSize}
            onChange={handleFontSizeChange}
            className="font-size-slider cursor-pointer"
            aria-valuemin={FONT_MIN}
            aria-valuemax={FONT_MAX}
            aria-valuenow={fontSize}
          />
          <div className="font-size-labels">
            <span>{FONT_MIN}px</span>
            <span>{FONT_MAX}px</span>
          </div>
        </div>
      </div>
    </div>
  );
};
