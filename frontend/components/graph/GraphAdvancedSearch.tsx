import React, { useState, useCallback } from 'react';
import { useNotification } from '../../hooks/useNotification';

interface FilterDefinition {
    field: string;
    operator: 'equals' | 'contains' | 'startsWith' | 'endsWith' | 'greaterThan' | 'lessThan' | 'between' | 'in' | 'exists';
    value: any;
    type: 'string' | 'number' | 'boolean' | 'date' | 'array';
}

interface TraversalDefinition {
    relationshipType: string;
    direction: 'incoming' | 'outgoing' | 'both';
    nodeType?: string;
    minDepth?: number;
    maxDepth?: number;
    filters?: FilterDefinition[];
}

interface SearchQuery {
    nodeTypes: string[];
    filters: FilterDefinition[];
    traversals: TraversalDefinition[];
    options: {
        limit: number;
        offset: number;
        sortBy?: string;
        sortDirection?: 'asc' | 'desc';
    };
}

interface AdvancedSearchProps {
    availableNodeTypes: Array<{
        name: string;
        properties: Array<{
            name: string;
            type: string;
            description?: string;
        }>;
    }>;
    availableRelationshipTypes: Array<{
        name: string;
        sourceTypes: string[];
        targetTypes: string[];
        properties: Array<{
            name: string;
            type: string;
        }>;
    }>;
    onSearch: (query: SearchQuery) => Promise<any>;
    onSaveQuery?: (name: string, query: SearchQuery) => Promise<void>;
    savedQueries?: Array<{
        name: string;
        query: SearchQuery;
    }>;
    isLoading?: boolean;
}

