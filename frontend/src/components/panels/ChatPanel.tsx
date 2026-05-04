import React, { useState, useCallback, useRef, useEffect } from 'react';
import { useAnnounce } from '../../hooks/use-announce';
import { useRtl } from '../../hooks/use-rtl';

interface ChatMessage {
  id: string;
  role: 'user' | 'system';
  text: string;
  isGenerating?: boolean;
}

interface ChatPanelProps {
  collapsed: boolean;
  onToggleCollapse: () => void;
  onSendGoal: (text: string) => void;
  generating: boolean;
}

export const ChatPanel: React.FC<ChatPanelProps> = ({
  collapsed,
  onToggleCollapse,
  onSendGoal,
  generating,
}) => {
  const [input, setInput] = useState('');
  const [messages, setMessages] = useState<ChatMessage[]>([
    {
      id: '1',
      role: 'system',
      text: 'Describe your goal and I will generate a node graph for you.',
    },
  ]);
  const validationMsgRef = useRef<HTMLDivElement>(null);
  const announce = useAnnounce();
  const isRtl = useRtl();

  // Announce panel open/close (Subtask 3.6)
  useEffect(() => {
    if (collapsed) {
      announce('Chat panel closed', 'polite');
    } else {
      announce('Chat panel opened', 'polite');
    }
  }, [collapsed, announce]);

  const inputLength = input.trim().length;
  const isValidInput = inputLength >= 10 && inputLength <= 500;
  const isTooLong = inputLength > 500;

  const handleSubmit = useCallback(() => {
    const text = input.trim();
    if (!text || text.length < 10 || text.length > 500 || generating) return;

    const userMessage: ChatMessage = {
      id: crypto.randomUUID(),
      role: 'user',
      text,
    };

    setMessages((prev) => [...prev, userMessage]);
    try {
      onSendGoal(text);
    } catch {
      // Parent rejected the goal; remove the orphaned message and restore input
      setMessages((prev) => prev.slice(0, -1));
      setInput(text);
    }
    setInput('');
  }, [input, generating, onSendGoal]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        handleSubmit();
      }
    },
    [handleSubmit]
  );

  return (
    <div className={`chat-panel ${collapsed ? 'collapsed' : ''}`}>
      <div className="chat-header">
        <h3>Chat</h3>
        <button
          className="collapse-btn"
          onClick={onToggleCollapse}
          title={collapsed ? 'Expand chat' : 'Collapse chat'}
          aria-label={collapsed ? 'Expand chat' : 'Collapse chat'}
          style={{ transform: isRtl ? 'scaleX(-1)' : undefined }}
        >
          {collapsed ? (isRtl ? '←' : '→') : (isRtl ? '→' : '←')}
        </button>
      </div>

      <div className="chat-messages" role="log" aria-live="polite" style={{ textAlign: isRtl ? 'right' : 'left' }}>
        {messages.map((msg) => (
          <div key={msg.id} className={`chat-message ${msg.role} ${msg.isGenerating ? 'generating' : ''}`}>
            {msg.isGenerating ? (
              <>
                {msg.text}
                <span className="generating-ellipsis"></span>
              </>
            ) : (
              msg.text
            )}
          </div>
        ))}
        {generating && (
          <div className="chat-message system generating">
            Generating graph<span className="generating-ellipsis"></span>
          </div>
        )}
      </div>

      <div className="chat-input">
        <input
          type="text"
          placeholder="Describe your goal..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={generating}
          aria-label="Goal description"
          aria-describedby="chat-validation-msg"
        />
        <button onClick={handleSubmit} disabled={!isValidInput || generating}>
          Send
        </button>
      </div>
      {!isValidInput && input.length > 0 && (
        <div id="chat-validation-msg" ref={validationMsgRef} style={{ padding: '0 12px 8px', fontSize: '12px', color: isTooLong ? 'var(--error)' : 'var(--text-secondary)' }}>
          {isTooLong
            ? `Maximum 500 characters allowed (${inputLength}/500)`
            : `Minimum 10 characters required (${inputLength}/10)`}
        </div>
      )}
    </div>
  );
};
