import React, { useState } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import BatchOperations from '../../components/graph/BatchOperations';
import BatchUpload from '../../components/graph/BatchUpload';
import { useNotification } from '../../hooks/useNotification';

const BatchPage: React.FC = () => {
    const [isLoading, setIsLoading] = useState(false);
    const { addNotification } = useNotification();

    const handleBatchOperation = async (operation: any) => {
        setIsLoading(true);
        try {
            const response = await fetch('/api/nodes/batch', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(operation),
            });

            const data = await response.json();

            if (!data.success) {
                throw new Error(data.error || 'Failed to execute batch operation');
            }

            addNotification(
                'success',
                `Successfully processed batch operation. Created ${data.meta.nodesCreated} nodes and ${data.meta.relationshipsCreated} relationships.`
            );

            return {
                success: data.meta.nodesCreated + data.meta.relationshipsCreated,
                failed: 0,
                errors: [],
            };
        } catch (error: any) {
            addNotification(
                'error',
                `Failed to execute batch operation: ${error.message}`
            );

            return {
                success: 0,
                failed: 1,
                errors: [{ index: 0, error: error.message }],
            };
        } finally {
            setIsLoading(false);
        }
    };

    const handleFileImport = async (files: File[], config: any) => {
        setIsLoading(true);
        try {
            const formData = new FormData();
            files.forEach(file => formData.append('files', file));
            formData.append('config', JSON.stringify(config));

            const response = await fetch('/api/import/batch', {
                method: 'POST',
                body: formData,
            });

            const data = await response.json();

            if (!data.success) {
                throw new Error(data.error || 'Failed to import files');
            }

            addNotification(
                'success',
                `Successfully imported data. Created ${data.meta.nodesCreated} nodes and ${data.meta.relationshipsCreated} relationships.`
            );

            return {
                success: data.meta.nodesCreated + data.meta.relationshipsCreated,
                failed: 0,
                errors: [],
            };
        } catch (error: any) {
            addNotification(
                'error',
                `Failed to import files: ${error.message}`
            );

            return {
                success: 0,
                failed: 1,
                errors: [{ line: 0, error: error.message }],
            };
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <MainLayout>
            <div className="space-y-8">
                <div className="bg-white shadow sm:rounded-lg">
                    <div className="px-4 py-5 sm:p-6">
                        <h2 className="text-lg font-medium leading-6 text-gray-900 mb-4">
                            Batch Operations
                        </h2>
                        <BatchOperations 
                            onExecuteBatch={handleBatchOperation}
                            isLoading={isLoading}
                        />
                    </div>
                </div>

                <div className="bg-white shadow sm:rounded-lg">
                    <div className="px-4 py-5 sm:p-6">
                        <h2 className="text-lg font-medium leading-6 text-gray-900 mb-4">
                            Batch Import
                        </h2>
                        <BatchUpload
                            onImport={handleFileImport}
                            isLoading={isLoading}
                        />
                    </div>
                </div>
            </div>
        </MainLayout>
    );
};

export default BatchPage;
