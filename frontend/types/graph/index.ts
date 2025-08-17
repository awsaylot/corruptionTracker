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
