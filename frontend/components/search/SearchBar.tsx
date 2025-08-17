// components/SearchBar.tsx - Updated with enhanced error handling
import { useState, useCallback, useEffect } from 'react';
import { api, Node } from '../../utils/api';
import debounce from 'lodash/debounce';
import { MagnifyingGlassIcon, ExclamationTriangleIcon } from '@heroicons/react/24/outline';

interface SearchBarProps {
    onSelectNode: (node: Node) => void;
    onError?: (error: string) => void;
    placeholder?: string;
}

const SearchBar: React.FC<SearchBarProps> = ({ 
    onSelectNode, 
    onError,
    placeholder = "Search for people, organizations, or events..."
}) => {
    const [query, setQuery] = useState('');
    const [results, setResults] = useState<Node[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [mode, setMode] = useState<'basic' | 'advanced'>('basic');
    const [showPaths, setShowPaths] = useState(false);
    const [showResults, setShowResults] = useState(false);

    const debouncedSearch = useCallback(
        debounce(async (searchQuery: string) => {
            if (!searchQuery.trim()) {
                setResults([]);
                setError(null);
                setShowResults(false);
                return;
            }

            setLoading(true);
            setError(null);
            
            try {
                const response = await api.searchNodes(searchQuery, {
                    mode,
                    includePath: showPaths
                });
                
                let nodes: Node[];
                if ('nodes' in response) {
                    nodes = response.nodes;
                } else {
                    nodes = response;
                }
                
                setResults(nodes);
                setShowResults(true);
                
                if (nodes.length === 0) {
                    setError('No results found');
                }
                
            } catch (err) {
                const errorMessage = err instanceof Error ? err.message : 'Search failed';
                setError(errorMessage);
                setResults([]);
                setShowResults(false);
                onError?.(errorMessage);
                console.error('Search failed:', err);
            } finally {
                setLoading(false);
            }
        }, 300),
        [mode, showPaths, onError]
    );

    useEffect(() => {
        debouncedSearch(query);
    }, [query, debouncedSearch]);

    const handleSelectNode = (node: Node) => {
        try {
            onSelectNode(node);
            setQuery('');
            setResults([]);
            setShowResults(false);
            setError(null);
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'Failed to select node';
            setError(errorMessage);
            onError?.(errorMessage);
        }
    };

    const handleClearSearch = () => {
        setQuery('');
        setResults([]);
        setError(null);
        setShowResults(false);
    };

    return (
        <div className="relative">
            <div className="space-y-3">
                {/* Main Search Input */}
                <div className="flex gap-2">
                    <div className="relative flex-1">
                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                            <MagnifyingGlassIcon className="h-5 w-5 text-gray-400" aria-hidden="true" />
                        </div>
                        
                        <input
                            type="text"
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                            placeholder={placeholder}
                            className={`w-full pl-10 pr-10 py-2 border rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 ${
                                error ? 'border-red-300' : 'border-gray-300'
                            }`}
                        />
                        
                        {/* Loading/Clear Button */}
                        <div className="absolute inset-y-0 right-0 pr-3 flex items-center">
                            {loading ? (
                                <div className="animate-spin h-4 w-4 border-2 border-blue-500 rounded-full border-t-transparent" />
                            ) : query ? (
                                <button
                                    type="button"
                                    onClick={handleClearSearch}
                                    className="text-gray-400 hover:text-gray-500"
                                >
                                    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                    </svg>
                                </button>
                            ) : null}
                        </div>
                    </div>

                    {/* Search Mode Selector */}
                    <select
                        value={mode}
                        onChange={(e) => setMode(e.target.value as 'basic' | 'advanced')}
                        className="py-2 pl-3 pr-8 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 text-sm"
                    >
                        <option value="basic">Basic</option>
                        <option value="advanced">Advanced</option>
                    </select>
                </div>

                {/* Advanced Options */}
                {mode === 'advanced' && (
                    <div className="flex items-center justify-between bg-gray-50 px-3 py-2 rounded-md">
                        <label className="flex items-center text-sm text-gray-600">
                            <input
                                type="checkbox"
                                checked={showPaths}
                                onChange={(e) => setShowPaths(e.target.checked)}
                                className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded mr-2"
                            />
                            Include connection paths
                        </label>
                        <span className="text-xs text-gray-500">
                            {results.length} result{results.length !== 1 ? 's' : ''}
                        </span>
                    </div>
                )}

                {/* Inline Error Message */}
                {error && !loading && (
                    <div className="flex items-center space-x-2 text-sm text-red-600">
                        <ExclamationTriangleIcon className="h-4 w-4" />
                        <span>{error}</span>
                    </div>
                )}
            </div>

            {/* Search Results Dropdown */}
            {showResults && results.length > 0 && (
                <div className="absolute w-full mt-1 bg-white border rounded-lg shadow-lg max-h-96 overflow-y-auto z-20">
                    <div className="p-2 border-b bg-gray-50">
                        <div className="flex items-center justify-between text-xs text-gray-500">
                            <span>{results.length} result{results.length !== 1 ? 's' : ''} found</span>
                            <button
                                onClick={() => setShowResults(false)}
                                className="text-gray-400 hover:text-gray-500"
                            >
                                Close
                            </button>
                        </div>
                    </div>
                    
                    {results.map((node: any) => (
                        <button
                            key={node.id}
                            onClick={() => handleSelectNode(node)}
                            className="w-full px-4 py-3 text-left hover:bg-gray-100 focus:bg-gray-100 focus:outline-none border-b border-gray-100 last:border-b-0"
                        >
                            <div className="flex items-center justify-between">
                                <div className="flex items-center space-x-3">
                                    <div className={`w-3 h-3 rounded-full flex-shrink-0 ${
                                        node.type === 'Person' ? 'bg-green-500' :
                                        node.type === 'Organization' ? 'bg-blue-500' :
                                        'bg-yellow-500'
                                    }`} />
                                    <div className="min-w-0 flex-1">
                                        <div className="font-medium text-gray-900 truncate">
                                            {node.properties.name || node.properties.title || `Node ${node.id}`}
                                        </div>
                                        {node.properties.aliases && node.properties.aliases.length > 0 && (
                                            <div className="text-xs text-gray-500 truncate">
                                                Also known as: {node.properties.aliases.join(', ')}
                                            </div>
                                        )}
                                    </div>
                                </div>
                                
                                <div className="flex items-center space-x-2 flex-shrink-0">
                                    {/* Search Score */}
                                    {'score' in node && (
                                        <span className="text-xs bg-gray-100 text-gray-600 px-2 py-1 rounded">
                                            {Math.round(node.score * 100)}%
                                        </span>
                                    )}
                                    
                                    {/* Connection Count */}
                                    {'connectionCount' in node && (
                                        <span className="text-xs text-gray-500">
                                            {node.connectionCount} conn.
                                        </span>
                                    )}
                                    
                                    {/* Node Type Badge */}
                                    <span className={`text-xs px-2 py-1 rounded-full ${
                                        node.type === 'Person' ? 'bg-green-100 text-green-800' :
                                        node.type === 'Organization' ? 'bg-blue-100 text-blue-800' :
                                        'bg-yellow-100 text-yellow-800'
                                    }`}>
                                        {node.type}
                                    </span>
                                </div>
                            </div>
                        </button>
                    ))}
                </div>
            )}
            
            {/* Click outside to close results */}
            {showResults && (
                <div 
                    className="fixed inset-0 z-10" 
                    onClick={() => setShowResults(false)}
                />
            )}
        </div>
    );
};

export default SearchBar;