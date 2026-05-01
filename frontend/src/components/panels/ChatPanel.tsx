import React, { useState } from 'react';

interface ChatPanelProps {
  onCreateProject: (projectName: string) => void;
}

// Validate project name: letters, numbers, hyphens, underscores only
function isValidProjectName(name: string): boolean {
  return /^[a-zA-Z0-9_-]+$/.test(name);
}

export const ChatPanel: React.FC<ChatPanelProps> = ({ onCreateProject }) => {
  const [input, setInput] = useState('');

  const handleSubmit = () => {
    const trimmed = input.trim();
    // Parse "new <project-name>" command
    const newMatch = trimmed.match(/^new\s+(\S+)$/i);
    if (newMatch) {
      const projectName = newMatch[1];
      if (isValidProjectName(projectName)) {
        onCreateProject(projectName);
      } else {
        alert('Invalid project name. Use only letters, numbers, hyphens, and underscores.');
      }
      setInput('');
      return;
    }
    // Handle other chat input (not implemented yet)
    setInput('');
  };

  return (
    <div className="chat-panel">
      <div className="chat-messages">
        {/* Chat messages will render here */}
      </div>
      <div className="chat-input">
        <input
          type="text"
          placeholder="Type a message or 'new <project-name>'..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSubmit()}
        />
        <button onClick={handleSubmit}>Send</button>
      </div>
    </div>
  );
};
