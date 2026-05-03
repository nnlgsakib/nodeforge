import React, { useState, useEffect, useCallback } from 'react';
import * as Dialog from '@radix-ui/react-dialog';

interface NodeConfigValues {
  timeout: number;
  retryCount: number;
  tokenBudget: number;
}

interface NodeConfigErrors {
  timeout?: string;
  retryCount?: string;
  tokenBudget?: string;
}

interface NodeConfigProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeId: string | null;
  initialConfig?: Partial<NodeConfigValues>;
  onSave: (nodeId: string, config: NodeConfigValues) => void;
}

const DEFAULT_TIMEOUT = 60;
const DEFAULT_RETRY_COUNT = 3;
const DEFAULT_TOKEN_BUDGET = 10000;

export const NodeConfig: React.FC<NodeConfigProps> = ({
  open,
  onOpenChange,
  nodeId,
  initialConfig,
  onSave,
}) => {
  const [timeout, setTimeout] = useState(initialConfig?.timeout ?? DEFAULT_TIMEOUT);
  const [retryCount, setRetryCount] = useState(initialConfig?.retryCount ?? DEFAULT_RETRY_COUNT);
  const [tokenBudget, setTokenBudget] = useState(initialConfig?.tokenBudget ?? DEFAULT_TOKEN_BUDGET);
  const [errors, setErrors] = useState<NodeConfigErrors>({});
  const [saving, setSaving] = useState(false);

  // Reset form when dialog opens with new node
  useEffect(() => {
    if (open && nodeId) {
      setTimeout(initialConfig?.timeout ?? DEFAULT_TIMEOUT);
      setRetryCount(initialConfig?.retryCount ?? DEFAULT_RETRY_COUNT);
      setTokenBudget(initialConfig?.tokenBudget ?? DEFAULT_TOKEN_BUDGET);
      setErrors({});
      setSaving(false);
    }
  }, [open, nodeId, initialConfig]);

  // Real-time validation
  const validate = useCallback((): NodeConfigErrors => {
    const newErrors: NodeConfigErrors = {};

    if (!Number.isFinite(timeout) || timeout < 1 || timeout > 300) {
      newErrors.timeout = 'Timeout must be 1-300 seconds';
    }
    if (!Number.isFinite(retryCount) || retryCount < 0 || retryCount > 10) {
      newErrors.retryCount = 'Retry count must be 0-10';
    }
    if (!Number.isFinite(tokenBudget) || tokenBudget < 100 || tokenBudget > 100000) {
      newErrors.tokenBudget = 'Budget must be 100-100000 tokens';
    }

    return newErrors;
  }, [timeout, retryCount, tokenBudget]);

  useEffect(() => {
    if (open) {
      setErrors(validate());
    }
  }, [timeout, retryCount, tokenBudget, open, validate]);

  const handleSave = useCallback(async () => {
    if (!nodeId) return;

    const validationErrors = validate();
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    setSaving(true);
    try {
      onSave(nodeId, { timeout, retryCount, tokenBudget });
    } finally {
      setSaving(false);
    }
  }, [nodeId, timeout, retryCount, tokenBudget, validate, onSave]);

  const isSaveDisabled = Object.keys(errors).length > 0 || saving || !nodeId;

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay
          style={{
            backgroundColor: 'rgba(0, 0, 0, 0.6)',
            position: 'fixed',
            inset: 0,
            zIndex: 1000,
          }}
        />
        <Dialog.Content
          style={{
            position: 'fixed',
            top: '50%',
            left: '50%',
            transform: 'translate(-50%, -50%)',
            background: 'var(--bg-secondary)',
            border: '1px solid var(--bg-tertiary)',
            borderRadius: '12px',
            padding: '24px',
            width: '90%',
            maxWidth: '420px',
            zIndex: 1001,
            boxShadow: '0 20px 60px rgba(0, 0, 0, 0.5)',
          }}
        >
          <Dialog.Title
            style={{
              margin: '0 0 4px 0',
              fontSize: '16px',
              fontWeight: 600,
              color: 'var(--text-primary)',
            }}
          >
            Node Configuration
          </Dialog.Title>
          <Dialog.Description
            style={{
              margin: '0 0 20px 0',
              fontSize: '13px',
              color: 'var(--text-secondary)',
            }}
          >
            Configure execution parameters for this node
          </Dialog.Description>

          {/* Timeout Input */}
          <div style={{ marginBottom: '16px' }}>
            <label
              htmlFor="node-timeout"
              style={{
                display: 'block',
                fontSize: '12px',
                fontWeight: 500,
                color: 'var(--text-secondary)',
                marginBottom: '6px',
              }}
            >
              Timeout (seconds)
            </label>
            <input
              id="node-timeout"
              type="number"
              min={1}
              max={300}
              value={timeout}
              onChange={(e) => setTimeout(Number(e.target.value))}
              style={{
                width: '100%',
                padding: '8px 10px',
                border: errors.timeout ? '1px solid var(--error)' : '1px solid var(--bg-tertiary)',
                borderRadius: '6px',
                background: 'var(--bg-primary)',
                color: 'var(--text-primary)',
                fontSize: '13px',
                boxSizing: 'border-box',
              }}
            />
            {errors.timeout && (
              <span style={{ fontSize: '11px', color: 'var(--error)', marginTop: '4px', display: 'block' }}>
                {errors.timeout}
              </span>
            )}
          </div>

          {/* Retry Count Input */}
          <div style={{ marginBottom: '16px' }}>
            <label
              htmlFor="node-retry-count"
              style={{
                display: 'block',
                fontSize: '12px',
                fontWeight: 500,
                color: 'var(--text-secondary)',
                marginBottom: '6px',
              }}
            >
              Retry Count
            </label>
            <input
              id="node-retry-count"
              type="number"
              min={0}
              max={10}
              value={retryCount}
              onChange={(e) => setRetryCount(Number(e.target.value))}
              style={{
                width: '100%',
                padding: '8px 10px',
                border: errors.retryCount ? '1px solid var(--error)' : '1px solid var(--bg-tertiary)',
                borderRadius: '6px',
                background: 'var(--bg-primary)',
                color: 'var(--text-primary)',
                fontSize: '13px',
                boxSizing: 'border-box',
              }}
            />
            {errors.retryCount && (
              <span style={{ fontSize: '11px', color: 'var(--error)', marginTop: '4px', display: 'block' }}>
                {errors.retryCount}
              </span>
            )}
          </div>

          {/* Token Budget Input */}
          <div style={{ marginBottom: '20px' }}>
            <label
              htmlFor="node-token-budget"
              style={{
                display: 'block',
                fontSize: '12px',
                fontWeight: 500,
                color: 'var(--text-secondary)',
                marginBottom: '6px',
              }}
            >
              Token Budget
            </label>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
              <input
                type="range"
                min={100}
                max={100000}
                step={100}
                value={tokenBudget}
                onChange={(e) => setTokenBudget(Number(e.target.value))}
                style={{ flex: 1 }}
              />
              <input
                id="node-token-budget"
                type="number"
                min={100}
                max={100000}
                value={tokenBudget}
                onChange={(e) => setTokenBudget(Number(e.target.value))}
                style={{
                  width: '100px',
                  padding: '8px 10px',
                  border: errors.tokenBudget ? '1px solid var(--error)' : '1px solid var(--bg-tertiary)',
                  borderRadius: '6px',
                  background: 'var(--bg-primary)',
                  color: 'var(--text-primary)',
                  fontSize: '13px',
                  boxSizing: 'border-box',
                }}
              />
            </div>
            {errors.tokenBudget && (
              <span style={{ fontSize: '11px', color: 'var(--error)', marginTop: '4px', display: 'block' }}>
                {errors.tokenBudget}
              </span>
            )}
          </div>

          {/* Action Buttons */}
          <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
            <Dialog.Close asChild>
              <button
                style={{
                  padding: '8px 16px',
                  background: 'var(--bg-tertiary)',
                  color: 'var(--text-secondary)',
                  border: '1px solid var(--bg-tertiary)',
                  borderRadius: '6px',
                  cursor: 'pointer',
                  fontSize: '13px',
                  transition: 'color 200ms, border-color 200ms',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.color = 'var(--text-primary)';
                  e.currentTarget.style.borderColor = 'var(--accent)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.color = 'var(--text-secondary)';
                  e.currentTarget.style.borderColor = 'var(--bg-tertiary)';
                }}
              >
                Cancel
              </button>
            </Dialog.Close>
            <button
              onClick={handleSave}
              disabled={isSaveDisabled}
              style={{
                padding: '8px 16px',
                background: isSaveDisabled ? 'var(--bg-tertiary)' : 'var(--accent)',
                color: 'white',
                border: 'none',
                borderRadius: '6px',
                cursor: isSaveDisabled ? 'not-allowed' : 'pointer',
                fontSize: '13px',
                fontWeight: 500,
                transition: 'background-color 200ms',
              }}
            >
              {saving ? 'Saving...' : 'Save'}
            </button>
          </div>

          {/* Close button */}
          <Dialog.Close asChild>
            <button
              style={{
                position: 'absolute',
                top: '12px',
                right: '12px',
                background: 'transparent',
                border: 'none',
                color: 'var(--text-secondary)',
                cursor: 'pointer',
                fontSize: '18px',
                padding: '4px',
                lineHeight: 1,
              }}
              aria-label="Close"
            >
              ×
            </button>
          </Dialog.Close>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
};
