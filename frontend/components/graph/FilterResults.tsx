import React, { useState, useMemo } from 'react';
import { FilterTags, FilterSummaryPanel } from './FilterVisualization';

interface FilterResult {
    id: string;
    type: string;
    properties: Record<string, any>;
    relationships?: Array<{
        id: string;
        type: string;
        direction: 'incoming' | 'outgoing';
        nodeId: string;
    }>;
    score?: number;
    matchedFilters?: string[];
}

interface FilterResultsProps {
    results: FilterResult[];
    activeFilters: Array<{
        id: string;
        filter: {
            field: string;
            operator: string;
            value: any;
            type: 'date-range' | 'numeric-range' | 'enum-multi' | 'string-pattern' | 'boolean' | 'array';
        };
        color?: 'blue' | 'green' | 'yellow' | 'red' | 'purple' | 'gray';
    }>;
    onRemoveFilter: (id: string) => void;
    onClearFilters: () => void;
    onEditFilter?: (id: string) => void;
    onSelectResult?: (result: FilterResult) => void;
    isLoading?: boolean;
    error?: string;
    className?: string;
}

const FilterResults: React.FC<FilterResultsProps> = ({
    results,
    activeFilters,
    onRemoveFilter,
    onClearFilters,
    onEditFilter,
    onSelectResult,
    isLoading,
    error,
    className = ''
}) => {
    const [sortField, setSortField] = useState<string>('score');
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');
    const [viewMode, setViewMode] = useState<'list' | 'grid'>('list');

    const sortedResults = useMemo(() => {
        if (!results.length) return [];

        return [...results].sort((a, b) => {
            let aValue = a[sortField as keyof FilterResult];
            let bValue = b[sortField as keyof FilterResult];

            if (typeof aValue === 'object') {
                aValue = JSON.stringify(aValue);
                bValue = JSON.stringify(bValue);
            }

            if (aValue === undefined) aValue = null;
            if (bValue === undefined) bValue = null;

            if (aValue === null && bValue === null) return 0;
            if (aValue === null) return sortDirection === 'asc' ? -1 : 1;
            if (bValue === null) return sortDirection === 'asc' ? 1 : -1;

            const comparison = aValue < bValue ? -1 : aValue > bValue ? 1 : 0;
            return sortDirection === 'asc' ? comparison : -comparison;
        });
    }, [results, sortField, sortDirection]);

    if (error) {
        return (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700">
                {error}
            </div>
        );
    }

    if (isLoading) {
        return (
            <div className="space-y-4">
                {[1, 2, 3].map((i) => (
                    <div key={i} className="animate-pulse">
                        <div className="h-24 bg-gray-200 rounded-lg"></div>
                    </div>
                ))}
            </div>
        );
    }

    if (results.length === 0) {
        return (
            <div className="text-center py-8">
                <div className="text-gray-500">No results found</div>
                {activeFilters.length > 0 && (
                    <button
                        onClick={onClearFilters}
                        className="mt-2 text-blue-600 hover:text-blue-700"
                    >
                        Clear all filters
                    </button>
                )}
            </div>
        );
    }

    return (
        <div className={className}>
            {/* Active Filters */}
            <div className="mb-4">
                <FilterTags
                    filters={activeFilters}
                    onRemove={onRemoveFilter}
                    onClearAll={onClearFilters}
                />
            </div>

            {/* Controls */}
            <div className="flex justify-between items-center mb-4">
                <div className="flex items-center space-x-4">
                    <select
                        value={sortField}
                        onChange={(e) => setSortField(e.target.value)}
                        className="px-3 py-2 border border-gray-300 rounded-md"
                    >
                        <option value="score">Relevance</option>
                        <option value="type">Type</option>
                        <option value="id">ID</option>
                    </select>
                    <button
                        onClick={() => setSortDirection(prev => prev === 'asc' ? 'desc' : 'asc')}
                        className="p-2 text-gray-500 hover:text-gray-700"
                    >
                        {sortDirection === 'asc' ? '↑' : '↓'}
                    </button>
                </div>
                <div className="flex items-center space-x-2">
                    <button
                        onClick={() => setViewMode('list')}
                        className={`p-2 rounded ${
                            viewMode === 'list'
                                ? 'bg-blue-100 text-blue-700'
                                : 'text-gray-500 hover:text-gray-700'
                        }`}
                    >
                        List
                    </button>
                    <button
                        onClick={() => setViewMode('grid')}
                        className={`p-2 rounded ${
                            viewMode === 'grid'
                                ? 'bg-blue-100 text-blue-700'
                                : 'text-gray-500 hover:text-gray-700'
                        }`}
                    >
                        Grid
                    </button>
                </div>
            </div>

            {/* Results */}
            <div className={viewMode === 'grid' ? 'grid grid-cols-2 gap-4' : 'space-y-4'}>
                {sortedResults.map((result) => (
                    <div
                        key={result.id}
                        className={`bg-white rounded-lg shadow hover:shadow-md transition-shadow 
                            ${onSelectResult ? 'cursor-pointer' : ''}`}
                        onClick={() => onSelectResult?.(result)}
                    >
                        <div className="p-4">
                            <div className="flex justify-between items-start">
                                <div>
                                    <div className="flex items-center space-x-2">
                                        <span className="px-2 py-1 bg-gray-100 rounded text-sm">
                                            {result.type}
                                        </span>
                                        {result.score !== undefined && (
                                            <span className="text-sm text-gray-500">
                                                Score: {result.score.toFixed(2)}
                                            </span>
                                        )}
                                    </div>
                                    <h3 className="mt-2 font-medium">
                                        {result.properties.name || result.id}
                                    </h3>
                                </div>
                                {result.matchedFilters && result.matchedFilters.length > 0 && (
                                    <div className="text-sm text-gray-500">
                                        Matches: {result.matchedFilters.length}
                                    </div>
                                )}
                            </div>

                            <div className="mt-2">
                                {Object.entries(result.properties)
                                    .filter(([key]) => key !== 'name')
                                    .map(([key, value]) => (
                                        <div key={key} className="text-sm">
                                            <span className="font-medium">{key}:</span>{' '}
                                            {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                                        </div>
                                    ))}
                            </div>

                            {result.relationships && result.relationships.length > 0 && (
                                <div className="mt-2">
                                    <div className="text-sm font-medium">Relationships:</div>
                                    <div className="text-sm text-gray-600">
                                        {result.relationships.map((rel, index) => (
                                            <div key={index}>
                                                {rel.direction === 'incoming' ? '←' : '→'} {rel.type}
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default FilterResults;
