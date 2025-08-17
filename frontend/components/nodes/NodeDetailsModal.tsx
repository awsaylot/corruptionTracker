import { useState } from 'react';
import Link from 'next/link';
import { Node } from '../../utils/api';

interface NodeDetailsModalProps {
    node: Node;
    onClose: () => void;
    onDelete: () => Promise<void>;
    onUpdate: (data: Partial<Node>) => Promise<void>;
}

interface EditableFields {
    name?: string;
    title?: string;
    description?: string;
    type?: string;
    aliases?: string;
    date?: string;
    sources?: string;
    notes?: string;
}

export const NodeDetailsModal: React.FC<NodeDetailsModalProps> = ({
    node,
    onClose,
    onDelete,
    onUpdate
}) => {
    const [isEditing, setIsEditing] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [editedData, setEditedData] = useState<EditableFields>({});

    const startEdit = () => {
        setEditedData({
            name: node.properties.name,
            title: node.properties.title,
            description: node.properties.description,
            type: node.properties.type,
            aliases: Array.isArray(node.properties.aliases) 
                ? node.properties.aliases.join(', ')
                : node.properties.aliases,
            date: node.properties.date,
            sources: Array.isArray(node.properties.sources)
                ? node.properties.sources.join('\n')
                : node.properties.sources,
            notes: node.properties.notes
        });
        setIsEditing(true);
    };

    const handleUpdate = async () => {
        setLoading(true);
        setError(null);
        try {
            const updatedData: Record<string, any> = {};
            
            // Process arrays
            if (editedData.aliases) {
                updatedData.aliases = editedData.aliases.split(',').map(s => s.trim()).filter(Boolean);
            }
            if (editedData.sources) {
                updatedData.sources = editedData.sources.split('\n').map(s => s.trim()).filter(Boolean);
            }

            // Add other fields
            Object.entries(editedData).forEach(([key, value]) => {
                if (value !== undefined && !['aliases', 'sources'].includes(key)) {
                    updatedData[key] = value;
                }
            });

            await onUpdate(updatedData);
            setIsEditing(false);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to update node');
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async () => {
        if (!isDeleting) {
            setIsDeleting(true);
            return;
        }

        setLoading(true);
        setError(null);
        try {
            await onDelete();
            onClose();
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to delete node');
            setIsDeleting(false);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4">
            <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
                <div className="p-6 space-y-4">
                    <div className="flex justify-between items-start">
                        <h2 className="text-xl font-semibold text-gray-900">
                            {isEditing ? 'Edit Node' : node.type}
                        </h2>
                        <button
                            onClick={onClose}
                            className="text-gray-400 hover:text-gray-500"
                        >
                            <span className="sr-only">Close</span>
                            <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                            </svg>
                        </button>
                    </div>

                    {error && (
                        <div className="text-red-600 text-sm">
                            {error}
                        </div>
                    )}

                    {isEditing ? (
                        <div className="space-y-4">
                            {node.type !== 'Event' && (
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">Name</label>
                                    <input
                                        type="text"
                                        value={editedData.name || ''}
                                        onChange={(e) => setEditedData(prev => ({ ...prev, name: e.target.value }))}
                                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                                    />
                                </div>
                            )}

                            {node.type === 'Event' && (
                                <>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700">Title</label>
                                        <input
                                            type="text"
                                            value={editedData.title || ''}
                                            onChange={(e) => setEditedData(prev => ({ ...prev, title: e.target.value }))}
                                            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700">Description</label>
                                        <textarea
                                            value={editedData.description || ''}
                                            onChange={(e) => setEditedData(prev => ({ ...prev, description: e.target.value }))}
                                            rows={3}
                                            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700">Date</label>
                                        <input
                                            type="date"
                                            value={editedData.date || ''}
                                            onChange={(e) => setEditedData(prev => ({ ...prev, date: e.target.value }))}
                                            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                                        />
                                    </div>
                                </>
                            )}

                            {node.type === 'Organization' && (
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">Type</label>
                                    <select
                                        value={editedData.type || ''}
                                        onChange={(e) => setEditedData(prev => ({ ...prev, type: e.target.value }))}
                                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                                    >
                                        <option value="">Select Type...</option>
                                        <option value="Government">Government</option>
                                        <option value="Corporation">Corporation</option>
                                        <option value="NGO">NGO</option>
                                        <option value="Other">Other</option>
                                    </select>
                                </div>
                            )}

                            {(node.type === 'Person' || node.type === 'Organization') && (
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">Aliases (comma-separated)</label>
                                    <input
                                        type="text"
                                        value={editedData.aliases || ''}
                                        onChange={(e) => setEditedData(prev => ({ ...prev, aliases: e.target.value }))}
                                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                                        placeholder="e.g., John Doe, JD"
                                    />
                                </div>
                            )}

                            {node.type === 'Event' && (
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">Sources (one per line)</label>
                                    <textarea
                                        value={editedData.sources || ''}
                                        onChange={(e) => setEditedData(prev => ({ ...prev, sources: e.target.value }))}
                                        rows={3}
                                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                                        placeholder="https://example.com"
                                    />
                                </div>
                            )}

                            <div>
                                <label className="block text-sm font-medium text-gray-700">Notes</label>
                                <textarea
                                    value={editedData.notes || ''}
                                    onChange={(e) => setEditedData(prev => ({ ...prev, notes: e.target.value }))}
                                    rows={3}
                                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                                />
                            </div>
                        </div>
                    ) : (
                        <div className="space-y-4">
                            {node.type !== 'Event' ? (
                                <div>
                                    <h3 className="text-sm font-medium text-gray-700">Name</h3>
                                    <p className="mt-1 text-sm text-gray-900">{node.properties.name}</p>
                                </div>
                            ) : (
                                <>
                                    <div>
                                        <h3 className="text-sm font-medium text-gray-700">Title</h3>
                                        <p className="mt-1 text-sm text-gray-900">{node.properties.title}</p>
                                    </div>
                                    {node.properties.description && (
                                        <div>
                                            <h3 className="text-sm font-medium text-gray-700">Description</h3>
                                            <p className="mt-1 text-sm text-gray-900">{node.properties.description}</p>
                                        </div>
                                    )}
                                    {node.properties.date && (
                                        <div>
                                            <h3 className="text-sm font-medium text-gray-700">Date</h3>
                                            <p className="mt-1 text-sm text-gray-900">{node.properties.date}</p>
                                        </div>
                                    )}
                                </>
                            )}

                            {node.type === 'Organization' && node.properties.type && (
                                <div>
                                    <h3 className="text-sm font-medium text-gray-700">Type</h3>
                                    <p className="mt-1 text-sm text-gray-900">{node.properties.type}</p>
                                </div>
                            )}

                            {node.properties.aliases && node.properties.aliases.length > 0 && (
                                <div>
                                    <h3 className="text-sm font-medium text-gray-700">Aliases</h3>
                                    <p className="mt-1 text-sm text-gray-900">
                                        {Array.isArray(node.properties.aliases) 
                                            ? node.properties.aliases.join(', ')
                                            : node.properties.aliases}
                                    </p>
                                </div>
                            )}

                            {node.properties.sources && node.properties.sources.length > 0 && (
                                <div>
                                    <h3 className="text-sm font-medium text-gray-700">Sources</h3>
                                    <ul className="mt-1 text-sm text-blue-600 space-y-1">
                                        {(Array.isArray(node.properties.sources) ? node.properties.sources : [node.properties.sources]).map((source, index) => (
                                            <li key={index}>
                                                <a href={source} target="_blank" rel="noopener noreferrer" className="hover:underline">
                                                    {source}
                                                </a>
                                            </li>
                                        ))}
                                    </ul>
                                </div>
                            )}

                            {node.properties.notes && (
                                <div>
                                    <h3 className="text-sm font-medium text-gray-700">Notes</h3>
                                    <p className="mt-1 text-sm text-gray-900">{node.properties.notes}</p>
                                </div>
                            )}
                        </div>
                    )}

                    <div className="mt-6 flex justify-between space-x-3">
                        {isEditing ? (
                            <>
                                <button
                                    onClick={() => setIsEditing(false)}
                                    className="inline-flex justify-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                                >
                                    Cancel
                                </button>
                                <button
                                    onClick={handleUpdate}
                                    disabled={loading}
                                    className="inline-flex justify-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
                                >
                                    {loading ? 'Saving...' : 'Save Changes'}
                                </button>
                            </>
                        ) : (
                            <>
                                <Link
                                    href={`/nodes/${node.id}`}
                                    className="inline-flex justify-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                                >
                                    View Details
                                </Link>
                                <button
                                    onClick={startEdit}
                                    className="ml-3 inline-flex justify-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                                >
                                    Edit
                                </button>
                                <button
                                    onClick={handleDelete}
                                    disabled={loading}
                                    className={`ml-3 inline-flex justify-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 ${
                                        isDeleting
                                            ? 'bg-red-600 hover:bg-red-700'
                                            : 'bg-gray-600 hover:bg-gray-700'
                                    }`}
                                >
                                    {loading ? 'Deleting...' : isDeleting ? 'Click again to confirm' : 'Delete'}
                                </button>
                            </>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default NodeDetailsModal;
