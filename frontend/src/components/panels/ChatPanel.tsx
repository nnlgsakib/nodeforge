import React, { useState, useCallback } from 'react';

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

  const isValidInput = input.trim().length >= 10;

  const handleSubmit = () => {
    const text = input.trim();
    if (!text || text.length < 10 || generating) return;

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      text,
    };

    setMessages((prev) => [...prev, userMessage]);
    onSendGoal(text);
    setInput('');
  };

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
        >
          {collapsed ? '→' : '←'}
        </button>
      </div>

      <div className="chat-messages">
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
          minLength={10}
        />
        <button onClick={handleSubmit} disabled={!isValidInput || generating}>
          Send
        </button>
      </div>
      {!isValidInput && input.length > 0 && (
        <div style={{ padding: '0 12px 8px', fontSize: '12px', color: 'var(--error)' }}>
          Minimum 10 characters required
        </div>
      )}
    </div>
  );
};
