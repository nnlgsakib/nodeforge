import React, { useState, useRef, useEffect, useCallback } from 'react';

interface MonologueMessage {
  id: string;
  text: string;
  timestamp: number;
}

interface MonologuePanelProps {
  collapsed: boolean;
  onToggleCollapse: () => void;
  messages: MonologueMessage[];
  isStreaming?: boolean;
}

export const MonologuePanel: React.FC<MonologuePanelProps> = ({
  collapsed,
  onToggleCollapse,
  messages,
  isStreaming = false,
}) => {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const [exporting, setExporting] = useState(false);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (autoScroll && messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, autoScroll]);

  const handleExport = useCallback(() => {
    setExporting(true);
    const history = messages
      .map(
        (m) =>
          `[${new Date(m.timestamp).toLocaleTimeString()}] ${m.text}`
      )
      .join('\n\n');
    const blob = new Blob([history], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `monologue-${new Date().toISOString().slice(0, 10)}.txt`;
    a.click();
    URL.revokeObjectURL(url);
    setExporting(false);
  }, [messages]);

  const handleClear = useCallback(() => {
    // This would need a callback prop to actually clear
    console.log('Clear monologue history');
  }, []);

  return (
    <div
      className={`monologue-panel ${collapsed ? 'collapsed' : ''}`}
      style={{
        width: '400px',
        minWidth: '400px',
        background: 'var(--bg-secondary)',
        borderLeft: '1px solid var(--bg-tertiary)',
        display: 'flex',
        flexDirection: 'column',
        transition: 'margin-right 0.3s ease',
        marginRight: collapsed ? '-400px' : '0',
      }}
    >
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
                display: 'inline-block',
                width: '8px',
                height: '8px',
                borderRadius: '50%',
                background: 'var(--accent)',
                marginLeft: '8px',
                animation: 'pulse 1.5s infinite',
              }}
            />
          )}
        </h3>
        <div style={{ display: 'flex', gap: '8px' }}>
          <button
            onClick={handleExport}
            disabled={exporting || messages.length === 0}
            title="Export history"
            style={{
              background: 'none',
              border: 'none',
              color: 'var(--text-secondary)',
              cursor: 'pointer',
              fontSize: '12px',
              padding: '4px 8px',
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
              border: 'none',
              color: 'var(--text-secondary)',
              cursor: 'pointer',
              fontSize: '12px',
              padding: '4px 8px',
            }}
          >
            Clear
          </button>
          <button
            className="collapse-btn"
            onClick={onToggleCollapse}
            title={collapsed ? 'Expand monologue' : 'Collapse monologue'}
            style={{
              background: 'none',
              border: 'none',
              color: 'var(--text-secondary)',
              cursor: 'pointer',
              fontSize: '18px',
              padding: '0 4px',
            }}
          >
            {collapsed ? '→' : '←'}
          </button>
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
        }}
        onScroll={(e) => {
          const target = e.target as HTMLDivElement;
          const isAtBottom =
            target.scrollHeight - target.scrollTop - target.clientHeight < 50;
          setAutoScroll(isAtBottom);
        }}
      >
        {messages.length === 0 ? (
          <div
            style={{
              color: 'var(--text-secondary)',
              fontStyle: 'italic',
              textAlign: 'center',
              marginTop: '40px',
            }}
          >
            LLM thoughts will appear here during graph execution...
          </div>
        ) : (
          messages.map((msg) => (
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
                {new Date(msg.timestamp).toLocaleTimeString()}
              </div>
              <div style={{ whiteSpace: 'pre-wrap' }}>{msg.text}</div>
            </div>
          ))
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

      {/* Add pulse animation */}
      <style>{`
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.3; }
        }
      `}</style>
    </div>
  );
};
