import React, { useState, useRef, useEffect, useCallback } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { exportMonologueAsMarkdown } from '../../utils/monologue-export';
import type { MonologueMessage } from '../../hooks/useWebSocket';
import { useAnnounce } from '../../hooks/use-announce';
import { useRtl } from '../../hooks/use-rtl';
import { EmptyState } from '../ui/EmptyState';

interface MonologuePanelProps {
  collapsed: boolean;
  onToggleCollapse: () => void;
  messages: MonologueMessage[];
  isStreaming?: boolean;
  onClear?: () => void;
  sessionId?: string;
}

export const MonologuePanel: React.FC<MonologuePanelProps> = ({
  collapsed,
  onToggleCollapse,
  messages,
  isStreaming = false,
  onClear,
  sessionId,
}) => {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const [exporting, setExporting] = useState(false);
  const isRtl = useRtl();
  const announce = useAnnounce();

  // Announce panel open/close (Subtask 3.6)
  useEffect(() => {
    if (collapsed) {
      announce('LLM Monologue Panel closed', 'polite');
    } else {
      announce('LLM Monologue Panel opened', 'polite');
    }
  }, [collapsed, announce]);

  // Pre-compute display messages to avoid < in JSX context
  const displayMessages = messages.length > 100
    ? messages.slice(messages.length - 100)
    : messages;

  useEffect(() => {
    if (autoScroll && messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'instant' });
    }
  }, [messages.length, autoScroll]);

  const handleExport = useCallback(async () => {
    setExporting(true);
    // Defer export to next tick so React flushes the exporting state
    await Promise.resolve();
    exportMonologueAsMarkdown(messages, sessionId);
    setExporting(false);
  }, [messages, sessionId]);

  const handleClear = () => {
    if (onClear && (messages.length === 0 || window.confirm('Clear all monologue history?'))) {
      onClear();
    }
  };

  return (
    <Dialog.Root open={!collapsed} onOpenChange={(open) => {
      if (!open) onToggleCollapse();
    }}>
      <Dialog.Trigger asChild>
        <button
          onClick={onToggleCollapse}
          title="Toggle Monologue Panel (M)"
          style={{
            position: 'fixed',
            right: collapsed ? '8px' : 'calc(var(--monologue-panel-width, 400px) + 8px)',
            bottom: '16px',
            zIndex: 50,
            background: 'var(--bg-secondary)',
            border: '1px solid var(--bg-tertiary)',
            borderRadius: '6px',
            padding: '6px 10px',
            cursor: 'pointer',
            fontSize: '12px',
            color: 'var(--text-secondary)',
            transition: 'right 0.3s ease',
          }}
        >
          {collapsed ? 'Show Monologue' : 'Hide Monologue'}
        </button>
      </Dialog.Trigger>

      <Dialog.Portal>
        <Dialog.Overlay
          style={{
            background: 'rgba(0, 0, 0, 0.4)',
            position: 'fixed',
            inset: 0,
            zIndex: 40,
            opacity: collapsed ? 0 : 1,
            transition: 'opacity 0.3s ease',
          }}
        />
        <Dialog.Content
          className="monologue-panel-content"
          style={{
            position: 'fixed',
            top: 0,
            right: isRtl ? 'auto' : 0,
            left: isRtl ? 0 : 'auto',
            bottom: 0,
            width: 'var(--monologue-panel-width, 400px)',
            background: 'var(--bg-secondary)',
            borderLeft: isRtl ? 'none' : '1px solid var(--bg-tertiary)',
            borderRight: isRtl ? '1px solid var(--bg-tertiary)' : 'none',
            display: 'flex',
            flexDirection: 'column',
            zIndex: 50,
            transform: collapsed ? (isRtl ? 'translateX(-100%)' : 'translateX(100%)') : 'translateX(0)',
            transition: 'transform 0.3s ease',
            boxShadow: isRtl ? '4px 0 24px rgba(0,0,0,0.15)' : '-4px 0 24px rgba(0,0,0,0.15)',
          }}
          role="dialog"
          aria-modal="true"
          aria-label="LLM Monologue Panel, open"
        >
          <Dialog.Title style={{
            position: 'absolute',
            width: '1px',
            height: '1px',
            padding: 0,
            margin: '-1px',
            overflow: 'hidden',
            clip: 'rect(0, 0, 0, 0)',
            whiteSpace: 'nowrap',
            border: 0,
          }}>LLM Inner Monologue</Dialog.Title>
          <Dialog.Description style={{
            position: 'absolute',
            width: '1px',
            height: '1px',
            padding: 0,
            margin: '-1px',
            overflow: 'hidden',
            clip: 'rect(0, 0, 0, 0)',
            whiteSpace: 'nowrap',
            border: 0,
          }}>
            LLM inner monologue streaming panel showing chain-of-thought tokens during graph execution.
          </Dialog.Description>

          {/* Header */}
          <div
            style={{
              padding: '12px 16px',
              borderBottom: '1px solid var(--bg-tertiary)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <h3 style={{ margin: 0, fontSize: '14px', fontWeight: 600 }}>
              LLM Thoughts
              {isStreaming && (
                <span
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: '6px',
                    marginLeft: '8px',
                    fontSize: '11px',
                    color: '#ef4444',
                    fontWeight: 500,
                  }}
                >
                  <span
                    style={{
                      display: 'inline-block',
                      width: '8px',
                      height: '8px',
                      borderRadius: '50%',
                      background: '#ef4444',
                      animation: 'pulse-recording 1.5s infinite',
                    }}
                    aria-hidden="true"
                  />
                  <span aria-label="Recording in progress">REC</span>
                </span>
              )}
            </h3>
            <div style={{ display: 'flex', gap: '8px' }}>
              <button
                onClick={handleExport}
                disabled={exporting || messages.length === 0}
                title="Export as Markdown"
                style={{
                  background: 'none',
                  border: '1px solid var(--bg-tertiary)',
                  color: 'var(--text-secondary)',
                  cursor: 'pointer',
                  fontSize: '12px',
                  padding: '4px 8px',
                  borderRadius: '4px',
                }}
              >
                Export
              </button>
              <button
                onClick={handleClear}
                disabled={messages.length === 0}
                title="Clear history"
                style={{
                  background: 'none',
                  border: '1px solid var(--bg-tertiary)',
                  color: 'var(--text-secondary)',
                  cursor: 'pointer',
                  fontSize: '12px',
                  padding: '4px 8px',
                  borderRadius: '4px',
                }}
              >
                Clear
              </button>
              <Dialog.Close asChild>
                <button
                  title="Close (Esc)"
                  style={{
                    background: 'none',
                    border: '1px solid var(--bg-tertiary)',
                    color: 'var(--text-secondary)',
                    cursor: 'pointer',
                    fontSize: '18px',
                    padding: '0 4px',
                    borderRadius: '4px',
                  }}
                >
                  ×
                </button>
              </Dialog.Close>
            </div>
          </div>

          {/* Messages area */}
          <div
            style={{
              flex: 1,
              overflowY: 'auto',
              padding: '16px',
              fontSize: '13px',
              lineHeight: 1.6,
              textAlign: isRtl ? 'right' : 'left',
            }}
            onScroll={(e) => {
              const target = e.target as HTMLDivElement;
              const isAtBottom =
                target.scrollHeight - target.scrollTop - target.clientHeight < 50;
              setAutoScroll(isAtBottom);
            }}
            role="log"
            aria-live="polite"
            aria-label="Monologue messages"
          >
            {messages.length === 0 ? (
              <EmptyState
                icon={<span aria-hidden="true">🕭</span>}
                title="Waiting..."
                description="LLM thoughts will appear here during graph execution"
                animated
              />
            ) : (
              <>
                {messages.length > 100 && (
                  <div
                    style={{
                      fontSize: '11px',
                      color: 'var(--text-secondary)',
                      textAlign: 'center',
                      marginBottom: '8px',
                    }}
                  >
                    Showing last {displayMessages.length} of {messages.length} messages
                  </div>
                )}
                {displayMessages.map((msg) => (
                  <div
                    key={msg.id}
                    style={{
                      marginBottom: '12px',
                      padding: '8px 12px',
                      background: 'var(--bg-tertiary)',
                      borderRadius: '6px',
                      color: 'var(--text-primary)',
                    }}
                  >
                    <div
                      style={{
                        fontSize: '11px',
                        color: 'var(--text-secondary)',
                        marginBottom: '4px',
                      }}
                    >
                      {Number.isFinite(msg.timestamp)
                        ? new Date(msg.timestamp).toLocaleTimeString()
                        : '—'}
                    </div>
                    <div style={{ whiteSpace: 'pre-wrap' }}>{msg.text}</div>
                  </div>
                ))}
              </>
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* Auto-scroll toggle */}
          <div
            style={{
              padding: '8px 16px',
              borderTop: '1px solid var(--bg-tertiary)',
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              fontSize: '12px',
              color: 'var(--text-secondary)',
            }}
          >
            <input
              type="checkbox"
              checked={autoScroll}
              onChange={(e) => setAutoScroll(e.target.checked)}
              id="autoscroll-toggle"
            />
            <label htmlFor="autoscroll-toggle">Auto-scroll</label>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
};
