import { Node } from '../graph/index';

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

export interface EntityConnection {
    source: Node;
    target: Node;
    relationship: {
        type: string;
        properties: Record<string, any>;
    };
    depth: number;
}

export interface EntityConnectionsResponse {
    nodes: Node[];
    connections: EntityConnection[];
    centralNode: Node;
}

export interface TimelineEvent {
    id: string;
    date: string;
    type: string;
    description: string;
    entities: Array<{
        id: string;
        name: string;
        type: string;
    }>;
    sources: string[];
}

export interface TimelineResponse {
    events: TimelineEvent[];
    total: number;
}

export interface NetworkStats {
    totalNodes: number;
    totalRelationships: number;
    nodeTypes: Array<{
        type: string;
        count: number;
    }>;
    relationshipTypes: Array<{
        type: string;
        count: number;
    }>;
    riskDistribution: {
        high: number;
        medium: number;
        low: number;
    };
    topConnectors: Array<{
        id: string;
        name: string;
        type: string;
        connections: number;
    }>;
}
