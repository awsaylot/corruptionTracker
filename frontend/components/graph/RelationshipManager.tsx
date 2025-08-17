import React from 'react';
import { useNotification } from '../../hooks/useNotification';

interface NodeTypeRelation {
    sourceType: string;
    targetType: string;
    relationType: string;
    cardinality: 'one-to-one' | 'one-to-many' | 'many-to-many';
    properties: {
        name: string;
        type: string;
        required: boolean;
    }[];
}

interface RelationshipManagerProps {
    nodeTypes: {
        id: string;
        name: string;
    }[];
    existingRelations: NodeTypeRelation[];
    onAddRelation: (relation: NodeTypeRelation) => Promise<void>;
    onUpdateRelation: (
        sourceType: string,
        targetType: string,
        relationType: string,
        updates: Partial<NodeTypeRelation>
    ) => Promise<void>;
    onDeleteRelation: (
        sourceType: string,
        targetType: string,
        relationType: string
    ) => Promise<void>;
    isLoading?: boolean;
}

const RelationshipManager: React.FC<RelationshipManagerProps> = ({
    nodeTypes,
    existingRelations,
    onAddRelation,
    onUpdateRelation,
    onDeleteRelation,
    isLoading
}) => {
    const { addNotification } = useNotification();
    const [selectedRelation, setSelectedRelation] = React.useState<NodeTypeRelation | null>(null);
    const [showForm, setShowForm] = React.useState(false);
    const [formData, setFormData] = React.useState<NodeTypeRelation>({
        sourceType: '',
        targetType: '',
        relationType: '',
        cardinality: 'many-to-many',
        properties: []
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            if (selectedRelation) {
                await onUpdateRelation(
                    selectedRelation.sourceType,
                    selectedRelation.targetType,
                    selectedRelation.relationType,
                    formData
                );
                addNotification('success', 'Relationship updated successfully');
            } else {
                await onAddRelation(formData);
                addNotification('success', 'Relationship created successfully');
            }
            setShowForm(false);
            setSelectedRelation(null);
            setFormData({
                sourceType: '',
                targetType: '',
                relationType: '',
                cardinality: 'many-to-many',
                properties: []
            });
        } catch (error) {
            addNotification('error', error instanceof Error ? error.message : 'Operation failed');
        }
    };

    const handleDelete = async (relation: NodeTypeRelation) => {
        if (window.confirm('Are you sure you want to delete this relationship?')) {
            try {
                await onDeleteRelation(
                    relation.sourceType,
                    relation.targetType,
                    relation.relationType
                );
                addNotification('success', 'Relationship deleted successfully');
            } catch (error) {
                addNotification('error', error instanceof Error ? error.message : 'Delete failed');
            }
        }
    };

    const handleAddProperty = () => {
        setFormData(prev => ({
            ...prev,
            properties: [
                ...prev.properties,
                { name: '', type: 'string', required: false }
            ]
        }));
    };

    const handleRemoveProperty = (index: number) => {
        setFormData(prev => ({
            ...prev,
            properties: prev.properties.filter((_, i) => i !== index)
        }));
    };

    return (
        <div className="bg-white rounded-lg shadow p-6">
            <div className="mb-6 flex justify-between items-center">
                <h2 className="text-xl font-semibold">Relationship Manager</h2>
                <button
                    onClick={() => setShowForm(true)}
                    className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                    disabled={isLoading}
                >
                    Add New Relationship
                </button>
            </div>

            {/* Relationship List */}
            <div className="mb-8">
                <h3 className="text-lg font-medium mb-4">Existing Relationships</h3>
                <div className="space-y-4">
                    {existingRelations.map((relation, index) => (
                        <div
                            key={index}
                            className="p-4 border rounded-lg hover:bg-gray-50"
                        >
                            <div className="flex justify-between items-start">
                                <div>
                                    <div className="font-medium">
                                        {relation.sourceType} → {relation.relationType} → {relation.targetType}
                                    </div>
                                    <div className="text-sm text-gray-500">
                                        Cardinality: {relation.cardinality}
                                    </div>
                                    {relation.properties.length > 0 && (
                                        <div className="mt-2">
                                            <div className="text-sm font-medium">Properties:</div>
                                            <ul className="text-sm text-gray-600">
                                                {relation.properties.map((prop, i) => (
                                                    <li key={i}>
                                                        {prop.name} ({prop.type})
                                                        {prop.required && ' *'}
                                                    </li>
                                                ))}
                                            </ul>
                                        </div>
                                    )}
                                </div>
                                <div className="flex space-x-2">
                                    <button
                                        onClick={() => {
                                            setSelectedRelation(relation);
                                            setFormData(relation);
                                            setShowForm(true);
                                        }}
                                        className="px-3 py-1 text-blue-600 hover:bg-blue-50 rounded"
                                    >
                                        Edit
                                    </button>
                                    <button
                                        onClick={() => handleDelete(relation)}
                                        className="px-3 py-1 text-red-600 hover:bg-red-50 rounded"
                                    >
                                        Delete
                                    </button>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>

            {/* Add/Edit Form */}
            {showForm && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
                    <div className="bg-white rounded-lg p-6 max-w-2xl w-full mx-4">
                        <h3 className="text-lg font-medium mb-4">
                            {selectedRelation ? 'Edit Relationship' : 'Add New Relationship'}
                        </h3>
                        <form onSubmit={handleSubmit} className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">
                                        Source Type
                                    </label>
                                    <select
                                        value={formData.sourceType}
                                        onChange={(e) => setFormData(prev => ({
                                            ...prev,
                                            sourceType: e.target.value
                                        }))}
                                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm"
                                        required
                                    >
                                        <option value="">Select source type</option>
                                        {nodeTypes.map(type => (
                                            <option key={type.id} value={type.name}>
                                                {type.name}
                                            </option>
                                        ))}
                                    </select>
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">
                                        Target Type
                                    </label>
                                    <select
                                        value={formData.targetType}
                                        onChange={(e) => setFormData(prev => ({
                                            ...prev,
                                            targetType: e.target.value
                                        }))}
                                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm"
                                        required
                                    >
                                        <option value="">Select target type</option>
                                        {nodeTypes.map(type => (
                                            <option key={type.id} value={type.name}>
                                                {type.name}
                                            </option>
                                        ))}
                                    </select>
                                </div>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700">
                                    Relationship Type
                                </label>
                                <input
                                    type="text"
                                    value={formData.relationType}
                                    onChange={(e) => setFormData(prev => ({
                                        ...prev,
                                        relationType: e.target.value.toUpperCase()
                                    }))}
                                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm"
                                    placeholder="e.g., OWNS, MANAGES"
                                    required
                                />
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700">
                                    Cardinality
                                </label>
                                <select
                                    value={formData.cardinality}
                                    onChange={(e) => setFormData(prev => ({
                                        ...prev,
                                        cardinality: e.target.value as NodeTypeRelation['cardinality']
                                    }))}
                                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm"
                                >
                                    <option value="one-to-one">One-to-One</option>
                                    <option value="one-to-many">One-to-Many</option>
                                    <option value="many-to-many">Many-to-Many</option>
                                </select>
                            </div>

                            {/* Properties Section */}
                            <div>
                                <div className="flex justify-between items-center mb-2">
                                    <label className="block text-sm font-medium text-gray-700">
                                        Properties
                                    </label>
                                    <button
                                        type="button"
                                        onClick={handleAddProperty}
                                        className="text-sm text-blue-600 hover:text-blue-500"
                                    >
                                        Add Property
                                    </button>
                                </div>
                                {formData.properties.map((prop, index) => (
                                    <div key={index} className="flex gap-4 items-center mb-2">
                                        <input
                                            type="text"
                                            value={prop.name}
                                            onChange={(e) => {
                                                const newProps = [...formData.properties];
                                                newProps[index] = {
                                                    ...newProps[index],
                                                    name: e.target.value
                                                };
                                                setFormData(prev => ({
                                                    ...prev,
                                                    properties: newProps
                                                }));
                                            }}
                                            className="flex-1 rounded-md border-gray-300 shadow-sm"
                                            placeholder="Property name"
                                        />
                                        <select
                                            value={prop.type}
                                            onChange={(e) => {
                                                const newProps = [...formData.properties];
                                                newProps[index] = {
                                                    ...newProps[index],
                                                    type: e.target.value
                                                };
                                                setFormData(prev => ({
                                                    ...prev,
                                                    properties: newProps
                                                }));
                                            }}
                                            className="w-32 rounded-md border-gray-300 shadow-sm"
                                        >
                                            <option value="string">String</option>
                                            <option value="number">Number</option>
                                            <option value="boolean">Boolean</option>
                                            <option value="date">Date</option>
                                        </select>
                                        <label className="flex items-center">
                                            <input
                                                type="checkbox"
                                                checked={prop.required}
                                                onChange={(e) => {
                                                    const newProps = [...formData.properties];
                                                    newProps[index] = {
                                                        ...newProps[index],
                                                        required: e.target.checked
                                                    };
                                                    setFormData(prev => ({
                                                        ...prev,
                                                        properties: newProps
                                                    }));
                                                }}
                                                className="mr-2"
                                            />
                                            Required
                                        </label>
                                        <button
                                            type="button"
                                            onClick={() => handleRemoveProperty(index)}
                                            className="text-red-600 hover:text-red-500"
                                        >
                                            Remove
                                        </button>
                                    </div>
                                ))}
                            </div>

                            <div className="flex justify-end space-x-4 mt-6">
                                <button
                                    type="button"
                                    onClick={() => {
                                        setShowForm(false);
                                        setSelectedRelation(null);
                                    }}
                                    className="px-4 py-2 border text-gray-700 rounded hover:bg-gray-50"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                                    disabled={isLoading}
                                >
                                    {isLoading ? 'Processing...' : selectedRelation ? 'Update' : 'Create'}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
};

export default RelationshipManager;
