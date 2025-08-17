// components/NodeForm.tsx - Updated with full integration
import { useState } from 'react';
import { useRouter } from 'next/router';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';
import { api } from '../../utils/api';
import LoadingSpinner from '../ui/LoadingSpinner';

interface FormData {
    name: string;
    notes?: string;
    aliases?: string;
    type?: string;
    date?: string;
    sources?: string;
    description?: string;
    title?: string;
}

interface NodeFormProps {
    onComplete?: (nodeId: string) => void;
    onError?: (error: string) => void;
    onSuccess?: (message: string) => void;
    redirectOnSuccess?: boolean;
}

const NodeForm: React.FC<NodeFormProps> = ({ 
    onComplete, 
    onError, 
    onSuccess,
    redirectOnSuccess = false 
}) => {
    const router = useRouter();
    const [nodeType, setNodeType] = useState<'Person' | 'Organization' | 'Event'>('Person');
    const [formData, setFormData] = useState<FormData>({
        name: '',
        notes: '',
        aliases: '',
    });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});

    const validateForm = (): boolean => {
        const errors: Record<string, string> = {};

        // Name/Title is required
        if (!formData.name?.trim()) {
            errors.name = nodeType === 'Event' ? 'Title is required' : 'Name is required';
        }

        // Date validation for events
        if (nodeType === 'Event' && formData.date) {
            const selectedDate = new Date(formData.date);
            const today = new Date();
            if (selectedDate > today) {
                errors.date = 'Event date cannot be in the future';
            }
        }

        // URL validation for sources
        if (formData.sources?.trim()) {
            const urls = formData.sources.split(',').map(s => s.trim());
            for (const url of urls) {
                try {
                    new URL(url);
                } catch {
                    errors.sources = 'Please provide valid URLs separated by commas';
                    break;
                }
            }
        }

        setValidationErrors(errors);
        return Object.keys(errors).length === 0;
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        
        if (!validateForm()) {
            return;
        }

        setError(null);
        setLoading(true);

        try {
            const properties: Record<string, any> = {
                notes: formData.notes?.trim() || ''
            };

            // Handle type-specific properties
            if (nodeType === 'Event') {
                properties.title = formData.name.trim();
                if (formData.description?.trim()) properties.description = formData.description.trim();
                if (formData.date) properties.date = formData.date;
                if (formData.sources?.trim()) {
                    properties.sources = formData.sources
                        .split(',')
                        .map(s => s.trim())
                        .filter(Boolean);
                }
            } else {
                properties.name = formData.name.trim();
                if (formData.aliases?.trim()) {
                    properties.aliases = formData.aliases
                        .split(',')
                        .map(s => s.trim())
                        .filter(Boolean);
                }
            }

            // Organization-specific properties
            if (nodeType === 'Organization' && formData.type?.trim()) {
                properties.type = formData.type.trim();
            }

            const nodeId = await api.createNode(nodeType, properties);
            
            // Reset form
            setFormData({
                name: '',
                notes: '',
                aliases: '',
                type: '',
                date: '',
                sources: '',
                description: '',
            });
            setValidationErrors({});

            // Success callbacks
            const successMessage = `${nodeType} "${properties.name || properties.title}" created successfully`;
            onSuccess?.(successMessage);
            onComplete?.(nodeId);

            // Redirect if requested
            if (redirectOnSuccess) {
                setTimeout(() => {
                    router.push(`/nodes/${nodeId}`);
                }, 1500);
            }
            
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'Failed to create node';
            setError(errorMessage);
            onError?.(errorMessage);
        } finally {
            setLoading(false);
        }
    };

    const handleFieldChange = (field: keyof FormData, value: string) => {
        setFormData(prev => ({ ...prev, [field]: value }));
        
        // Clear validation error for this field
        if (validationErrors[field]) {
            setValidationErrors(prev => {
                const newErrors = { ...prev };
                delete newErrors[field];
                return newErrors;
            });
        }
    };

    const getFieldError = (field: string): string | undefined => {
        return validationErrors[field];
    };

    return (
        <div className="bg-white rounded-lg shadow-sm">
            <div className="px-6 py-4 border-b border-gray-200">
                <h2 className="text-lg font-semibold text-gray-900">Create New {nodeType}</h2>
            </div>
            
            <form onSubmit={handleSubmit} className="px-6 py-4 space-y-4">
                {/* Global Error */}
                {error && (
                    <div className="bg-red-50 border-l-4 border-red-400 p-4">
                        <div className="flex">
                            <div className="flex-shrink-0">
                                <ExclamationTriangleIcon className="h-5 w-5 text-red-400" />
                            </div>
                            <div className="ml-3">
                                <p className="text-sm text-red-700">{error}</p>
                            </div>
                        </div>
                    </div>
                )}

                {/* Node Type Selection */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">Node Type</label>
                    <select
                        value={nodeType}
                        onChange={(e) => setNodeType(e.target.value as any)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
                        disabled={loading}
                    >
                        <option value="Person">Person</option>
                        <option value="Organization">Organization</option>
                        <option value="Event">Event</option>
                    </select>
                </div>

                {/* Name/Title Field */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        {nodeType === 'Event' ? 'Title' : 'Name'} *
                    </label>
                    <input
                        type="text"
                        required
                        value={formData.name}
                        onChange={(e) => handleFieldChange('name', e.target.value)}
                        className={`w-full px-3 py-2 border rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 ${
                            getFieldError('name') ? 'border-red-300' : 'border-gray-300'
                        }`}
                        placeholder={nodeType === 'Event' ? 'Enter event title...' : 'Enter name...'}
                        disabled={loading}
                    />
                    {getFieldError('name') && (
                        <p className="mt-1 text-sm text-red-600">{getFieldError('name')}</p>
                    )}
                </div>

                {/* Type-specific Fields */}
                {nodeType !== 'Event' && (
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Aliases (comma-separated)
                        </label>
                        <input
                            type="text"
                            value={formData.aliases}
                            onChange={(e) => handleFieldChange('aliases', e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
                            placeholder="e.g., John Doe, JD, Johnny"
                            disabled={loading}
                        />
                        <p className="mt-1 text-xs text-gray-500">
                            Enter alternative names or nicknames, separated by commas
                        </p>
                    </div>
                )}

                {/* Organization Type */}
                {nodeType === 'Organization' && (
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">Organization Type</label>
                        <select
                            value={formData.type}
                            onChange={(e) => handleFieldChange('type', e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
                            disabled={loading}
                        >
                            <option value="">Select Type...</option>
                            <option value="Government">Government</option>
                            <option value="Corporation">Corporation</option>
                            <option value="NGO">NGO</option>
                            <option value="Educational">Educational</option>
                            <option value="Religious">Religious</option>
                            <option value="Other">Other</option>
                        </select>
                    </div>
                )}

                {/* Event-specific Fields */}
                {nodeType === 'Event' && (
                    <>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">Description</label>
                            <textarea
                                value={formData.description}
                                onChange={(e) => handleFieldChange('description', e.target.value)}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
                                rows={3}
                                placeholder="Describe what happened..."
                                disabled={loading}
                            />
                        </div>
                        
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">Date</label>
                            <input
                                type="date"
                                value={formData.date}
                                onChange={(e) => handleFieldChange('date', e.target.value)}
                                className={`w-full px-3 py-2 border rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 ${
                                    getFieldError('date') ? 'border-red-300' : 'border-gray-300'
                                }`}
                                max={new Date().toISOString().split('T')[0]}
                                disabled={loading}
                            />
                            {getFieldError('date') && (
                                <p className="mt-1 text-sm text-red-600">{getFieldError('date')}</p>
                            )}
                        </div>
                        
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Sources (comma-separated URLs)
                            </label>
                            <textarea
                                value={formData.sources}
                                onChange={(e) => handleFieldChange('sources', e.target.value)}
                                className={`w-full px-3 py-2 border rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 ${
                                    getFieldError('sources') ? 'border-red-300' : 'border-gray-300'
                                }`}
                                rows={3}
                                placeholder="https://example.com, https://another.com"
                                disabled={loading}
                            />
                            {getFieldError('sources') && (
                                <p className="mt-1 text-sm text-red-600">{getFieldError('sources')}</p>
                            )}
                            <p className="mt-1 text-xs text-gray-500">
                                Enter source URLs, separated by commas
                            </p>
                        </div>
                    </>
                )}

                {/* Notes Field (Common) */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">Notes</label>
                    <textarea
                        value={formData.notes}
                        onChange={(e) => handleFieldChange('notes', e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
                        rows={3}
                        placeholder="Additional notes or information..."
                        disabled={loading}
                    />
                </div>

                {/* Form Actions */}
                <div className="flex justify-end space-x-3 pt-6 border-t border-gray-200">
                    <button
                        type="button"
                        onClick={() => {
                            setFormData({
                                name: '',
                                notes: '',
                                aliases: '',
                                type: '',
                                date: '',
                                sources: '',
                                description: '',
                            });
                            setValidationErrors({});
                            setError(null);
                        }}
                        className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                        disabled={loading}
                    >
                        Reset
                    </button>
                    
                    <button
                        type="submit"
                        disabled={loading || !formData.name?.trim()}
                        className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:bg-gray-300 disabled:cursor-not-allowed"
                    >
                        {loading ? (
                            <>
                                <LoadingSpinner size="small" color="white" />
                                <span className="ml-2">Creating...</span>
                            </>
                        ) : (
                            `Create ${nodeType}`
                        )}
                    </button>
                </div>
            </form>
        </div>
    );
};

export default NodeForm;