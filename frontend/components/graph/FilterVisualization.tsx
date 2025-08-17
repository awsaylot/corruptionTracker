import React from 'react';

interface FilterValue {
    field: string;
    operator: string;
    value: any;
    displayValue?: string;
    type: 'date-range' | 'numeric-range' | 'enum-multi' | 'string-pattern' | 'boolean' | 'array';
}

interface FilterTag {
    id: string;
    filter: FilterValue;
    color?: 'blue' | 'green' | 'yellow' | 'red' | 'purple' | 'gray';
}

interface FilterTagsProps {
    filters: FilterTag[];
    onRemove: (id: string) => void;
    onClearAll: () => void;
    className?: string;
}

const getFilterDisplayValue = (filter: FilterValue): string => {
    switch (filter.type) {
        case 'date-range':
            if (typeof filter.value === 'object' && filter.value !== null) {
                const { startDate, endDate } = filter.value;
                if (startDate && endDate) {
                    return `${new Date(startDate).toLocaleDateString()} - ${new Date(endDate).toLocaleDateString()}`;
                } else if (startDate) {
                    return `From ${new Date(startDate).toLocaleDateString()}`;
                } else if (endDate) {
                    return `Until ${new Date(endDate).toLocaleDateString()}`;
                }
            }
            return 'Invalid date range';

        case 'numeric-range':
            if (typeof filter.value === 'object' && filter.value !== null) {
                const { minValue, maxValue } = filter.value;
                if (minValue !== null && maxValue !== null) {
                    return `${minValue} - ${maxValue}`;
                } else if (minValue !== null) {
                    return `≥ ${minValue}`;
                } else if (maxValue !== null) {
                    return `≤ ${maxValue}`;
                }
            }
            return 'Invalid number range';

        case 'enum-multi':
            if (Array.isArray(filter.value)) {
                return filter.value.join(', ');
            }
            return String(filter.value);

        case 'boolean':
            return filter.value ? 'Yes' : 'No';

        case 'array':
            if (Array.isArray(filter.value)) {
                return `[${filter.value.length} items]`;
            }
            return String(filter.value);

        default:
            return filter.displayValue || String(filter.value);
    }
};

const getOperatorSymbol = (operator: string): string => {
    switch (operator) {
        case 'equals': return '=';
        case 'contains': return '∋';
        case 'startsWith': return '≺';
        case 'endsWith': return '≻';
        case 'greaterThan': return '>';
        case 'lessThan': return '<';
        case 'between': return '↔';
        case 'in': return '∈';
        case 'exists': return '∃';
        default: return operator;
    }
};

const getTagColorClasses = (color: FilterTag['color'] = 'blue') => {
    const baseClasses = 'inline-flex items-center rounded-full text-sm';
    switch (color) {
        case 'green':
            return `${baseClasses} bg-green-100 text-green-800`;
        case 'yellow':
            return `${baseClasses} bg-yellow-100 text-yellow-800`;
        case 'red':
            return `${baseClasses} bg-red-100 text-red-800`;
        case 'purple':
            return `${baseClasses} bg-purple-100 text-purple-800`;
        case 'gray':
            return `${baseClasses} bg-gray-100 text-gray-800`;
        default:
            return `${baseClasses} bg-blue-100 text-blue-800`;
    }
};

export const FilterTags: React.FC<FilterTagsProps> = ({
    filters,
    onRemove,
    onClearAll,
    className = ''
}) => {
    if (filters.length === 0) return null;

    return (
        <div className={`flex flex-wrap gap-2 ${className}`}>
            {filters.map(({ id, filter, color }) => (
                <span
                    key={id}
                    className={`${getTagColorClasses(color)} pl-3 pr-2 py-1`}
                >
                    <span className="font-medium mr-1">{filter.field}</span>
                    <span className="mx-1">{getOperatorSymbol(filter.operator)}</span>
                    <span className="mr-1">{getFilterDisplayValue(filter)}</span>
                    <button
                        onClick={() => onRemove(id)}
                        className="ml-1 text-lg leading-none hover:text-gray-600"
                    >
                        ×
                    </button>
                </span>
            ))}
            {filters.length > 1 && (
                <button
                    onClick={onClearAll}
                    className="text-sm text-gray-500 hover:text-gray-700 underline"
                >
                    Clear All
                </button>
            )}
        </div>
    );
};

