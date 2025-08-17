import React, { useState } from 'react';
import { useNotification } from '../../hooks/useNotification';

interface GraphOperation {
    operation: 'merge' | 'shortest_path' | 'subgraph' | 'neighbors';
    parameters: Record<string, any>;
    description: string;
}

interface GraphOperationsProps {
    onExecute: (operation: GraphOperation) => Promise<void>;
    isLoading?: boolean;
}

const GraphOperations: React.FC<GraphOperationsProps> = ({ onExecute, isLoading }) => {
    const { addNotification } = useNotification();
    const [selectedOperation, setSelectedOperation] = useState<GraphOperation['operation']>('merge');
    const [parameters, setParameters] = useState<Record<string, any>>({});

    const operationConfigs = {
        merge: {
            description: 'Merge multiple subgraphs into one',
            fields: [
                {
                    name: 'graphIds',
                    label: 'Graph IDs',
                    type: 'array',
                    placeholder: 'Enter graph IDs to merge'
                }
            ]
        },
        shortest_path: {
            description: 'Find shortest path between two nodes',
            fields: [
                {
                    name: 'sourceId',
                    label: 'Source Node ID',
                    type: 'string',
                    placeholder: 'Enter source node ID'
                },
                {
                    name: 'targetId',
                    label: 'Target Node ID',
                    type: 'string',
                    placeholder: 'Enter target node ID'
                },
                {
                    name: 'relationshipTypes',
                    label: 'Relationship Types',
                    type: 'array',
                    placeholder: 'Enter relationship types (optional)'
                }
            ]
        },
        subgraph: {
            description: 'Extract a subgraph based on criteria',
            fields: [
                {
                    name: 'nodeIds',
                    label: 'Node IDs',
                    type: 'array',
                    placeholder: 'Enter node IDs'
                },
                {
                    name: 'depth',
                    label: 'Depth',
                    type: 'number',
                    placeholder: 'Enter traversal depth'
                },
                {
                    name: 'relationshipTypes',
                    label: 'Relationship Types',
                    type: 'array',
                    placeholder: 'Enter relationship types (optional)'
                }
            ]
        },
        neighbors: {
            description: 'Get neighboring nodes',
            fields: [
                {
                    name: 'nodeId',
                    label: 'Node ID',
                    type: 'string',
                    placeholder: 'Enter node ID'
                },
                {
                    name: 'direction',
                    label: 'Direction',
                    type: 'select',
                    options: ['incoming', 'outgoing', 'both'],
                    placeholder: 'Select direction'
                },
                {
                    name: 'relationshipTypes',
                    label: 'Relationship Types',
                    type: 'array',
                    placeholder: 'Enter relationship types (optional)'
                }
            ]
        }
    };

    const handleParameterChange = (field: string, value: any) => {
        setParameters(prev => ({
            ...prev,
            [field]: value
        }));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await onExecute({
                operation: selectedOperation,
                parameters,
                description: operationConfigs[selectedOperation].description
            });
            addNotification('success', 'Operation executed successfully');
        } catch (error) {
            addNotification('error', error instanceof Error ? error.message : 'Failed to execute operation');
        }
    };

    const renderField = (field: any) => {
        switch (field.type) {
            case 'array':
                return (
                    <div key={field.name} className="mb-4">
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                            {field.label}
                        </label>
                        <input
                            type="text"
                            placeholder={field.placeholder}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                            onChange={(e) => handleParameterChange(
                                field.name,
                                e.target.value.split(',').map(s => s.trim())
                            )}
                        />
                    </div>
                );
            case 'select':
                return (
                    <div key={field.name} className="mb-4">
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                            {field.label}
                        </label>
                        <select
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                            onChange={(e) => handleParameterChange(field.name, e.target.value)}
                            value={parameters[field.name] || ''}
                        >
                            <option value="">{field.placeholder}</option>
                            {field.options.map((option: string) => (
                                <option key={option} value={option}>
                                    {option.charAt(0).toUpperCase() + option.slice(1)}
                                </option>
                            ))}
                        </select>
                    </div>
                );
            default:
                return (
                    <div key={field.name} className="mb-4">
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                            {field.label}
                        </label>
                        <input
                            type={field.type}
                            placeholder={field.placeholder}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                            onChange={(e) => handleParameterChange(
                                field.name,
                                field.type === 'number' ? Number(e.target.value) : e.target.value
                            )}
                        />
                    </div>
                );
        }
    };

    return (
        <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-6">Graph Operations</h2>
            
            <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                    Operation Type
                </label>
                <select
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    value={selectedOperation}
                    onChange={(e) => {
                        setSelectedOperation(e.target.value as GraphOperation['operation']);
                        setParameters({});
                    }}
                >
                    {Object.entries(operationConfigs).map(([key, config]) => (
                        <option key={key} value={key}>
                            {config.description}
                        </option>
                    ))}
                </select>
            </div>

            <form onSubmit={handleSubmit}>
                {operationConfigs[selectedOperation].fields.map(renderField)}
                
                <button
                    type="submit"
                    disabled={isLoading}
                    className={`w-full py-2 px-4 rounded-md text-white font-medium 
                        ${isLoading 
                            ? 'bg-gray-400 cursor-not-allowed'
                            : 'bg-blue-600 hover:bg-blue-700'
                        }`}
                >
                    {isLoading ? 'Processing...' : 'Execute Operation'}
                </button>
            </form>
        </div>
    );
};

export default GraphOperations;
