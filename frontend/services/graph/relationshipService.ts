import { apiRequest, withRetry } from '../../utils/apiUtils';

export class RelationshipService {
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
    }

    async updateRelationship(id: string, properties: any): Promise<void> {
        if (!id) {
            throw new Error('Relationship ID is required');
        }
        await apiRequest<void>(`/relationship/${id}`, {
            method: 'PUT',
            body: JSON.stringify({ properties }),
        });
    }

    async deleteRelationship(id: string): Promise<void> {
        if (!id) {
            throw new Error('Relationship ID is required');
        }
        await apiRequest<void>(`/relationship/${id}`, {
            method: 'DELETE',
        });
    }

    async getRelationship(id: string): Promise<any> {
        if (!id) {
            throw new Error('Relationship ID is required');
        }
        return withRetry(() => apiRequest(`/relationship/${id}`));
    }
}

export const relationshipService = new RelationshipService();