interface FilterSummaryPanelProps {
    filters: FilterTag[];
    onRemove: (id: string) => void;
    onClearAll: () => void;
    onEdit?: (id: string) => void;
    className?: string;
}

export const FilterSummaryPanel: React.FC<FilterSummaryPanelProps> = ({
    filters,
    onRemove,
    onClearAll,
    onEdit,
    className = ''
}) => {
    if (filters.length === 0) {
        return (
            <div className={`p-4 bg-gray-50 rounded-lg text-gray-500 ${className}`}>
                No filters applied
            </div>
        );
    }

    const groupedFilters = filters.reduce((acc, filter) => {
        const type = filter.filter.type;
        if (!acc[type]) {
            acc[type] = [];
        }
        acc[type].push(filter);
        return acc;
    }, {} as Record<string, FilterTag[]>);

    return (
        <div className={`bg-white rounded-lg shadow ${className}`}>
            <div className="p-4 border-b">
                <div className="flex justify-between items-center">
                    <h3 className="text-lg font-medium">Active Filters</h3>
                    {filters.length > 1 && (
                        <button
                            onClick={onClearAll}
                            className="text-sm text-red-600 hover:text-red-700"
                        >
                            Clear All Filters
                        </button>
                    )}
                </div>
            </div>
            <div className="p-4">
                {Object.entries(groupedFilters).map(([type, typeFilters]) => (
                    <div key={type} className="mb-4 last:mb-0">
                        <h4 className="text-sm font-medium text-gray-700 mb-2">
                            {type.split('-').map(word => 
                                word.charAt(0).toUpperCase() + word.slice(1)
                            ).join(' ')} Filters
                        </h4>
                        <div className="space-y-2">
                            {typeFilters.map(({ id, filter, color }) => (
                                <div
                                    key={id}
                                    className="flex items-center justify-between p-2 bg-gray-50 rounded"
                                >
                                    <div className="flex items-center">
                                        <span className={`w-2 h-2 rounded-full mr-2 ${
                                            color === 'blue' ? 'bg-blue-500' :
                                            color === 'green' ? 'bg-green-500' :
                                            color === 'yellow' ? 'bg-yellow-500' :
                                            color === 'red' ? 'bg-red-500' :
                                            color === 'purple' ? 'bg-purple-500' :
                                            'bg-gray-500'
                                        }`} />
                                        <span className="text-sm">
                                            <span className="font-medium">{filter.field}</span>
                                            <span className="mx-1">{getOperatorSymbol(filter.operator)}</span>
                                            <span>{getFilterDisplayValue(filter)}</span>
                                        </span>
                                    </div>
                                    <div className="flex space-x-2">
                                        {onEdit && (
                                            <button
                                                onClick={() => onEdit(id)}
                                                className="text-sm text-blue-600 hover:text-blue-700"
                                            >
                                                Edit
                                            </button>
                                        )}
                                        <button
                                            onClick={() => onRemove(id)}
                                            className="text-sm text-red-600 hover:text-red-700"
                                        >
                                            Remove
                                        </button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export const getFilterSummary = (filters: FilterTag[]): string => {
    if (filters.length === 0) {
        return 'No filters applied';
    }

    const summaries = filters.map(({ filter }) => {
        const value = getFilterDisplayValue(filter);
        return `${filter.field} ${getOperatorSymbol(filter.operator)} ${value}`;
    });

    if (summaries.length === 1) {
        return summaries[0];
    }

    return `${summaries.length} filters applied: ${summaries.join('; ')}`;
};
