import React from 'react';

interface NetworkStatistics {
    totalNodes: number;
    totalRelationships: number;
    nodeTypes: {
        type: string;
        count: number;
        percentage: number;
    }[];
    relationshipTypes: {
        type: string;
        count: number;
        percentage: number;
    }[];
    averageConnections: number;
    density: number;
    mostConnectedNodes: {
        id: string;
        name: string;
        type: string;
        connections: number;
    }[];
    graphMetrics: {
        diameter: number;
        averagePathLength: number;
        clusteringCoefficient: number;
    };
}

interface NetworkStatsProps {
    stats: NetworkStatistics;
    isLoading?: boolean;
    error?: string;
}

const NetworkStats: React.FC<NetworkStatsProps> = ({ stats, isLoading, error }) => {
    if (isLoading) {
        return (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 animate-pulse">
                {[1, 2, 3, 4, 5, 6].map((i) => (
                    <div key={i} className="bg-white p-6 rounded-lg shadow">
                        <div className="h-4 bg-gray-200 rounded w-1/2 mb-4"></div>
                        <div className="h-8 bg-gray-200 rounded"></div>
                    </div>
                ))}
            </div>
        );
    }

    if (error) {
        return (
            <div className="bg-red-50 border border-red-200 rounded-lg p-6">
                <p className="text-red-600">{error}</p>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Overview Cards */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                <StatCard
                    title="Total Nodes"
                    value={stats.totalNodes}
                    description="Total entities in the network"
                />
                <StatCard
                    title="Total Relationships"
                    value={stats.totalRelationships}
                    description="Total connections between entities"
                />
                <StatCard
                    title="Network Density"
                    value={`${(stats.density * 100).toFixed(2)}%`}
                    description="Percentage of possible connections that exist"
                />
            </div>

            {/* Distribution Charts */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="bg-white rounded-lg shadow p-6">
                    <h3 className="text-lg font-medium mb-4">Node Type Distribution</h3>
                    <div className="space-y-4">
                        {stats.nodeTypes.map((type) => (
                            <DistributionBar
                                key={type.type}
                                label={type.type}
                                count={type.count}
                                percentage={type.percentage}
                            />
                        ))}
                    </div>
                </div>

                <div className="bg-white rounded-lg shadow p-6">
                    <h3 className="text-lg font-medium mb-4">Relationship Type Distribution</h3>
                    <div className="space-y-4">
                        {stats.relationshipTypes.map((type) => (
                            <DistributionBar
                                key={type.type}
                                label={type.type}
                                count={type.count}
                                percentage={type.percentage}
                            />
                        ))}
                    </div>
                </div>
            </div>

            {/* Most Connected Nodes */}
            <div className="bg-white rounded-lg shadow p-6">
                <h3 className="text-lg font-medium mb-4">Most Connected Entities</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {stats.mostConnectedNodes.map((node) => (
                        <div 
                            key={node.id}
                            className="p-4 border rounded-lg bg-gray-50"
                        >
                            <div className="text-sm text-gray-500">{node.type}</div>
                            <div className="font-medium">{node.name}</div>
                            <div className="text-sm text-blue-600">{node.connections} connections</div>
                        </div>
                    ))}
                </div>
            </div>

            {/* Graph Metrics */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <StatCard
                    title="Network Diameter"
                    value={stats.graphMetrics.diameter}
                    description="Maximum distance between any two nodes"
                />
                <StatCard
                    title="Average Path Length"
                    value={stats.graphMetrics.averagePathLength.toFixed(2)}
                    description="Average number of steps between nodes"
                />
                <StatCard
                    title="Clustering Coefficient"
                    value={stats.graphMetrics.clusteringCoefficient.toFixed(3)}
                    description="Degree of node clustering"
                />
            </div>
        </div>
    );
};

interface StatCardProps {
    title: string;
    value: number | string;
    description: string;
}

const StatCard: React.FC<StatCardProps> = ({ title, value, description }) => (
    <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-sm font-medium text-gray-500">{title}</h3>
        <div className="mt-2 flex items-baseline">
            <div className="text-2xl font-semibold text-gray-900">{value}</div>
        </div>
        <div className="mt-1 text-sm text-gray-500">{description}</div>
    </div>
);

interface DistributionBarProps {
    label: string;
    count: number;
    percentage: number;
}

const DistributionBar: React.FC<DistributionBarProps> = ({ label, count, percentage }) => (
    <div>
        <div className="flex justify-between text-sm mb-1">
            <span className="font-medium">{label}</span>
            <span className="text-gray-500">{count}</span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-2">
            <div
                className="bg-blue-600 rounded-full h-2"
                style={{ width: `${percentage}%` }}
            />
        </div>
    </div>
);

export default NetworkStats;
