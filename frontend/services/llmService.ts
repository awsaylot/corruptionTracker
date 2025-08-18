import { ExtractedContent, ExtractionResponse } from '../types/extraction';

interface APIEndpoints {
  HEALTH: string;
  STATUS: string;
  CHAT_WS: string;
  EXTRACTION: string;
}

const API_ENDPOINTS: APIEndpoints = {
  HEALTH: '/api/health',
  STATUS: '/api/metrics',
  CHAT_WS: '/api/chat/ws',
  EXTRACTION: '/api/extraction'
};

export async function extractContent(url: string): Promise<ExtractedContent> {
  const baseUrl = (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080').replace(/\/$/, '');
  
  try {
    const response = await fetch(`${baseUrl}${API_ENDPOINTS.EXTRACTION}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ url })
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Failed to extract content');
    }

    const data: ExtractionResponse = await response.json();
    if (!data.success) {
      throw new Error(data.error || 'Failed to extract content');
    }

    return data.data;
  } catch (error) {
    console.error('Error extracting content:', error);
    throw error;
  }
}

export class LLMService {
  private baseUrl: string;
  private ws: WebSocket | null = null;
  private messageHandler: ((message: string) => void) | null = null;
  private readyResolver: (() => void) | null = null;
  private readyPromise: Promise<void> | null = null;
  private conversationId: string | null = null;

  constructor(baseUrl?: string) {
    this.baseUrl = (baseUrl || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080').replace(/\/$/, '');
    console.log('LLMService initialized with baseUrl:', this.baseUrl);
  }

  getConversationId(): string | null {
    return this.conversationId;
  }

  async sendMessage(content: string, messages: any[]): Promise<any> {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      await this.startConversation();
    }

    const payload = {
      type: 'message',
      content,
      messages
    };

    this.ws!.send(JSON.stringify(payload));
    return new Promise((resolve) => {
      this.messageHandler = (message) => {
        resolve(message);
      };
    });
  }

  async startConversation(): Promise<void> {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      console.log('WebSocket already open, skipping startConversation.');
      return;
    }

    const wsUrl = `${this.baseUrl.replace(/^http/, 'ws')}${API_ENDPOINTS.CHAT_WS}`;
    console.log('Connecting WebSocket to', wsUrl);
    this.ws = new WebSocket(wsUrl);

    this.readyPromise = new Promise((resolve) => {
      this.readyResolver = resolve;
    });

    this.ws.onopen = () => {
      console.log('✅ WebSocket connection opened');
      // Send start message immediately when connection opens
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'start' }));
        console.log('➡️ Sent start message to backend');
      }
    };

    this.ws.onerror = (error) => console.error('❌ WebSocket error:', error);

    this.ws.onmessage = (event) => {
      console.log('⬅️ WebSocket message received:', event.data);

      try {
        const data = JSON.parse(event.data);

        // When conversation is started, store the ID and mark as ready
        if (data.type === 'system' && data.content === 'Conversation started') {
          this.conversationId = data.conversationId;
          console.log('✅ Stored conversationId:', this.conversationId);
          this.readyResolver?.();
          return;
        }

        if (this.messageHandler) {
          // For error messages, preserve all error details including metadata
          if (data.type === 'error') {
            console.error('❌ Chat error from backend:', data.content);
            // Pass through the complete error response to preserve metadata
            this.messageHandler(JSON.stringify({
              ...data,
              type: 'error',
              content: data.content || 'Unknown error'
            }));
          } 
          // For normal messages, pass through the whole message
          else {
            this.messageHandler(JSON.stringify(data));
          }
        }
      } catch (e) {
        console.error('❌ Failed to parse WebSocket message:', e);
        // If we can't parse it, wrap it as an error
        this.messageHandler?.(JSON.stringify({
          type: 'error',
          content: event.data
        }));
      }
    };

    this.ws.onclose = (event) => {
      console.log(`🛑 WebSocket closed: code=${event.code}, reason=${event.reason}`);
      this.ws = null;
      this.conversationId = null;
    };

    await this.readyPromise;
    console.log('✅ startConversation complete, backend ready');
  }

  async processStream(
    text: string,
    onMessage: (message: string) => void,
    type: 'start' | 'message' = 'message'
  ): Promise<void> {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket connection not open');
    }

    // Wait until conversationId is set
    if (!this.conversationId) {
      await new Promise<void>((resolve) => {
        const interval = setInterval(() => {
          if (this.conversationId) {
            clearInterval(interval);
            resolve();
          }
        }, 10);
      });
    }

    this.messageHandler = onMessage;

    // Ensure type is valid
    if (type !== 'start' && type !== 'message') {
      console.warn(`Invalid type "${type}" passed to processStream(), defaulting to "message"`);
      type = 'message';
    }

    const payload = {
      type,
      content: text,
      conversationId: this.conversationId,
      promptId: 'chat'  // Specify the prompt ID for all messages
    };
    console.log('➡️ Sending message over WebSocket:', payload);

    try {
      this.ws.send(JSON.stringify(payload));
      console.log('✅ Message sent successfully');
    } catch (err) {
      console.error('❌ Failed to send message:', err);
      throw err;
    }
  }

  async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}${API_ENDPOINTS.HEALTH}`);
      return response.ok;
    } catch {
      return false;
    }
  }

  async getStatus() {
    try {
      const response = await fetch(`${this.baseUrl}${API_ENDPOINTS.STATUS}`);
      return await response.json();
    } catch (error) {
      return {
        status: 'offline',
        model: 'unknown',
        uptime: 0,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  closeConnection() {
    if (this.ws) {
      console.log('🛑 Closing WebSocket connection');
      this.ws.close();
      this.ws = null;
      this.conversationId = null;
      this.messageHandler = null;
      this.readyResolver = null;
      this.readyPromise = null;
    } else {
      console.log('WebSocket already closed or not initialized');
    }
  }
}

export const llmService = new LLMService();
