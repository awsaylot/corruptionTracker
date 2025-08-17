import { apiRequest, withRetry } from '../../utils/apiUtils';
import { 
    CorruptionScore, 
    EntityConnectionsResponse, 
    TimelineResponse, 
    NetworkStats 
} from '../../types/analytics/index';

export class AnalyticsService {
    async getCorruptionScore(nodeId: string): Promise<CorruptionScore> {
        if (!nodeId) {
            throw new Error('Node ID is required');
        }
        return withRetry(() => apiRequest<CorruptionScore>(`/analytics/corruption-score/${nodeId}`));
    }

    async getEntityConnections(nodeId: string, depth: number = 2): Promise<EntityConnectionsResponse> {
        if (!nodeId) {
            throw new Error('Node ID is required');
        }
        return withRetry(() => apiRequest<EntityConnectionsResponse>(`/analytics/entity-connections/${nodeId}?depth=${depth}`));
    }

    async getTimeline(options?: {
        entity?: string;
        from?: string;
        to?: string;
        limit?: number;
    }): Promise<TimelineResponse> {
        const params = new URLSearchParams();
        if (options?.entity) params.append('entity', options.entity);
        if (options?.from) params.append('from', options.from);
        if (options?.to) params.append('to', options.to);
        if (options?.limit) params.append('limit', options.limit.toString());
        
        const endpoint = `/analytics/timeline${params.toString() ? `?${params.toString()}` : ''}`;
        return withRetry(() => apiRequest<TimelineResponse>(endpoint));
    }

    async getNetworkStats(): Promise<NetworkStats> {
        return withRetry(() => apiRequest<NetworkStats>('/analytics/network-stats'));
    }
}

export const analyticsService = new AnalyticsService();
