import { apiRequest, withRetry } from '../../utils/apiUtils';
import { Node, NodeWithConnections } from '../../types/graph/index';

export class NodeService {
    async getAllNodes(): Promise<Node[]> {
        return withRetry(() => apiRequest<Node[]>('/nodes'));
    }

    async createNode(type: string, properties: any): Promise<string> {
        const response = await apiRequest<{ id: string }>('/node', {
            method: 'POST',
            body: JSON.stringify({ type, properties }),
        });
        return response.id;
    }

    async getNode(id: string): Promise<NodeWithConnections> {
        if (!id) {
            throw new Error('Node ID is required');
        }
        return withRetry(() => apiRequest<NodeWithConnections>(`/node/${id}`));
    }

    async updateNode(id: string, properties: any): Promise<void> {
        if (!id) {
            throw new Error('Node ID is required');
        }
        await apiRequest<void>(`/node/${id}`, {
            method: 'PUT',
            body: JSON.stringify({ properties }),
        });
    }

    async deleteNode(id: string): Promise<void> {
        if (!id) {
            throw new Error('Node ID is required');
        }
        await apiRequest<void>(`/node/${id}`, {
            method: 'DELETE',
        });
    }

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
    }

    async batchDeleteNodes(ids: string[]): Promise<any> {
        if (!ids || ids.length === 0) {
            throw new Error('At least one node ID is required');
        }
        
        return await apiRequest('/nodes/batch', {
            method: 'DELETE',
            body: JSON.stringify({ ids }),
        });
    }

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
    }
}

export const nodeService = new NodeService();
