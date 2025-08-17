// pages/nodes/create.tsx - Updated with full integration
import { NextPage } from 'next';
import Head from 'next/head';
import { useRouter } from 'next/router';
import Link from 'next/link';
import NodeForm from '../../components/nodes/NodeForm';
import NotificationContainer from '../../components/ui/NotificationContainer';
import { useNotification } from '../../hooks/useNotification';

const CreateNodePage: NextPage = () => {
    const router = useRouter();
    const { notifications, addNotification, removeNotification } = useNotification();

    const handleNodeCreated = (nodeId: string) => {
        addNotification('success', 'Node created successfully!');
        
        // Redirect to the new node's detail page after a short delay
        setTimeout(() => {
            router.push(`/nodes/${nodeId}`);
        }, 1500);
    };

    const handleError = (error: string) => {
        addNotification('error', error);
    };

    const handleSuccess = (message: string) => {
        addNotification('success', message);
    };

    return (
        <>
            <Head>
                <title>Create Node - GraphDB</title>
                <meta name="description" content="Create a new node in the graph database" />
            </Head>

            <NotificationContainer 
                notifications={notifications} 
                onRemove={removeNotification} 
            />

            <div className="max-w-2xl mx-auto">
                {/* Header */}
                <div className="mb-8">
                    <div className="flex items-center justify-between">
                        <div>
                            <h1 className="text-2xl font-bold leading-7 text-gray-900">Create New Node</h1>
                            <p className="mt-1 text-sm text-gray-600">
                                Add a new person, organization, or event to your graph database.
                            </p>
                        </div>
                        <Link
                            href="/nodes"
                            className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                        >
                            <svg className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                            </svg>
                            Back to Nodes
                        </Link>
                    </div>
                </div>

                {/* Node Creation Form */}
                <NodeForm
                    onComplete={handleNodeCreated}
                    onError={handleError}
                    onSuccess={handleSuccess}
                    redirectOnSuccess={true}
                />

                {/* Helper Information */}
                <div className="mt-8 bg-blue-50 border-l-4 border-blue-400 p-4">
                    <div className="flex">
                        <div className="flex-shrink-0">
                            <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
                                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                            </svg>
                        </div>
                        <div className="ml-3">
                            <h3 className="text-sm font-medium text-blue-800">Tips for Creating Nodes</h3>
                            <div className="mt-2 text-sm text-blue-700">
                                <ul className="list-disc list-inside space-y-1">
                                    <li><strong>People:</strong> Use full names and include common aliases</li>
                                    <li><strong>Organizations:</strong> Specify the organization type for better categorization</li>
                                    <li><strong>Events:</strong> Include dates and source documentation when available</li>
                                    <li><strong>Notes:</strong> Add contextual information that might be useful for analysis</li>
                                </ul>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </>
    );
};

export default CreateNodePage;