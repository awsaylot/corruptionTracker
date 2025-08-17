import React, { useState, useEffect } from 'react';
import { useNotification } from '../../hooks/useNotification';

interface NodeType {
    id: string;
    name: string;
    description: string;
    properties: PropertyDefinition[];
    validRelations: RelationDefinition[];
}

interface PropertyDefinition {
    name: string;
    type: 'string' | 'number' | 'boolean' | 'date' | 'array' | 'object';
    required: boolean;
    defaultValue?: any;
    validation?: {
        pattern?: string;
        min?: number;
        max?: number;
        enum?: any[];
    };
}

interface RelationDefinition {
    type: string;
    targetType: string;
    description: string;
    cardinality: 'one-to-one' | 'one-to-many' | 'many-to-many';
    properties?: PropertyDefinition[];
}

interface NodeTypeManagerProps {
    onCreateType: (nodeType: Omit<NodeType, 'id'>) => Promise<void>;
    onUpdateType: (id: string, nodeType: Partial<NodeType>) => Promise<void>;
    onDeleteType: (id: string) => Promise<void>;
    existingTypes: NodeType[];
    isLoading?: boolean;
}

const NodeTypeManager: React.FC<NodeTypeManagerProps> = ({
    onCreateType,
    onUpdateType,
    onDeleteType,
    existingTypes,
    isLoading
}) => {
    const { addNotification } = useNotification();
    const [mode, setMode] = useState<'create' | 'edit'>('create');
    const [selectedTypeId, setSelectedTypeId] = useState<string>('');
    const [formData, setFormData] = useState<Omit<NodeType, 'id'>>({
        name: '',
        description: '',
        properties: [],
        validRelations: []
    });

    // Reset form when switching between create/edit modes
    useEffect(() => {
        if (mode === 'edit' && selectedTypeId) {
            const selectedType = existingTypes.find(t => t.id === selectedTypeId);
            if (selectedType) {
                const { id, ...typeData } = selectedType;
                setFormData(typeData);
            }
        } else if (mode === 'create') {
            setFormData({
                name: '',
                description: '',
                properties: [],
                validRelations: []
            });
        }
    }, [mode, selectedTypeId, existingTypes]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            if (mode === 'create') {
                await onCreateType(formData);
                addNotification('success', 'Node type created successfully');
                setFormData({
                    name: '',
                    description: '',
                    properties: [],
                    validRelations: []
                });
            } else {
                await onUpdateType(selectedTypeId, formData);
                addNotification('success', 'Node type updated successfully');
            }
        } catch (error) {
            addNotification('error', error instanceof Error ? error.message : 'Operation failed');
        }
    };

    const handleAddProperty = () => {
        setFormData(prev => ({
            ...prev,
            properties: [
                ...prev.properties,
                {
                    name: '',
                    type: 'string',
                    required: false
                }
            ]
        }));
    };

    const handleRemoveProperty = (index: number) => {
        setFormData(prev => ({
            ...prev,
            properties: prev.properties.filter((_, i) => i !== index)
        }));
    };

    const handlePropertyChange = (index: number, field: keyof PropertyDefinition, value: any) => {
        setFormData(prev => ({
            ...prev,
            properties: prev.properties.map((prop, i) => 
                i === index ? { ...prop, [field]: value } : prop
            )
        }));
    };

    const handleAddRelation = () => {
        setFormData(prev => ({
            ...prev,
            validRelations: [
                ...prev.validRelations,
                {
                    type: '',
                    targetType: '',
                    description: '',
                    cardinality: 'many-to-many',
                    properties: []
                }
            ]
        }));
    };

    const handleRemoveRelation = (index: number) => {
        setFormData(prev => ({
            ...prev,
            validRelations: prev.validRelations.filter((_, i) => i !== index)
        }));
    };

    const handleRelationChange = (index: number, field: keyof RelationDefinition, value: any) => {
        setFormData(prev => ({
            ...prev,
            validRelations: prev.validRelations.map((rel, i) => 
                i === index ? { ...rel, [field]: value } : rel
            )
        }));
    };

    return (
        <div className="bg-white rounded-lg shadow p-6">
            <div className="mb-6 flex justify-between items-center">
                <h2 className="text-xl font-semibold">Node Type Manager</h2>
                <div className="space-x-4">
                    <button
                        className={`px-4 py-2 rounded ${
                            mode === 'create' 
                                ? 'bg-blue-600 text-white' 
                                : 'bg-gray-200 text-gray-700'
                        }`}
                        onClick={() => setMode('create')}
                    >
                        Create New
                    </button>
                    <button
                        className={`px-4 py-2 rounded ${
                            mode === 'edit' 
                                ? 'bg-blue-600 text-white' 
                                : 'bg-gray-200 text-gray-700'
                        }`}
                        onClick={() => setMode('edit')}
                    >
                        Edit Existing
                    </button>
                </div>
            </div>

            {mode === 'edit' && (
                <div className="mb-6">
                    <select
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        value={selectedTypeId}
                        onChange={(e) => setSelectedTypeId(e.target.value)}
                    >
                        <option value="">Select a type to edit</option>
                        {existingTypes.map(type => (
                            <option key={type.id} value={type.id}>
                                {type.name}
                            </option>
                        ))}
                    </select>
                </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-6">
                {/* Basic Information */}
                <div className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700">
                            Name
                        </label>
                        <input
                            type="text"
                            value={formData.name}
                            onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                            className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700">
                            Description
                        </label>
                        <textarea
                            value={formData.description}
                            onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                            className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                            rows={3}
                        />
                    </div>
                </div>

                {/* Properties Section */}
                <div>
                    <div className="flex justify-between items-center mb-4">
                        <h3 className="text-lg font-medium">Properties</h3>
                        <button
                            type="button"
                            onClick={handleAddProperty}
                            className="px-3 py-1 text-sm bg-blue-100 text-blue-600 rounded hover:bg-blue-200"
                        >
                            Add Property
                        </button>
                    </div>
                    {formData.properties.map((prop, index) => (
                        <div key={index} className="flex gap-4 items-start mb-4 p-4 bg-gray-50 rounded">
                            <div className="flex-1">
                                <input
                                    type="text"
                                    value={prop.name}
                                    onChange={(e) => handlePropertyChange(index, 'name', e.target.value)}
                                    placeholder="Property name"
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                                />
                            </div>
                            <div className="flex-1">
                                <select
                                    value={prop.type}
                                    onChange={(e) => handlePropertyChange(index, 'type', e.target.value)}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                                >
                                    <option value="string">String</option>
                                    <option value="number">Number</option>
                                    <option value="boolean">Boolean</option>
                                    <option value="date">Date</option>
                                    <option value="array">Array</option>
                                    <option value="object">Object</option>
                                </select>
                            </div>
                            <div className="flex items-center">
                                <input
                                    type="checkbox"
                                    checked={prop.required}
                                    onChange={(e) => handlePropertyChange(index, 'required', e.target.checked)}
                                    className="mr-2"
                                />
                                <span className="text-sm">Required</span>
                            </div>
                            <button
                                type="button"
                                onClick={() => handleRemoveProperty(index)}
                                className="px-2 py-1 text-red-600 hover:bg-red-50 rounded"
                            >
                                Remove
                            </button>
                        </div>
                    ))}
                </div>

                {/* Relations Section */}
                <div>
                    <div className="flex justify-between items-center mb-4">
                        <h3 className="text-lg font-medium">Valid Relations</h3>
                        <button
                            type="button"
                            onClick={handleAddRelation}
                            className="px-3 py-1 text-sm bg-blue-100 text-blue-600 rounded hover:bg-blue-200"
                        >
                            Add Relation
                        </button>
                    </div>
                    {formData.validRelations.map((relation, index) => (
                        <div key={index} className="mb-4 p-4 bg-gray-50 rounded space-y-4">
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">
                                        Relation Type
                                    </label>
                                    <input
                                        type="text"
                                        value={relation.type}
                                        onChange={(e) => handleRelationChange(index, 'type', e.target.value)}
                                        className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md"
                                        placeholder="e.g., OWNS, MANAGES"
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">
                                        Target Type
                                    </label>
                                    <select
                                        value={relation.targetType}
                                        onChange={(e) => handleRelationChange(index, 'targetType', e.target.value)}
                                        className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md"
                                    >
                                        <option value="">Select target type</option>
                                        {existingTypes.map(type => (
                                            <option key={type.id} value={type.name}>
                                                {type.name}
                                            </option>
                                        ))}
                                    </select>
                                </div>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700">
                                    Description
                                </label>
                                <input
                                    type="text"
                                    value={relation.description}
                                    onChange={(e) => handleRelationChange(index, 'description', e.target.value)}
                                    className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700">
                                    Cardinality
                                </label>
                                <select
                                    value={relation.cardinality}
                                    onChange={(e) => handleRelationChange(
                                        index,
                                        'cardinality',
                                        e.target.value as RelationDefinition['cardinality']
                                    )}
                                    className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md"
                                >
                                    <option value="one-to-one">One-to-One</option>
                                    <option value="one-to-many">One-to-Many</option>
                                    <option value="many-to-many">Many-to-Many</option>
                                </select>
                            </div>
                            <button
                                type="button"
                                onClick={() => handleRemoveRelation(index)}
                                className="px-3 py-1 text-red-600 hover:bg-red-50 rounded"
                            >
                                Remove Relation
                            </button>
                        </div>
                    ))}
                </div>

                {/* Submit Buttons */}
                <div className="flex justify-end space-x-4">
                    {mode === 'edit' && (
                        <button
                            type="button"
                            onClick={() => {
                                if (selectedTypeId && window.confirm('Are you sure you want to delete this node type?')) {
                                    onDeleteType(selectedTypeId)
                                        .then(() => {
                                            addNotification('success', 'Node type deleted successfully');
                                            setMode('create');
                                        })
                                        .catch(error => {
                                            addNotification('error', error instanceof Error ? error.message : 'Delete failed');
                                        });
                                }
                            }}
                            className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
                            disabled={isLoading}
                        >
                            Delete Type
                        </button>
                    )}
                    <button
                        type="submit"
                        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                        disabled={isLoading}
                    >
                        {isLoading ? 'Processing...' : mode === 'create' ? 'Create Type' : 'Update Type'}
                    </button>
                </div>
            </form>
        </div>
    );
};

export default NodeTypeManager;
