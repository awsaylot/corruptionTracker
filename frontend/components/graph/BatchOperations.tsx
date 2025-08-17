import React, { useState, useCallback } from 'react';
import { useNotification } from '../../hooks/useNotification';

interface BatchOperation {
    type: 'create' | 'update' | 'delete' | 'merge' | 'import';
    target: 'nodes' | 'relationships';
    data: any[];
    options?: {
        skipDuplicates?: boolean;
        updateExisting?: boolean;
        validateRelationships?: boolean;
        batchSize?: number;
    };
}

interface BatchOperationsProps {
    onExecuteBatch: (operation: BatchOperation) => Promise<{
        success: number;
        failed: number;
        errors: Array<{ index: number; error: string }>;
    }>;
    isLoading?: boolean;
}

const BatchOperations: React.FC<BatchOperationsProps> = ({
    onExecuteBatch,
    isLoading
}) => {
    const { addNotification } = useNotification();
    const [operation, setOperation] = useState<BatchOperation>({
        type: 'create',
        target: 'nodes',
        data: [],
        options: {
            skipDuplicates: true,
            updateExisting: false,
            validateRelationships: true,
            batchSize: 1000
        }
    });
    const [jsonInput, setJsonInput] = useState('');
    const [results, setResults] = useState<{
        success: number;
        failed: number;
        errors: Array<{ index: number; error: string }>;
    } | null>(null);

    const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        setJsonInput(e.target.value);
        try {
            const parsed = JSON.parse(e.target.value);
            setOperation(prev => ({
                ...prev,
                data: Array.isArray(parsed) ? parsed : [parsed]
            }));
        } catch (error) {
            // Don't update data if JSON is invalid
        }
    };

    const validateInput = (): boolean => {
        try {
            if (!jsonInput.trim()) {
                addNotification('error', 'Please enter JSON data');
                return false;
            }

            const parsed = JSON.parse(jsonInput);
            if (!Array.isArray(parsed) && typeof parsed !== 'object') {
                addNotification('error', 'Input must be a JSON array or object');
                return false;
            }

            return true;
        } catch (error) {
            addNotification('error', 'Invalid JSON format');
            return false;
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!validateInput()) return;

        try {
            const result = await onExecuteBatch(operation);
            setResults(result);
            addNotification(
                result.failed === 0 ? 'success' : 'warning',
                `Batch operation completed: ${result.success} successful, ${result.failed} failed`
            );
        } catch (error) {
            addNotification('error', error instanceof Error ? error.message : 'Batch operation failed');
        }
    };

    const handleFileUpload = useCallback((files: FileList | null) => {
        if (!files || files.length === 0) return;

        const file = files[0];
        const reader = new FileReader();

        reader.onload = (e) => {
            try {
                const content = e.target?.result as string;
                setJsonInput(content);
                const parsed = JSON.parse(content);
                setOperation(prev => ({
                    ...prev,
                    data: Array.isArray(parsed) ? parsed : [parsed]
                }));
                addNotification('success', 'File loaded successfully');
            } catch (error) {
                addNotification('error', 'Failed to parse JSON file');
            }
        };

        reader.onerror = () => {
            addNotification('error', 'Failed to read file');
        };

        reader.readAsText(file);
    }, [addNotification]);

    return (
        <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-6">Batch Operations</h2>

            <form onSubmit={handleSubmit} className="space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Operation Type
                        </label>
                        <select
                            value={operation.type}
                            onChange={(e) => setOperation(prev => ({
                                ...prev,
                                type: e.target.value as BatchOperation['type']
                            }))}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md"
                            disabled={isLoading}
                        >
                            <option value="create">Create</option>
                            <option value="update">Update</option>
                            <option value="delete">Delete</option>
                            <option value="merge">Merge</option>
                            <option value="import">Import</option>
                        </select>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Target
                        </label>
                        <select
                            value={operation.target}
                            onChange={(e) => setOperation(prev => ({
                                ...prev,
                                target: e.target.value as BatchOperation['target']
                            }))}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md"
                            disabled={isLoading}
                        >
                            <option value="nodes">Nodes</option>
                            <option value="relationships">Relationships</option>
                        </select>
                    </div>
                </div>

                <div className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Options
                        </label>
                        <div className="space-y-2">
                            <label className="flex items-center">
                                <input
                                    type="checkbox"
                                    checked={operation.options?.skipDuplicates}
                                    onChange={(e) => setOperation(prev => ({
                                        ...prev,
                                        options: {
                                            ...prev.options,
                                            skipDuplicates: e.target.checked
                                        }
                                    }))}
                                    className="mr-2"
                                    disabled={isLoading}
                                />
                                Skip Duplicates
                            </label>
                            <label className="flex items-center">
                                <input
                                    type="checkbox"
                                    checked={operation.options?.updateExisting}
                                    onChange={(e) => setOperation(prev => ({
                                        ...prev,
                                        options: {
                                            ...prev.options,
                                            updateExisting: e.target.checked
                                        }
                                    }))}
                                    className="mr-2"
                                    disabled={isLoading}
                                />
                                Update Existing
                            </label>
                            <label className="flex items-center">
                                <input
                                    type="checkbox"
                                    checked={operation.options?.validateRelationships}
                                    onChange={(e) => setOperation(prev => ({
                                        ...prev,
                                        options: {
                                            ...prev.options,
                                            validateRelationships: e.target.checked
                                        }
                                    }))}
                                    className="mr-2"
                                    disabled={isLoading}
                                />
                                Validate Relationships
                            </label>
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Batch Size
                        </label>
                        <input
                            type="number"
                            value={operation.options?.batchSize}
                            onChange={(e) => setOperation(prev => ({
                                ...prev,
                                options: {
                                    ...prev.options,
                                    batchSize: parseInt(e.target.value) || 1000
                                }
                            }))}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md"
                            min="1"
                            max="10000"
                            disabled={isLoading}
                        />
                    </div>
                </div>

                <div>
                    <div className="flex justify-between items-center mb-2">
                        <label className="block text-sm font-medium text-gray-700">
                            JSON Data
                        </label>
                        <div>
                            <input
                                type="file"
                                accept=".json"
                                onChange={(e) => handleFileUpload(e.target.files)}
                                className="hidden"
                                id="json-upload"
                                disabled={isLoading}
                            />
                            <label
                                htmlFor="json-upload"
                                className="cursor-pointer text-sm text-blue-600 hover:text-blue-500"
                            >
                                Upload JSON File
                            </label>
                        </div>
                    </div>
                    <textarea
                        value={jsonInput}
                        onChange={handleInputChange}
                        className="w-full h-64 px-3 py-2 border border-gray-300 rounded-md font-mono text-sm"
                        placeholder="Enter JSON data here..."
                        disabled={isLoading}
                    />
                </div>

                {results && (
                    <div className="bg-gray-50 p-4 rounded-md">
                        <h3 className="text-lg font-medium mb-2">Results</h3>
                        <div className="space-y-2">
                            <div className="text-green-600">
                                Successful operations: {results.success}
                            </div>
                            <div className="text-red-600">
                                Failed operations: {results.failed}
                            </div>
                            {results.errors.length > 0 && (
                                <div>
                                    <h4 className="font-medium mb-1">Errors:</h4>
                                    <ul className="text-sm text-red-600 list-disc list-inside">
                                        {results.errors.map((error, index) => (
                                            <li key={index}>
                                                Row {error.index}: {error.error}
                                            </li>
                                        ))}
                                    </ul>
                                </div>
                            )}
                        </div>
                    </div>
                )}

                <div className="flex justify-end">
                    <button
                        type="submit"
                        className={`px-4 py-2 rounded-md text-white font-medium ${
                            isLoading
                                ? 'bg-gray-400 cursor-not-allowed'
                                : 'bg-blue-600 hover:bg-blue-700'
                        }`}
                        disabled={isLoading}
                    >
                        {isLoading ? 'Processing...' : 'Execute Batch Operation'}
                    </button>
                </div>
            </form>
        </div>
    );
};

export default BatchOperations;