const GraphAdvancedSearch: React.FC<AdvancedSearchProps> = ({
    availableNodeTypes,
    availableRelationshipTypes,
    onSearch,
    onSaveQuery,
    savedQueries = [],
    isLoading
}) => {
    const { addNotification } = useNotification();
    const [query, setQuery] = useState<SearchQuery>({
        nodeTypes: [],
        filters: [],
        traversals: [],
        options: {
            limit: 50,
            offset: 0
        }
    });

    const handleAddFilter = () => {
        setQuery(prev => ({
            ...prev,
            filters: [
                ...prev.filters,
                {
                    field: '',
                    operator: 'equals',
                    value: '',
                    type: 'string'
                }
            ]
        }));
    };

    const handleFilterChange = (index: number, field: keyof FilterDefinition, value: any) => {
        setQuery(prev => ({
            ...prev,
            filters: prev.filters.map((filter, i) => 
                i === index ? { ...filter, [field]: value } : filter
            )
        }));
    };

    const handleRemoveFilter = (index: number) => {
        setQuery(prev => ({
            ...prev,
            filters: prev.filters.filter((_, i) => i !== index)
        }));
    };

    const handleAddTraversal = () => {
        setQuery(prev => ({
            ...prev,
            traversals: [
                ...prev.traversals,
                {
                    relationshipType: '',
                    direction: 'both',
                    filters: []
                }
            ]
        }));
    };

    const handleTraversalChange = (index: number, field: keyof TraversalDefinition, value: any) => {
        setQuery(prev => ({
            ...prev,
            traversals: prev.traversals.map((traversal, i) => 
                i === index ? { ...traversal, [field]: value } : traversal
            )
        }));
    };

    const handleRemoveTraversal = (index: number) => {
        setQuery(prev => ({
            ...prev,
            traversals: prev.traversals.filter((_, i) => i !== index)
        }));
    };

    const handleExecuteSearch = async () => {
        try {
            const results = await onSearch(query);
            // Results handling would be implemented by the parent component
            addNotification('success', 'Search completed successfully');
        } catch (error) {
            addNotification('error', error instanceof Error ? error.message : 'Search failed');
        }
    };

    const getAvailableProperties = (nodeTypes: string[]) => {
        return availableNodeTypes
            .filter(type => nodeTypes.includes(type.name))
            .flatMap(type => type.properties);
    };

    return (
        <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-6">Advanced Graph Search</h2>

            {/* Node Types Selection */}
            <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                    Node Types
                </label>
                <select
                    multiple
                    value={query.nodeTypes}
                    onChange={(e) => setQuery(prev => ({
                        ...prev,
                        nodeTypes: Array.from(e.target.selectedOptions, option => option.value)
                    }))}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                >
                    {availableNodeTypes.map(type => (
                        <option key={type.name} value={type.name}>
                            {type.name}
                        </option>
                    ))}
                </select>
            </div>

            {/* Filters Section */}
            <div className="mb-6">
                <div className="flex justify-between items-center mb-2">
                    <h3 className="text-lg font-medium">Filters</h3>
                    <button
                        onClick={handleAddFilter}
                        className="text-sm text-blue-600 hover:text-blue-500"
                    >
                        Add Filter
                    </button>
                </div>
                {query.filters.map((filter, index) => (
                    <div key={index} className="flex gap-2 mb-2 items-start bg-gray-50 p-2 rounded">
                        <select
                            value={filter.field}
                            onChange={(e) => {
                                const property = getAvailableProperties(query.nodeTypes)
                                    .find(p => p.name === e.target.value);
                                handleFilterChange(index, 'field', e.target.value);
                                if (property) {
                                    handleFilterChange(index, 'type', property.type);
                                }
                            }}
                            className="flex-1 px-3 py-2 border border-gray-300 rounded-md"
                        >
                            <option value="">Select property</option>
                            {getAvailableProperties(query.nodeTypes).map(prop => (
                                <option key={prop.name} value={prop.name}>
                                    {prop.name} ({prop.type})
                                </option>
                            ))}
                        </select>
                        <select
                            value={filter.operator}
                            onChange={(e) => handleFilterChange(
                                index,
                                'operator',
                                e.target.value as FilterDefinition['operator']
                            )}
                            className="w-40 px-3 py-2 border border-gray-300 rounded-md"
                        >
                            <option value="equals">Equals</option>
                            <option value="contains">Contains</option>
                            <option value="startsWith">Starts with</option>
                            <option value="endsWith">Ends with</option>
                            <option value="greaterThan">Greater than</option>
                            <option value="lessThan">Less than</option>
                            <option value="between">Between</option>
                            <option value="in">In list</option>
                            <option value="exists">Exists</option>
                        </select>
                        {filter.operator !== 'exists' && (
                            <input
                                type={filter.type === 'number' ? 'number' : 'text'}
                                value={filter.value}
                                onChange={(e) => handleFilterChange(index, 'value', e.target.value)}
                                className="flex-1 px-3 py-2 border border-gray-300 rounded-md"
                                placeholder="Value"
                            />
                        )}
                        <button
                            onClick={() => handleRemoveFilter(index)}
                            className="px-2 py-1 text-red-600 hover:bg-red-50 rounded"
                        >
                            Remove
                        </button>
                    </div>
                ))}
            </div>

            {/* Traversals Section */}
            <div className="mb-6">
                <div className="flex justify-between items-center mb-2">
                    <h3 className="text-lg font-medium">Graph Traversals</h3>
                    <button
                        onClick={handleAddTraversal}
                        className="text-sm text-blue-600 hover:text-blue-500"
                    >
                        Add Traversal
                    </button>
                </div>
                {query.traversals.map((traversal, index) => (
                    <div key={index} className="mb-4 p-4 border rounded-lg bg-gray-50">
                        <div className="grid grid-cols-2 gap-4 mb-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Relationship Type
                                </label>
                                <select
                                    value={traversal.relationshipType}
                                    onChange={(e) => handleTraversalChange(index, 'relationshipType', e.target.value)}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                                >
                                    <option value="">Select type</option>
                                    {availableRelationshipTypes.map(type => (
                                        <option key={type.name} value={type.name}>
                                            {type.name}
                                        </option>
                                    ))}
                                </select>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Direction
                                </label>
                                <select
                                    value={traversal.direction}
                                    onChange={(e) => handleTraversalChange(
                                        index,
                                        'direction',
                                        e.target.value as TraversalDefinition['direction']
                                    )}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                                >
                                    <option value="both">Both directions</option>
                                    <option value="incoming">Incoming</option>
                                    <option value="outgoing">Outgoing</option>
                                </select>
                            </div>
                        </div>
                        <div className="grid grid-cols-2 gap-4 mb-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Min Depth
                                </label>
                                <input
                                    type="number"
                                    value={traversal.minDepth || ''}
                                    onChange={(e) => handleTraversalChange(
                                        index,
                                        'minDepth',
                                        e.target.value ? parseInt(e.target.value) : undefined
                                    )}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                                    min="1"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Max Depth
                                </label>
                                <input
                                    type="number"
                                    value={traversal.maxDepth || ''}
                                    onChange={(e) => handleTraversalChange(
                                        index,
                                        'maxDepth',
                                        e.target.value ? parseInt(e.target.value) : undefined
                                    )}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                                    min="1"
                                />
                            </div>
                        </div>
                        <button
                            onClick={() => handleRemoveTraversal(index)}
                            className="text-sm text-red-600 hover:text-red-500"
                        >
                            Remove Traversal
                        </button>
                    </div>
                ))}
            </div>

            {/* Options Section */}
            <div className="mb-6 grid grid-cols-2 gap-4">
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Results Limit
                    </label>
                    <input
                        type="number"
                        value={query.options.limit}
                        onChange={(e) => setQuery(prev => ({
                            ...prev,
                            options: {
                                ...prev.options,
                                limit: parseInt(e.target.value) || 50
                            }
                        }))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        min="1"
                        max="1000"
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Sort By
                    </label>
                    <select
                        value={query.options.sortBy || ''}
                        onChange={(e) => setQuery(prev => ({
                            ...prev,
                            options: {
                                ...prev.options,
                                sortBy: e.target.value || undefined
                            }
                        }))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    >
                        <option value="">Default order</option>
                        {getAvailableProperties(query.nodeTypes).map(prop => (
                            <option key={prop.name} value={prop.name}>
                                {prop.name}
                            </option>
                        ))}
                    </select>
                </div>
            </div>

            {/* Action Buttons */}
            <div className="flex justify-end space-x-4">
                <button
                    onClick={() => setQuery({
                        nodeTypes: [],
                        filters: [],
                        traversals: [],
                        options: {
                            limit: 50,
                            offset: 0
                        }
                    })}
                    className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
                >
                    Clear
                </button>
                <button
                    onClick={handleExecuteSearch}
                    disabled={isLoading}
                    className={`px-6 py-2 rounded-md text-white ${
                        isLoading ? 'bg-gray-400' : 'bg-blue-600 hover:bg-blue-700'
                    }`}
                >
                    {isLoading ? 'Searching...' : 'Execute Search'}
                </button>
            </div>
        </div>
    );
};

export default GraphAdvancedSearch;
