import { useEffect, useState } from 'react';
import { api, NodeWithConnections } from '../../utils/api';

interface NodeDetailsProps {
    nodeId?: string;
}

const NodeDetails: React.FC<NodeDetailsProps> = ({ nodeId }) => {
    const [node, setNode] = useState<NodeWithConnections | null>(null);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        if (!nodeId) {
            setNode(null);
            return;
        }

        setLoading(true);
        api.getNode(nodeId)
            .then(setNode)
            .catch(console.error)
            .finally(() => setLoading(false));
    }, [nodeId]);

    if (!nodeId) return null;
    if (loading) return <div>Loading...</div>;
    if (!node) return <div>Node not found</div>;

    return (
        <div className="space-y-6">
            <div className="bg-white shadow rounded-lg p-6">
                <div className="flex items-center space-x-3">
                    <div className={`w-3 h-3 rounded-full ${
                        node.type === 'Person' ? 'bg-green-500' :
                        node.type === 'Organization' ? 'bg-blue-500' :
                        'bg-yellow-500'
                    }`} />
                    <h3 className="text-lg font-medium">
                        {node.type === 'Event' ? node.properties.title : node.properties.name}
                    </h3>
                </div>
                
                <div className="mt-4 grid grid-cols-1 gap-4">
                    {/* Common Properties */}
                    <div>
                        <div className="text-sm font-medium text-gray-500">Type</div>
                        <div className="mt-1">{node.type}</div>
                    </div>

                    {/* Person & Organization Properties */}
                    {(node.type === 'Person' || node.type === 'Organization') && node.properties.aliases && (
                        <div>
                            <div className="text-sm font-medium text-gray-500">Aliases</div>
                            <div className="flex flex-wrap gap-2 mt-1">
                                {node.properties.aliases.map((alias: string) => (
                                    <span key={alias} className="px-2 py-1 bg-gray-100 rounded-full text-sm">
                                        {alias}
                                    </span>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Organization-specific Properties */}
                    {node.type === 'Organization' && node.properties.type && (
                        <div>
                            <div className="text-sm font-medium text-gray-500">Organization Type</div>
                            <div className="mt-1">{node.properties.type}</div>
                        </div>
                    )}

                    {/* Event-specific Properties */}
                    {node.type === 'Event' && (
                        <>
                            {node.properties.description && (
                                <div>
                                    <div className="text-sm font-medium text-gray-500">Description</div>
                                    <div className="mt-1">{node.properties.description}</div>
                                </div>
                            )}
                            {node.properties.date && (
                                <div>
                                    <div className="text-sm font-medium text-gray-500">Date</div>
                                    <div className="mt-1">{new Date(node.properties.date).toLocaleDateString()}</div>
                                </div>
                            )}
                            {node.properties.sources && node.properties.sources.length > 0 && (
                                <div>
                                    <div className="text-sm font-medium text-gray-500">Sources</div>
                                    <div className="mt-1 space-y-1">
                                        {node.properties.sources.map((url: string) => (
                                            <a
                                                key={url}
                                                href={url}
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                className="text-blue-600 hover:underline block text-sm"
                                            >
                                                {url}
                                            </a>
                                        ))}
                                    </div>
                                </div>
                            )}
                        </>
                    )}

                    {/* Notes (Common) */}
                    {node.properties.notes && (
                        <div>
                            <div className="text-sm font-medium text-gray-500">Notes</div>
                            <div className="mt-1 text-sm whitespace-pre-wrap">{node.properties.notes}</div>
                        </div>
                    )}
                </div>
            </div>

            {/* Connections */}
            {node.connections.length > 0 && (
                <div className="bg-white shadow rounded-lg p-6">
                    <h4 className="text-lg font-medium mb-4">Connections</h4>
                    <div className="space-y-4">
                        {node.connections.map((conn) => (
                            <div 
                                key={conn.id} 
                                className="p-4 bg-gray-50 rounded-lg border border-gray-100 hover:border-gray-200 transition-colors"
                            >
                                <div className="flex items-center gap-3">
                                    <div className={`w-2 h-2 rounded-full ${
                                        conn.type === 'Person' ? 'bg-green-500' :
                                        conn.type === 'Organization' ? 'bg-blue-500' :
                                        'bg-yellow-500'
                                    }`} />
                                    <span className="text-sm font-medium">
                                        {(conn.properties && (conn.properties.name || conn.properties.title)) || `Unnamed ${conn.type}`}
                                    </span>
                                    <span className={`text-xs px-2 py-1 rounded ${
                                        conn.relationship.direction === 'out' ? 'bg-blue-100 text-blue-800' : 'bg-purple-100 text-purple-800'
                                    }`}>
                                        {conn.relationship.type}
                                    </span>
                                </div>
                                <div className="mt-2 text-sm text-gray-600">
                                    {conn.relationship.direction === 'out' ? 
                                        `${node.properties.name || node.properties.title || `This ${node.type}`} ${conn.relationship.type} this ${conn.type}` :
                                        `This ${conn.type} ${conn.relationship.type} ${node.properties.name || node.properties.title || `this ${node.type}`}`
                                    }
                                </div>
                                {(conn.relationship.properties?.date || conn.relationship.properties?.description) && (
                                    <div className="mt-2 text-xs text-gray-500 space-y-1">
                                        {conn.relationship.properties.date && (
                                            <div>Date: {new Date(conn.relationship.properties.date).toLocaleDateString()}</div>
                                        )}
                                        {conn.relationship.properties.description && (
                                            <div>{conn.relationship.properties.description}</div>
                                        )}
                                    </div>
                                )}
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

export default NodeDetails;
