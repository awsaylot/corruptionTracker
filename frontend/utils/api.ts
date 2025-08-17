// utils/api.ts - Enhanced with analytics and extraction features
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const API_PATH = `${API_BASE}/api`;

export interface Node {
    id: string;
    type: string;
    labels: string[];
    properties: {
        name?: string;
        title?: string;
        aliases?: string[];
        notes?: string;
        corruption_score?: number;
        [key: string]: any;
    };
}

export interface Connection {
    id: string;
    type: string;
    properties: any;
    relationship: {
        type: string;
        properties: any;
        direction: 'in' | 'out';
    };
}

export interface NodeWithConnections extends Node {
    connections: Connection[];
}

export interface ExtractionResult {
    success: boolean;
    source_url: string;
    text_length: number;
    processing_time: number;
    summary: {
        total_entities: number;
        total_events: number;
        unique_actors: number;
        unique_targets: number;
        date_mentions: number;
        location_mentions: number;
        financial_amounts: number;
        processing_time_seconds: number;
    };
    relevance_score: {
        total_score: number;
        corruption_indicators: number;
        legal_indicators: number;
        financial_indicators: number;
        event_indicators: number;
    };
    entities: {
        PERSON: string[];
        ORG: string[];
        GPE: string[];
        LOC: string[];
        DATE: string[];
        MONEY: string[];
        NORP: string[];
        LAW: string[];
        EVENT: string[];
        CARDINAL: string[];
        PERCENT: string[];
    };
    events: Array<{
        actor: string;
        action: string;
        target: string;
        dates: string[];
        locations: string[];
        organizations: string[];
        negated: boolean;
        modality?: string;
        certainty: string;
        context_keywords: Record<string, string[]>;
        source_sentence: string;
        sentence_start: number;
        sentence_end: number;
    }>;
    urls_in_text: string[];
    normalized_dates: string[];
    financial_amounts: Array<{
        text: string;
        value?: number;
        currency?: string;
    }>;
    enriched_entities: Record<string, {
        entity: string;
        mention_count: number;
        associated_events: number;
        context_types: string[];
        first_mention: number;
    }>;
}

export interface CorruptionScore {
    node_id: string;
    name: string;
    node_type: string;
    corruption_score: number;
    risk_level: string;
    risk_color: string;
    score_breakdown: {
        legal: number;
        financial: number;
        corruption: number;
        control: number;
        offshore: number;
        frequency: number;
    };
    mention_count: number;
    updated_at: string;
}

export interface ApiError extends Error {
    status?: number;
    code?: string;
}

// Enhanced fetch wrapper with better error handling
const apiRequest = async <T>(
    endpoint: string, 
    options: RequestInit = {}
): Promise<T> => {
    const url = `${API_PATH}${endpoint}`;
    const controller = new AbortController();
    
    // Set timeout for requests
    const timeoutId = setTimeout(() => controller.abort(), 30000);
    
    try {
        const response = await fetch(url, {
            ...options,
            signal: controller.signal,
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
        });

        clearTimeout(timeoutId);

        if (!response.ok) {
            let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
            
            try {
                const errorData = await response.json();
                errorMessage = errorData.message || errorData.error || errorMessage;
            } catch {
                // If we can't parse JSON, use the default message
            }

            const error = new Error(errorMessage) as ApiError;
            error.status = response.status;
            throw error;
        }

        // Handle empty responses
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            return await response.json();
        }
        
        return await response.text() as unknown as T;
        
    } catch (error) {
        clearTimeout(timeoutId);
        
        if (error instanceof Error) {
            if (error.name === 'AbortError') {
                throw new Error('Request timed out. Please try again.');
            }
            
            // Network errors
            if (!navigator.onLine) {
                throw new Error('No internet connection. Please check your network.');
            }
            
            // Connection refused or server down
            if (error.message.includes('fetch')) {
                throw new Error('Unable to connect to database server. Please ensure the backend is running.');
            }
        }
        
        throw error;
    }
};

