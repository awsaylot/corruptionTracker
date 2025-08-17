import React, { useState } from 'react';
import { useNotification } from '../../hooks/useNotification';

interface BaseFilterProps {
    label: string;
    value: any;
    onChange: (value: any) => void;
    isLoading?: boolean;
    disabled?: boolean;
    required?: boolean;
    className?: string;
}

// Date Range Filter
interface DateRangeFilterProps extends BaseFilterProps {
    startDate: Date | null;
    endDate: Date | null;
    minDate?: Date;
    maxDate?: Date;
    includeTime?: boolean;
}

export const DateRangeFilter: React.FC<DateRangeFilterProps> = ({
    label,
    startDate,
    endDate,
    onChange,
    minDate,
    maxDate,
    includeTime = false,
    isLoading,
    disabled,
    required,
    className
}) => {
    const handleStartDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        onChange({
            startDate: e.target.value ? new Date(e.target.value) : null,
            endDate
        });
    };

    const handleEndDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        onChange({
            startDate,
            endDate: e.target.value ? new Date(e.target.value) : null
        });
    };

    return (
        <div className={`space-y-2 ${className}`}>
            <label className="block text-sm font-medium text-gray-700">
                {label}
                {required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <div className="grid grid-cols-2 gap-4">
                <div>
                    <input
                        type={includeTime ? 'datetime-local' : 'date'}
                        value={startDate ? startDate.toISOString().slice(0, includeTime ? 16 : 10) : ''}
                        onChange={handleStartDateChange}
                        min={minDate?.toISOString().slice(0, includeTime ? 16 : 10)}
                        max={maxDate?.toISOString().slice(0, includeTime ? 16 : 10)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        disabled={isLoading || disabled}
                    />
                </div>
                <div>
                    <input
                        type={includeTime ? 'datetime-local' : 'date'}
                        value={endDate ? endDate.toISOString().slice(0, includeTime ? 16 : 10) : ''}
                        onChange={handleEndDateChange}
                        min={startDate?.toISOString().slice(0, includeTime ? 16 : 10)}
                        max={maxDate?.toISOString().slice(0, includeTime ? 16 : 10)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        disabled={isLoading || disabled}
                    />
                </div>
            </div>
        </div>
    );
};

// Numeric Range Filter
interface NumericRangeFilterProps extends BaseFilterProps {
    min?: number;
    max?: number;
    step?: number;
    minValue: number | null;
    maxValue: number | null;
}

export const NumericRangeFilter: React.FC<NumericRangeFilterProps> = ({
    label,
    minValue,
    maxValue,
    onChange,
    min,
    max,
    step = 1,
    isLoading,
    disabled,
    required,
    className
}) => {
    const handleMinValueChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value ? Number(e.target.value) : null;
        onChange({ minValue: value, maxValue });
    };

    const handleMaxValueChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value ? Number(e.target.value) : null;
        onChange({ minValue, maxValue: value });
    };

    return (
        <div className={`space-y-2 ${className}`}>
            <label className="block text-sm font-medium text-gray-700">
                {label}
                {required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <div className="grid grid-cols-2 gap-4">
                <div>
                    <input
                        type="number"
                        value={minValue ?? ''}
                        onChange={handleMinValueChange}
                        min={min}
                        max={max}
                        step={step}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        placeholder="Min"
                        disabled={isLoading || disabled}
                    />
                </div>
                <div>
                    <input
                        type="number"
                        value={maxValue ?? ''}
                        onChange={handleMaxValueChange}
                        min={minValue ?? min}
                        max={max}
                        step={step}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        placeholder="Max"
                        disabled={isLoading || disabled}
                    />
                </div>
            </div>
        </div>
    );
};

// Enum Multi-Select Filter
interface EnumMultiSelectFilterProps extends BaseFilterProps {
    options: Array<{
        value: string;
        label: string;
    }>;
    selectedValues: string[];
}

export const EnumMultiSelectFilter: React.FC<EnumMultiSelectFilterProps> = ({
    label,
    options,
    selectedValues,
    onChange,
    isLoading,
    disabled,
    required,
    className
}) => {
    const handleOptionToggle = (value: string) => {
        const newSelection = selectedValues.includes(value)
            ? selectedValues.filter(v => v !== value)
            : [...selectedValues, value];
        onChange(newSelection);
    };

    return (
        <div className={`space-y-2 ${className}`}>
            <label className="block text-sm font-medium text-gray-700">
                {label}
                {required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <div className="space-y-2">
                {options.map(option => (
                    <label key={option.value} className="flex items-center">
                        <input
                            type="checkbox"
                            checked={selectedValues.includes(option.value)}
                            onChange={() => handleOptionToggle(option.value)}
                            className="mr-2"
                            disabled={isLoading || disabled}
                        />
                        {option.label}
                    </label>
                ))}
            </div>
        </div>
    );
};

// String Pattern Filter
interface StringPatternFilterProps extends BaseFilterProps {
    pattern?: string;
    placeholder?: string;
    matchCase?: boolean;
    useRegex?: boolean;
}

export const StringPatternFilter: React.FC<StringPatternFilterProps> = ({
    label,
    value,
    onChange,
    pattern,
    placeholder,
    matchCase = false,
    useRegex = false,
    isLoading,
    disabled,
    required,
    className
}) => {
    const { addNotification } = useNotification();

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newValue = e.target.value;
        if (useRegex) {
            try {
                // Test if the pattern is valid regex
                new RegExp(newValue);
                onChange(newValue);
            } catch (error) {
                addNotification('error', 'Invalid regular expression pattern');
            }
        } else {
            onChange(newValue);
        }
    };

    return (
        <div className={`space-y-2 ${className}`}>
            <label className="block text-sm font-medium text-gray-700">
                {label}
                {required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <div className="space-y-2">
                <input
                    type="text"
                    value={value}
                    onChange={handleChange}
                    pattern={pattern}
                    placeholder={placeholder}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    disabled={isLoading || disabled}
                />
                {useRegex && (
                    <div className="text-sm text-gray-500">
                        Using regular expression pattern matching
                    </div>
                )}
            </div>
        </div>
    );
};

// Boolean Filter
interface BooleanFilterProps extends BaseFilterProps {
    trueLabel?: string;
    falseLabel?: string;
    includeNull?: boolean;
    nullLabel?: string;
}

export const BooleanFilter: React.FC<BooleanFilterProps> = ({
    label,
    value,
    onChange,
    trueLabel = 'Yes',
    falseLabel = 'No',
    includeNull = false,
    nullLabel = 'Any',
    isLoading,
    disabled,
    required,
    className
}) => {
    return (
        <div className={`space-y-2 ${className}`}>
            <label className="block text-sm font-medium text-gray-700">
                {label}
                {required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <div className="flex space-x-4">
                {includeNull && (
                    <label className="flex items-center">
                        <input
                            type="radio"
                            checked={value === null}
                            onChange={() => onChange(null)}
                            className="mr-2"
                            disabled={isLoading || disabled}
                        />
                        {nullLabel}
                    </label>
                )}
                <label className="flex items-center">
                    <input
                        type="radio"
                        checked={value === true}
                        onChange={() => onChange(true)}
                        className="mr-2"
                        disabled={isLoading || disabled}
                    />
                    {trueLabel}
                </label>
                <label className="flex items-center">
                    <input
                        type="radio"
                        checked={value === false}
                        onChange={() => onChange(false)}
                        className="mr-2"
                        disabled={isLoading || disabled}
                    />
                    {falseLabel}
                </label>
            </div>
        </div>
    );
};

// Array Filter
interface ArrayFilterProps extends BaseFilterProps {
    items: string[];
    placeholder?: string;
    maxItems?: number;
    allowDuplicates?: boolean;
}

export const ArrayFilter: React.FC<ArrayFilterProps> = ({
    label,
    items,
    onChange,
    placeholder = 'Add item...',
    maxItems,
    allowDuplicates = false,
    isLoading,
    disabled,
    required,
    className
}) => {
    const [newItem, setNewItem] = useState('');
    const { addNotification } = useNotification();

    const handleAddItem = () => {
        if (!newItem.trim()) return;

        if (maxItems && items.length >= maxItems) {
            addNotification('error', `Maximum of ${maxItems} items allowed`);
            return;
        }

        if (!allowDuplicates && items.includes(newItem.trim())) {
            addNotification('error', 'Duplicate items are not allowed');
            return;
        }

        onChange([...items, newItem.trim()]);
        setNewItem('');
    };

    const handleRemoveItem = (index: number) => {
        onChange(items.filter((_, i) => i !== index));
    };

    const handleKeyPress = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            handleAddItem();
        }
    };

    return (
        <div className={`space-y-2 ${className}`}>
            <label className="block text-sm font-medium text-gray-700">
                {label}
                {required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <div className="space-y-2">
                <div className="flex space-x-2">
                    <input
                        type="text"
                        value={newItem}
                        onChange={(e) => setNewItem(e.target.value)}
                        onKeyPress={handleKeyPress}
                        placeholder={placeholder}
                        className="flex-1 px-3 py-2 border border-gray-300 rounded-md"
                        disabled={isLoading || disabled || (maxItems && items.length >= maxItems)}
                    />
                    <button
                        onClick={handleAddItem}
                        disabled={isLoading || disabled || !newItem.trim() || (maxItems && items.length >= maxItems)}
                        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400"
                    >
                        Add
                    </button>
                </div>
                <div className="space-y-1">
                    {items.map((item, index) => (
                        <div key={index} className="flex items-center justify-between bg-gray-50 px-3 py-2 rounded">
                            <span>{item}</span>
                            <button
                                onClick={() => handleRemoveItem(index)}
                                disabled={isLoading || disabled}
                                className="text-red-600 hover:text-red-700"
                            >
                                Remove
                            </button>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export interface FilterComponentProps {
    type: 'date-range' | 'numeric-range' | 'enum-multi' | 'string-pattern' | 'boolean' | 'array';
    props: BaseFilterProps & 
        Partial<DateRangeFilterProps> &
        Partial<NumericRangeFilterProps> &
        Partial<EnumMultiSelectFilterProps> &
        Partial<StringPatternFilterProps> &
        Partial<BooleanFilterProps> &
        Partial<ArrayFilterProps>;
}

export const FilterComponent: React.FC<FilterComponentProps> = ({ type, props }) => {
    switch (type) {
        case 'date-range':
            return <DateRangeFilter {...props as DateRangeFilterProps} />;
        case 'numeric-range':
            return <NumericRangeFilter {...props as NumericRangeFilterProps} />;
        case 'enum-multi':
            return <EnumMultiSelectFilter {...props as EnumMultiSelectFilterProps} />;
        case 'string-pattern':
            return <StringPatternFilter {...props as StringPatternFilterProps} />;
        case 'boolean':
            return <BooleanFilter {...props as BooleanFilterProps} />;
        case 'array':
            return <ArrayFilter {...props as ArrayFilterProps} />;
        default:
            return null;
    }
};
