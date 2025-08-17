import React, { useState, useCallback } from 'react';
import { useNotification } from '../../hooks/useNotification';

interface ImportConfig {
    type: 'csv' | 'json' | 'cypher';
    target: 'nodes' | 'relationships';
    mapping?: {
        [key: string]: string;
    };
    options: {
        delimiter?: string;
        hasHeader?: boolean;
        skipRows?: number;
        batchSize?: number;
        validateData?: boolean;
        updateExisting?: boolean;
    };
}

interface BatchUploadProps {
    onImport: (files: File[], config: ImportConfig) => Promise<{
        success: number;
        failed: number;
        errors: Array<{ line: number; error: string }>;
    }>;
    isLoading?: boolean;
}

const BatchUpload: React.FC<BatchUploadProps> = ({
    onImport,
    isLoading
}) => {
    const { addNotification } = useNotification();
    const [files, setFiles] = useState<File[]>([]);
    const [config, setConfig] = useState<ImportConfig>({
        type: 'csv',
        target: 'nodes',
        mapping: {},
        options: {
            delimiter: ',',
            hasHeader: true,
            skipRows: 0,
            batchSize: 1000,
            validateData: true,
            updateExisting: false
        }
    });
    const [previewData, setPreviewData] = useState<string[][]>([]);
    const [columnMapping, setColumnMapping] = useState<{ [key: string]: string }>({});

    const handleFileSelect = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
        const selectedFiles = Array.from(e.target.files || []);
        setFiles(selectedFiles);

        // Preview first file
        if (selectedFiles.length > 0) {
            const reader = new FileReader();
            reader.onload = () => {
                const content = reader.result as string;
                if (config.type === 'csv') {
                    const lines = content.split('\n');
                    const delimiter = config.options.delimiter || ',';
                    const preview = lines
                        .slice(0, 5)
                        .map(line => line.split(delimiter));
                    setPreviewData(preview);

                    // Initialize column mapping if headers exist
                    if (config.options.hasHeader && preview.length > 0) {
                        const headers = preview[0];
                        const initialMapping = headers.reduce((acc, header) => ({
                            ...acc,
                            [header.trim()]: ''
                        }), {});
                        setColumnMapping(initialMapping);
                    }
                }
            };
            reader.readAsText(selectedFiles[0]);
        }
    }, [config.type, config.options.delimiter, config.options.hasHeader]);

    const handleImport = async () => {
        if (files.length === 0) {
            addNotification('error', 'Please select files to import');
            return;
        }

        try {
            const result = await onImport(files, {
                ...config,
                mapping: columnMapping
            });

            addNotification(
                result.failed === 0 ? 'success' : 'warning',
                `Import completed: ${result.success} successful, ${result.failed} failed`
            );

            if (result.errors.length > 0) {
                console.error('Import errors:', result.errors);
            }
        } catch (error) {
            addNotification('error', error instanceof Error ? error.message : 'Import failed');
        }
    };

    return (
        <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-6">Batch Upload</h2>

            <div className="space-y-6">
                {/* File Selection */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Upload Files
                    </label>
                    <input
                        type="file"
                        onChange={handleFileSelect}
                        multiple
                        accept=".csv,.json,.cypher"
                        className="block w-full text-sm text-gray-500
                            file:mr-4 file:py-2 file:px-4
                            file:rounded-md file:border-0
                            file:text-sm file:font-semibold
                            file:bg-blue-50 file:text-blue-700
                            hover:file:bg-blue-100"
                        disabled={isLoading}
                    />
                    {files.length > 0 && (
                        <div className="mt-2 text-sm text-gray-500">
                            Selected files: {files.map(f => f.name).join(', ')}
                        </div>
                    )}
                </div>

                {/* Import Configuration */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            File Type
                        </label>
                        <select
                            value={config.type}
                            onChange={(e) => setConfig(prev => ({
                                ...prev,
                                type: e.target.value as ImportConfig['type']
                            }))}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md"
                            disabled={isLoading}
                        >
                            <option value="csv">CSV</option>
                            <option value="json">JSON</option>
                            <option value="cypher">Cypher Script</option>
                        </select>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Target
                        </label>
                        <select
                            value={config.target}
                            onChange={(e) => setConfig(prev => ({
                                ...prev,
                                target: e.target.value as ImportConfig['target']
                            }))}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md"
                            disabled={isLoading}
                        >
                            <option value="nodes">Nodes</option>
                            <option value="relationships">Relationships</option>
                        </select>
                    </div>
                </div>

                {/* CSV-specific options */}
                {config.type === 'csv' && (
                    <div className="space-y-4">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-2">
                                    Delimiter
                                </label>
                                <input
                                    type="text"
                                    value={config.options.delimiter}
                                    onChange={(e) => setConfig(prev => ({
                                        ...prev,
                                        options: {
                                            ...prev.options,
                                            delimiter: e.target.value
                                        }
                                    }))}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                                    maxLength={1}
                                    disabled={isLoading}
                                />
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-2">
                                    Skip Rows
                                </label>
                                <input
                                    type="number"
                                    value={config.options.skipRows}
                                    onChange={(e) => setConfig(prev => ({
                                        ...prev,
                                        options: {
                                            ...prev.options,
                                            skipRows: parseInt(e.target.value) || 0
                                        }
                                    }))}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                                    min="0"
                                    disabled={isLoading}
                                />
                            </div>
                        </div>

                        <div className="flex items-center space-x-4">
                            <label className="flex items-center">
                                <input
                                    type="checkbox"
                                    checked={config.options.hasHeader}
                                    onChange={(e) => setConfig(prev => ({
                                        ...prev,
                                        options: {
                                            ...prev.options,
                                            hasHeader: e.target.checked
                                        }
                                    }))}
                                    className="mr-2"
                                    disabled={isLoading}
                                />
                                Has Header Row
                            </label>
                        </div>

                        {/* Column Mapping */}
                        {config.options.hasHeader && previewData.length > 0 && (
                            <div>
                                <h3 className="text-sm font-medium text-gray-700 mb-2">
                                    Column Mapping
                                </h3>
                                <div className="space-y-2">
                                    {Object.keys(columnMapping).map((column) => (
                                        <div key={column} className="flex items-center space-x-2">
                                            <span className="text-sm text-gray-600 w-1/3">
                                                {column}:
                                            </span>
                                            <input
                                                type="text"
                                                value={columnMapping[column]}
                                                onChange={(e) => setColumnMapping(prev => ({
                                                    ...prev,
                                                    [column]: e.target.value
                                                }))}
                                                className="flex-1 px-2 py-1 border border-gray-300 rounded-md text-sm"
                                                placeholder="Property name"
                                                disabled={isLoading}
                                            />
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}
                    </div>
                )}

                {/* Preview */}
                {previewData.length > 0 && (
                    <div>
                        <h3 className="text-sm font-medium text-gray-700 mb-2">
                            Preview (first 5 rows)
                        </h3>
                        <div className="overflow-x-auto">
                            <table className="min-w-full divide-y divide-gray-200">
                                <tbody className="divide-y divide-gray-200">
                                    {previewData.map((row, i) => (
                                        <tr key={i}>
                                            {row.map((cell, j) => (
                                                <td
                                                    key={j}
                                                    className={`px-3 py-2 text-sm ${
                                                        i === 0 && config.options.hasHeader
                                                            ? 'font-medium bg-gray-50'
                                                            : ''
                                                    }`}
                                                >
                                                    {cell}
                                                </td>
                                            ))}
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                )}

                {/* Import Options */}
                <div className="space-y-2">
                    <label className="flex items-center">
                        <input
                            type="checkbox"
                            checked={config.options.validateData}
                            onChange={(e) => setConfig(prev => ({
                                ...prev,
                                options: {
                                    ...prev.options,
                                    validateData: e.target.checked
                                }
                            }))}
                            className="mr-2"
                            disabled={isLoading}
                        />
                        Validate Data Before Import
                    </label>
                    <label className="flex items-center">
                        <input
                            type="checkbox"
                            checked={config.options.updateExisting}
                            onChange={(e) => setConfig(prev => ({
                                ...prev,
                                options: {
                                    ...prev.options,
                                    updateExisting: e.target.checked
                                }
                            }))}
                            className="mr-2"
                            disabled={isLoading}
                        />
                        Update Existing Entries
                    </label>
                </div>

                {/* Import Button */}
                <div className="flex justify-end">
                    <button
                        onClick={handleImport}
                        className={`px-4 py-2 rounded-md text-white font-medium ${
                            isLoading || files.length === 0
                                ? 'bg-gray-400 cursor-not-allowed'
                                : 'bg-blue-600 hover:bg-blue-700'
                        }`}
                        disabled={isLoading || files.length === 0}
                    >
                        {isLoading ? 'Importing...' : 'Start Import'}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default BatchUpload;
