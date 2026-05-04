import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ChatPanel } from './ChatPanel';

describe('ChatPanel', () => {
  const defaultProps = {
    collapsed: false,
    onToggleCollapse: vi.fn(),
    onSendGoal: vi.fn(),
    generating: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('rendering', () => {
    it('should render chat header and input', () => {
      render(<ChatPanel {...defaultProps} />);
      expect(screen.getByRole('heading', { name: /chat/i })).toBeTruthy();
      expect(screen.getByPlaceholderText('Describe your goal...')).toBeTruthy();
    });

    it('should render initial system message', () => {
      render(<ChatPanel {...defaultProps} />);
      expect(screen.getByText('Describe your goal and I will generate a node graph for you.')).toBeTruthy();
    });

    it('should show collapse button', () => {
      render(<ChatPanel {...defaultProps} />);
      const btn = screen.getByTitle('Collapse chat');
      expect(btn).toBeTruthy();
    });
  });

  describe('input validation', () => {
    it('should disable send button when input is less than 10 characters', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: 'short' } });
      const sendBtn = screen.getByRole('button', { name: 'Send' });
      expect(sendBtn).toBeDisabled();
    });

    it('should enable send button when input is 10+ characters', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: 'this is a valid goal' } });
      const sendBtn = screen.getByRole('button', { name: 'Send' });
      expect(sendBtn).not.toBeDisabled();
    });

    it('should disable send button when input exceeds 500 characters', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      const longText = 'a'.repeat(501);
      fireEvent.change(input, { target: { value: longText } });
      const sendBtn = screen.getByRole('button', { name: 'Send' });
      expect(sendBtn).toBeDisabled();
    });

    it('should enable send button at exactly 500 characters', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: 'a'.repeat(500) } });
      const sendBtn = screen.getByRole('button', { name: 'Send' });
      expect(sendBtn).not.toBeDisabled();
    });

    it('should not submit whitespace-only input', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: '          ' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(defaultProps.onSendGoal).not.toHaveBeenCalled();
    });

    it('should not submit when Shift+Enter is pressed', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: 'this is a valid goal' } });
      fireEvent.keyDown(input, { key: 'Enter', shiftKey: true });

      expect(defaultProps.onSendGoal).not.toHaveBeenCalled();
    });

    it('should show min character message when input is 1-9 chars', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: 'short' } });
      expect(screen.getByText(/Minimum 10 characters required/)).toBeTruthy();
    });

    it('should show max character message when input exceeds 500', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: 'a'.repeat(501) } });
      expect(screen.getByText(/Maximum 500 characters allowed/)).toBeTruthy();
    });

    it('should show character count in validation message', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: 'a'.repeat(510) } });
      expect(screen.getByText(/510\/500/)).toBeTruthy();
    });

    it('should have aria-label on input', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      expect(input).toHaveAttribute('aria-label', 'Goal description');
    });

    it('should have aria-describedby on input linked to validation message', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      expect(input).toHaveAttribute('aria-describedby', 'chat-validation-msg');
    });
  });

  describe('goal submission', () => {
    it('should call onSendGoal and add user message on Enter key', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: 'this is a valid goal text' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(defaultProps.onSendGoal).toHaveBeenCalledWith('this is a valid goal text');
      expect(screen.getByText('this is a valid goal text')).toBeTruthy();
    });

    it('should not submit when input is less than 10 characters', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: 'short' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(defaultProps.onSendGoal).not.toHaveBeenCalled();
    });

    it('should not submit when generating is true', () => {
      render(<ChatPanel {...defaultProps} generating={true} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: 'this is a valid goal text' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(defaultProps.onSendGoal).not.toHaveBeenCalled();
    });

    it('should clear input after submission', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: 'this is a valid goal text' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect((input as HTMLInputElement).value).toBe('');
    });

    it('should trim whitespace from input before submission', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: '  valid goal text here  ' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(defaultProps.onSendGoal).toHaveBeenCalledWith('valid goal text here');
    });
  });

  describe('generating state', () => {
    it('should disable input when generating is true', () => {
      render(<ChatPanel {...defaultProps} generating={true} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      expect(input).toBeDisabled();
    });

    it('should disable send button when generating is true', () => {
      render(<ChatPanel {...defaultProps} generating={true} />);
      const sendBtn = screen.getByRole('button', { name: 'Send' });
      expect(sendBtn).toBeDisabled();
    });

    it('should show generating message overlay', () => {
      render(<ChatPanel {...defaultProps} generating={true} />);
      expect(screen.getByText(/Generating graph/)).toBeTruthy();
    });
  });

  describe('collapse behavior', () => {
    it('should call onToggleCollapse when collapse button clicked', () => {
      render(<ChatPanel {...defaultProps} />);
      const btn = screen.getByTitle('Collapse chat');
      fireEvent.click(btn);
      expect(defaultProps.onToggleCollapse).toHaveBeenCalled();
    });

    it('should apply collapsed class when collapsed', () => {
      const { container } = render(<ChatPanel {...defaultProps} collapsed={true} />);
      const panel = container.firstChild as HTMLElement;
      expect(panel.classList.contains('collapsed')).toBe(true);
    });
  });

  describe('chat message history', () => {
    it('should display user messages after submission', () => {
      render(<ChatPanel {...defaultProps} />);
      const input = screen.getByPlaceholderText('Describe your goal...');
      fireEvent.change(input, { target: { value: 'first goal' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      fireEvent.change(input, { target: { value: 'second goal update' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(screen.getByText('first goal')).toBeTruthy();
      expect(screen.getByText('second goal update')).toBeTruthy();
    });
  });

  describe('new project button', () => {
    it('should render New Project button when onNewProject is provided', () => {
      const handleNewProject = vi.fn();
      render(<ChatPanel {...defaultProps} onNewProject={handleNewProject} />);

      expect(screen.getByRole('button', { name: /create new project/i })).toBeInTheDocument();
    });

    it('should not render New Project button when onNewProject is not provided', () => {
      render(<ChatPanel {...defaultProps} />);

      expect(screen.queryByRole('button', { name: /create new project/i })).not.toBeInTheDocument();
    });

    it('should call onNewProject when button is clicked', () => {
      const handleNewProject = vi.fn();
      render(<ChatPanel {...defaultProps} onNewProject={handleNewProject} />);

      const newProjectBtn = screen.getByRole('button', { name: /create new project/i });
      fireEvent.click(newProjectBtn);

      expect(handleNewProject).toHaveBeenCalledTimes(1);
    });
  });
});
