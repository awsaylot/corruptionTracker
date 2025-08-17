import { useState, FormEvent } from 'react';
import { useSimpleChat } from '@/hooks/useSimpleChat';
import ChatMessage from '@/components/chat/ChatMessage';
import LoadingSpinner from '@/components/ui/LoadingSpinner';

export default function SimpleChat() {
  const { messages, sendMessage, isLoading, error } = useSimpleChat();
  const [input, setInput] = useState('');

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!input.trim() || isLoading) return;

    await sendMessage(input.trim());
    setInput('');
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 overflow-auto p-4 space-y-4">
        {messages.map((message) => (
          <ChatMessage key={message.id} message={message} />
        ))}
        {isLoading && (
          <div className="flex justify-center">
            <LoadingSpinner />
          </div>
        )}
        {error && (
          <div className="text-red-500 text-center">
            Error: {error}
          </div>
        )}
      </div>

      <form onSubmit={handleSubmit} className="p-4 border-t">
        <div className="flex space-x-4">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Type a message..."
            className="flex-1 p-2 border rounded"
            disabled={isLoading}
          />
          <button
            type="submit"
            disabled={isLoading || !input.trim()}
            className="px-4 py-2 bg-blue-500 text-white rounded disabled:opacity-50"
          >
            Send
          </button>
        </div>
      </form>
    </div>
  );
}
