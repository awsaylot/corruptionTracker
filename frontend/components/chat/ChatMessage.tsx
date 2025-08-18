import React from 'react';
import { Message, MessageStatus } from '@/types/chat';

interface ChatMessageProps {
    message: Message;
    onRetry?: () => void;
}

export default function ChatMessage({ message, onRetry }: ChatMessageProps) {
    const isUser = message.role === 'user';
    
    const getStatusText = (status: MessageStatus) => {
        switch (status) {
            case 'sending': return 'Sending...';
            case 'sent': return 'Sent';
            case 'error': return 'Error';
            default: return '';
        }
    };

    const dots = message.status === 'sending' && (
        <span className="inline-block">
            <span className="animate-[loading_1s_ease-in-out_infinite]">.</span>
            <span className="animate-[loading_1s_ease-in-out_0.2s_infinite]">.</span>
            <span className="animate-[loading_1s_ease-in-out_0.4s_infinite]">.</span>
        </span>
    );

    return (
        <div className={`flex ${isUser ? 'justify-end' : 'justify-start'} mb-4`}>
            <div 
                className={`
                    p-3 rounded-lg max-w-[80%]
                    ${isUser 
                        ? 'bg-blue-500 text-white' 
                        : 'bg-gray-100 text-gray-900 dark:bg-gray-700 dark:text-white'
                    }
                `}
            >
                <pre className="whitespace-pre-wrap break-words font-mono text-sm overflow-x-auto">{message.content}</pre>
                {message.status !== 'sent' && (
                    <div className="mt-1">
                        <div className="flex items-center text-xs opacity-75">
                            {getStatusText(message.status)}
                            {dots}
                        </div>
                        {message.status === 'error' && onRetry && (
                            <button 
                                onClick={onRetry} 
                                className="text-blue-500 dark:text-blue-300 text-sm mt-1 hover:underline"
                            >
                                Retry
                            </button>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
}