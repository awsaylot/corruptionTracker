import { useState } from 'react';
import { UI_CONSTANTS } from '../../utils/constants';

interface InputFormProps {
  onSubmit: (message: string) => void;
  isLoading: boolean;
  isConnected: boolean;
}

export function InputForm({ onSubmit, isLoading, isConnected }: InputFormProps) {
  const [input, setInput] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (input.trim() && !isLoading && isConnected) {
      onSubmit(input.trim());
      setInput('');
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="w-full">
      <div className="flex items-center gap-2">
        <textarea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={
            !isConnected 
              ? "Connecting to server..." 
              : isLoading 
              ? "Processing..." 
              : "Type your message..."
          }
          className="flex-1 min-h-[50px] p-2 border rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
          maxLength={UI_CONSTANTS.MAX_INPUT_LENGTH}
          disabled={isLoading || !isConnected}
        />
        <button 
          type="submit" 
          className={`px-4 py-2 rounded-lg font-medium ${
            !input.trim() || isLoading || !isConnected
              ? 'bg-gray-200 text-gray-500'
              : 'bg-blue-600 text-white hover:bg-blue-700'
          }`}
          disabled={!input.trim() || isLoading || !isConnected}
        >
          {isLoading ? 'Sending...' : 'Send'}
        </button>
      </div>
    </form>
  );
}