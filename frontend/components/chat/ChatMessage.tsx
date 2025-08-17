import React from 'react';
import styled from 'styled-components';

interface ChatMessageProps {
    message: {
        content: string;
        role: 'user' | 'assistant' | 'system';
        status: 'sending' | 'sent' | 'complete' | 'error';
        error?: string;
    };
    onRetry?: () => void;
}

const MessageContainer = styled.div<{ role: 'user' | 'assistant' | 'system' }>`
    padding: 12px;
    margin: 8px;
    border-radius: 8px;
    max-width: 80%;
    align-self: ${props => props.role === 'user' ? 'flex-end' : 'flex-start'};
    background-color: ${props => props.role === 'user' ? '#007AFF' : '#F0F0F0'};
    color: ${props => props.role === 'user' ? 'white' : 'black'};
`;

const StatusIndicator = styled.span<{ status: 'sending' | 'sent' | 'complete' | 'error' }>`
    font-size: 12px;
    color: ${props => props.status === 'error' ? 'red' : '#666'};
    margin-top: 4px;
    display: block;
`;

export function ChatMessage({ message, onRetry }: ChatMessageProps) {
    return (
        <MessageContainer role={message.role}>
            {message.content}
            {message.status !== 'complete' && (
                <div>
                    <StatusIndicator status={message.status}>
                        {message.status === 'sending' ? 'Sending...' : 
                         message.status === 'sent' ? 'Sent' : 'Error'}
                    </StatusIndicator>
                    {message.error && <div className="text-red-500 text-sm mt-1">{message.error}</div>}
                    {message.status === 'error' && onRetry && (
                        <button onClick={onRetry} className="text-blue-500 text-sm mt-1 hover:underline">
                            Retry
                        </button>
                    )}
                </div>
            )}
        </MessageContainer>
    );
}