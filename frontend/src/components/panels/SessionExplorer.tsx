import React, { useState, useCallback } from 'react';

interface SessionExplorerProps {
  onCreateProject: (projectName: string) => void;
}

export const SessionExplorer: React.FC<SessionExplorerProps> = ({ onCreateProject }) => {
  const [projectName, setProjectName] = useState('');
  const [isCreating, setIsCreating] = useState(false);

  const handleCreate = useCallback(() => {
    const trimmed = projectName.trim();
    if (trimmed && !isCreating) {
      setIsCreating(true);
      onCreateProject(trimmed);
      setProjectName('');
      setTimeout(() => setIsCreating(false), 2000);
    }
  }, [projectName, isCreating, onCreateProject]);

  return (
    <div className="session-explorer">
      <h3>Workspace</h3>
      <div className="new-project-form">
        <input
          type="text"
          placeholder="Project name"
          value={projectName}
          onChange={(e) => setProjectName(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
        />
        <button onClick={handleCreate} disabled={isCreating}>
          {isCreating ? 'Creating...' : 'New Project'}
        </button>
      </div>
    </div>
  );
};