// Retry logic for failed requests
const withRetry = async <T>(
    operation: () => Promise<T>,
    maxRetries: number = 3,
    delay: number = 1000
): Promise<T> => {
    let lastError: Error;
    
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
            return await operation();
        } catch (error) {
            lastError = error instanceof Error ? error : new Error('Unknown error');
            
            // Don't retry on client errors (4xx)
            if ('status' in lastError && typeof lastError.status === 'number' && lastError.status >= 400 && lastError.status < 500) {
                throw lastError;
            }
            
            if (attempt === maxRetries) {
                throw lastError;
            }
            
            // Wait before retrying
            await new Promise(resolve => setTimeout(resolve, delay * attempt));
        }
    }
    
    throw lastError!;
};

export const api = {
    // Node operations
    async getAllNodes(): Promise<Node[]> {
        return withRetry(() => apiRequest<Node[]>('/nodes'));
    },

    async createNode(type: string, properties: any): Promise<string> {
        const response = await apiRequest<{ id: string }>('/node', {
            method: 'POST',
            body: JSON.stringify({ type, properties }),
        });
        return response.id;
    },

    async getNode(id: string): Promise<NodeWithConnections> {
        if (!id) {
            throw new Error('Node ID is required');
        }
        return withRetry(() => apiRequest<NodeWithConnections>(`/node/${id}`));
    },

    async updateNode(id: string, properties: any): Promise<void> {
        if (!id) {
            throw new Error('Node ID is required');
        }
        await apiRequest<void>(`/node/${id}`, {
            method: 'PUT',
            body: JSON.stringify({ properties }),
        });
    },

    async deleteNode(id: string): Promise<void> {
        if (!id) {
            throw new Error('Node ID is required');
        }
        await apiRequest<void>(`/node/${id}`, {
            method: 'DELETE',
        });
    },

    // Search operations
    async searchNodes(
        query: string,
        options?: {
            type?: string;
            mode?: 'basic' | 'advanced';
            includePath?: boolean;
        }
    ): Promise<{ nodes: Node[]; paths?: Array<{ nodes: Node[]; relationships: any[] }> } | Node[]> {
        const params = new URLSearchParams();
        
        if (query) params.append('q', query);
        if (options?.type) params.append('type', options.type);
        if (options?.mode) params.append('mode', options.mode);
        if (options?.includePath) params.append('includePath', 'true');
        
        const endpoint = `/search${params.toString() ? `?${params.toString()}` : ''}`;
        return withRetry(() => apiRequest(endpoint));
    },

    // Relationship operations
    async createRelationship(fromId: string, toId: string, type: string, properties: any = {}): Promise<string> {
        if (!fromId || !toId) {
            throw new Error('Both source and target node IDs are required');
        }
        if (fromId === toId) {
            throw new Error('Cannot create relationship between the same node');
        }
        if (!type) {
            throw new Error('Relationship type is required');
        }

        const response = await apiRequest<{ id: string }>('/relationship', {
            method: 'POST',
            body: JSON.stringify({ fromId, toId, type, properties }),
        });
        return response.id;
    },

    async updateRelationship(id: string, properties: any): Promise<void> {
        if (!id) {
            throw new Error('Relationship ID is required');
        }
        await apiRequest<void>(`/relationship/${id}`, {
            method: 'PUT',
            body: JSON.stringify({ properties }),
        });
    },

    async deleteRelationship(id: string): Promise<void> {
        if (!id) {
            throw new Error('Relationship ID is required');
        }
        await apiRequest<void>(`/relationship/${id}`, {
            method: 'DELETE',
        });
    },

    async getRelationship(id: string): Promise<any> {
        if (!id) {
            throw new Error('Relationship ID is required');
        }
        return withRetry(() => apiRequest(`/relationship/${id}`));
    },

    // Network operations
    async getNetwork(): Promise<NodeWithConnections[]> {
        return withRetry(() => apiRequest<NodeWithConnections[]>('/network'));
    },

    async getSubgraph(nodeId: string, depth: number = 1): Promise<any> {
        if (!nodeId) {
            throw new Error('Node ID is required');
        }
        if (depth < 1 || depth > 5) {
            throw new Error('Depth must be between 1 and 5');
        }
        return withRetry(() => apiRequest(`/subgraph/${nodeId}?depth=${depth}`));
    },

    async getShortestPath(fromId: string, toId: string): Promise<any> {
        if (!fromId || !toId) {
            throw new Error('Both source and target node IDs are required');
        }
        return withRetry(() => apiRequest(`/path?from=${fromId}&to=${toId}`));
    },

    // Schema operations
    async getNodeTypes(): Promise<string[]> {
        const response = await withRetry(() => apiRequest<{ types: string[] }>('/node-types'));
        return response.types;
    },

    async getNodeTypeSchema(type: string): Promise<string[]> {
        if (!type) {
            throw new Error('Node type is required');
        }
        const response = await withRetry(() => apiRequest<{ propertyKeys: string[] }>(`/node-types/${type}/schema`));
        return response.propertyKeys;
    },

    // Batch operations
    async batchCreateNodes(nodes: { type: string, properties: any }[]): Promise<any[]> {
        if (!nodes || nodes.length === 0) {
            throw new Error('At least one node is required');
        }
        
        nodes.forEach((node, index) => {
            if (!node.type) {
                throw new Error(`Node at index ${index} is missing type`);
            }
            if (!node.properties) {
                throw new Error(`Node at index ${index} is missing properties`);
            }
        });

        const response = await apiRequest<{ nodes: any[] }>('/nodes/batch', {
            method: 'POST',
            body: JSON.stringify(nodes),
        });
        return response.nodes;
    },

    async batchDeleteNodes(ids: string[]): Promise<any> {
        if (!ids || ids.length === 0) {
            throw new Error('At least one node ID is required');
        }
        
        return await apiRequest('/nodes/batch', {
            method: 'DELETE',
            body: JSON.stringify({ ids }),
        });
    },

    // Enhanced Extraction operations
    async extractFromURL(url: string, options?: { createNodes?: boolean, createEvents?: boolean }): Promise<ExtractionResult> {
        if (!url) {
            throw new Error("URL is required");
        }
        
        const response = await apiRequest<ExtractionResult>('/extract', {
            method: 'POST',
            body: JSON.stringify({ 
                url, 
                create_nodes: options?.createNodes,
                create_events: options?.createEvents 
            }),
        });
        return response;
    },

    async extractFromText(text: string, options?: { createNodes?: boolean, createEvents?: boolean }): Promise<ExtractionResult> {
        if (!text) {
            throw new Error("Text is required");
        }
        
        const response = await apiRequest<ExtractionResult>('/extract', {
            method: 'POST',
            body: JSON.stringify({ 
                text, 
                create_nodes: options?.createNodes,
                create_events: options?.createEvents 
            }),
        });
        return response;
    },

    async createNodesFromExtraction(data: {
        selected_people: string[];
        selected_organizations: string[];
        selected_events: number[];
        source_url?: string;
        source_title?: string;
    }): Promise<any> {
        return await apiRequest('/extract/create-nodes', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    },

    async getExtractionHistory(limit: number = 10): Promise<any> {
        return await apiRequest(`/extract/history?limit=${limit}`);
    },

    // Analytics operations
    async getCorruptionScore(nodeId: string): Promise<CorruptionScore> {
        if (!nodeId) {
            throw new Error('Node ID is required');
        }
        return withRetry(() => apiRequest<CorruptionScore>(`/analytics/corruption-score/${nodeId}`));
    },

    async getEntityConnections(nodeId: string, depth: number = 2): Promise<any> {
        if (!nodeId) {
            throw new Error('Node ID is required');
        }
        return withRetry(() => apiRequest(`/analytics/entity-connections/${nodeId}?depth=${depth}`));
    },

    async getTimeline(options?: {
        entity?: string;
        from?: string;
        to?: string;
        limit?: number;
    }): Promise<any> {
        const params = new URLSearchParams();
        if (options?.entity) params.append('entity', options.entity);
        if (options?.from) params.append('from', options.from);
        if (options?.to) params.append('to', options.to);
        if (options?.limit) params.append('limit', options.limit.toString());
        
        const endpoint = `/analytics/timeline${params.toString() ? `?${params.toString()}` : ''}`;
        return withRetry(() => apiRequest(endpoint));
    },

    async getNetworkStats(): Promise<any> {
        return withRetry(() => apiRequest('/analytics/network-stats'));
    },

    // Health check
    async healthCheck(): Promise<{ status: string; timestamp: string }> {
        return apiRequest('/health');
    },

    // Utility function to test connection
    async testConnection(): Promise<boolean> {
        try {
            await this.healthCheck();
            return true;
        } catch {
            return false;
        }
    },
};

export const extractFromURL = api.extractFromURL;
export const extractFromText = api.extractFromText;