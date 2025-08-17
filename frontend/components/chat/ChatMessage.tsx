import React from 'react';
import styled from 'styled-components';

interface ChatMessageProps {
    message: {
        content: string;
        role: 'user' | 'assistant' | 'system';
        status: 'sending' | 'sent' | 'complete' | 'error' | 'streaming';
        error?: string;
    };
    onRetry?: () => void;
}

const MessageContainer = styled.div.withConfig({
    shouldForwardProp: (prop) => !['role'].includes(prop),
})<{ role: 'user' | 'assistant' | 'system' }>`
    padding: 12px;
    margin: 8px;
    border-radius: 8px;
    max-width: 80%;
    align-self: ${props => props.role === 'user' ? 'flex-end' : 'flex-start'};
    background-color: ${props => props.role === 'user' ? '#007AFF' : '#F0F0F0'};
    color: ${props => props.role === 'user' ? 'white' : 'black'};
`;

const StatusIndicator = styled.span.withConfig({
    shouldForwardProp: (prop) => !['status'].includes(prop),
})<{ status: 'sending' | 'sent' | 'complete' | 'error' | 'streaming' }>`
    font-size: 12px;
    color: ${props => props.status === 'error' ? 'red' : '#666'};
    margin-top: 4px;
    display: block;
`;

const StreamingDots = styled.span`
    &::after {
        content: '...';
        animation: dots 1.5s steps(4, end) infinite;
    }
    
    @keyframes dots {
        0%, 20% {
            color: rgba(0, 0, 0, 0);
            text-shadow:
                .25em 0 0 rgba(0, 0, 0, 0),
                .5em 0 0 rgba(0, 0, 0, 0);
        }
        40% {
            color: #666;
            text-shadow:
                .25em 0 0 rgba(0, 0, 0, 0),
                .5em 0 0 rgba(0, 0, 0, 0);
        }
        60% {
            text-shadow:
                .25em 0 0 #666,
                .5em 0 0 rgba(0, 0, 0, 0);
        }
        80%, 100% {
            text-shadow:
                .25em 0 0 #666,
                .5em 0 0 #666;
        }
    }
`;

export function ChatMessage({ message, onRetry }: ChatMessageProps) {
    const getStatusText = (status: string) => {
        switch (status) {
            case 'sending': return 'Sending...';
            case 'streaming': return 'Generating';
            case 'sent': return 'Sent';
            case 'complete': return '';
            case 'error': return 'Error';
            default: return '';
        }
    };

    return (
        <MessageContainer role={message.role}>
            <div>{message.content}</div>
            {message.status !== 'complete' && (
                <div>
                    <StatusIndicator status={message.status}>
                        {getStatusText(message.status)}
                        {message.status === 'streaming' && <StreamingDots />}
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