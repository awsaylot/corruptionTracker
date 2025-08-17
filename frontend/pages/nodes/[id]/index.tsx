// pages/nodes/[id]/index.tsx - Updated with defensive null checks
import { useState, useEffect } from 'react'
import { useRouter } from 'next/router'
import Head from 'next/head'
import Link from 'next/link'
import { NodeWithConnections, api } from '../../../utils/api'
import RelationshipForm from '../../../components/relationships/RelationshipForm'
import NotificationContainer from '../../../components/ui/NotificationContainer'
import LoadingSpinner from '../../../components/ui/LoadingSpinner'
import { useNotification } from '../../../hooks/useNotification'

export default function NodeDetailsPage() {
  const router = useRouter()
  const { id } = router.query
  const { notifications, addNotification, removeNotification } = useNotification()
  
  // State management
  const [node, setNode] = useState<NodeWithConnections | null>(null)
  const [showRelationshipForm, setShowRelationshipForm] = useState(false)
  
  // Loading states
  const [isLoading, setIsLoading] = useState(true)
  const [isUpdating, setIsUpdating] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  
  // Error state
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (id && typeof id === 'string') {
      fetchNodeDetails(id)
    }
  }, [id])

  const fetchNodeDetails = async (nodeId: string) => {
    try {
      setIsLoading(true)
      setError(null)
      
      const data = await api.getNode(nodeId)
      setNode(data)
      
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch node details'
      setError(errorMessage)
      addNotification('error', errorMessage)
    } finally {
      setIsLoading(false)
    }
  }

  const handleNodeUpdate = async (properties: any) => {
    if (!node) return
    
    try {
      setIsUpdating(true)
      setError(null)
      
      await api.updateNode(node.id, properties)
      
      // Update local state
      setNode(prev => prev ? {
        ...prev,
        properties: { ...prev.properties, ...properties }
      } : null)
      
      addNotification('success', 'Node updated successfully')
      
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to update node'
      setError(errorMessage)
      addNotification('error', errorMessage)
      throw err // Re-throw for form handling
    } finally {
      setIsUpdating(false)
    }
  }

  const handleNodeDelete = async () => {
    if (!node) return
    
    const displayName = node.properties?.name ?? node.properties?.title ?? node.id
    if (!confirm(`Are you sure you want to delete "${displayName}"? This action cannot be undone.`)) {
      return
    }

    try {
      setIsDeleting(true)
      setError(null)
      
      await api.deleteNode(node.id)
      
      addNotification('success', 'Node deleted successfully')
      
      // Redirect to nodes list after successful deletion
      setTimeout(() => {
        router.push('/nodes')
      }, 1500)
      
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to delete node'
      setError(errorMessage)
      addNotification('error', errorMessage)
    } finally {
      setIsDeleting(false)
    }
  }

  const handleAddRelationship = () => {
    setShowRelationshipForm(true)
  }

  const handleRelationshipComplete = () => {
    setShowRelationshipForm(false)
    addNotification('success', 'Relationship created successfully')
    
    // Refresh node data to show new relationship
    if (node) {
      fetchNodeDetails(node.id)
    }
  }

  const handleRelationshipError = (error: string) => {
    addNotification('error', `Failed to create relationship: ${error}`)
  }

  const handleRetry = () => {
    if (id && typeof id === 'string') {
      fetchNodeDetails(id)
    }
  }

  const navigateToGraph = () => {
    router.push(`/?select=${node?.id}`)
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="min-h-screen py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-center py-12">
            <LoadingSpinner size="large" text="Loading node details..." />
          </div>
        </div>
      </div>
    )
  }

  // Error state
  if (error && !node) {
    return (
      <div className="min-h-screen py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="bg-red-50 border-l-4 border-red-400 p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">Error Loading Node</h3>
                <div className="mt-2 text-sm text-red-700">
                  <p>{error}</p>
                </div>
                <div className="mt-4">
                  <button
                    onClick={handleRetry}
                    className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                  >
                    Try Again
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (!node) return null

  // Use safe defaults for properties and connections
  const nodeProps = node.properties || {}
  const connections = node.connections || []

  return (
    <>
      <Head>
        <title>{nodeProps.name || nodeProps.title || node.id} - Node Details</title>
        <meta name="description" content={`Details for ${node.type} node ${node.id}`} />
      </Head>

      <NotificationContainer 
        notifications={notifications} 
        onRemove={removeNotification} 
      />

      <div className="py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="md:flex md:items-center md:justify-between mb-8">
            <div className="flex-1 min-w-0">
              <div className="flex items-center space-x-3">
                <div className={`w-4 h-4 rounded-full ${
                  node.type === 'Person' ? 'bg-green-500' :
                  node.type === 'Organization' ? 'bg-blue-500' :
                  'bg-yellow-500'
                }`} />
                <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
                  {nodeProps.name || nodeProps.title || `Node ${node.id}`}
                </h1>
                {(isUpdating || isDeleting) && (
                  <LoadingSpinner size="small" />
                )}
              </div>
              
              <div className="mt-1 flex flex-col sm:flex-row sm:flex-wrap sm:mt-0 sm:space-x-6">
                <div className="mt-2 flex items-center text-sm text-gray-500">
                  <svg className="flex-shrink-0 mr-1.5 h-5 w-5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M17.707 9.293a1 1 0 010 1.414l-7 7a1 1 0 01-1.414 0l-7-7A.997.997 0 012 10V5a3 3 0 013-3h5c.256 0 .512.098.707.293l7 7zM5 6a1 1 0 100-2 1 1 0 000 2z" clipRule="evenodd" />
                  </svg>
                  {node.type}
                </div>
                <div className="mt-2 flex items-center text-sm text-gray-500">
                  <svg className="flex-shrink-0 mr-1.5 h-5 w-5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M9 6a3 3 0 11-6 0 3 3 0 016 0zM17 6a3 3 0 11-6 0 3 3 0 016 0zM12.93 17c.046-.327.07-.66.07-1a6.97 6.97 0 00-1.5-4.33A5 5 0 0119 16v1h-6.07zM6 11a5 5 0 015 5v1H1v-1a5 5 0 015-5z" />
                  </svg>
                  {connections.length} connection{connections.length !== 1 ? 's' : ''}
                </div>
                <div className="mt-2 flex items-center text-sm text-gray-500">
                  <svg className="flex-shrink-0 mr-1.5 h-5 w-5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clipRule="evenodd" />
                  </svg>
                  ID: {node.id}
                </div>
              </div>
            </div>
            
            <div className="mt-4 flex flex-wrap gap-3 md:mt-0 md:ml-4">
              <Link
                href="/nodes"
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                <svg className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                </svg>
                Back to List
              </Link>
              
              <button
                type="button"
                onClick={navigateToGraph}
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                <svg className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
                View in Graph
              </button>
              
              <button
                type="button"
                onClick={handleAddRelationship}
                disabled={isUpdating || isDeleting}
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
              >
                <svg className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                Add Relationship
              </button>
              
              <button
                type="button"
                onClick={handleNodeDelete}
                disabled={isUpdating || isDeleting}
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50"
              >
                {isDeleting ? (
                  <>
                    <LoadingSpinner size="small" color="white" />
                    <span className="ml-2">Deleting...</span>
                  </>
                ) : (
                  <>
                    <svg className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                    Delete Node
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Error Display */}
          {error && (
            <div className="mb-6 bg-red-50 border-l-4 border-red-400 p-4">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3 flex-1">
                  <p className="text-sm text-red-700">{error}</p>
                </div>
                <button
                  onClick={() => setError(null)}
                  className="ml-3 text-red-400 hover:text-red-500"
                >
                  <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
          )}

          {/* Content Grid */}
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            {/* Properties Card */}
            <div className="bg-white shadow-sm overflow-hidden sm:rounded-lg">
              <div className="px-4 py-5 sm:px-6 border-b border-gray-200">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">Properties</h3>
                  {isUpdating && (
                    <div className="flex items-center space-x-2 text-sm text-gray-600">
                      <LoadingSpinner size="small" />
                      <span>Updating...</span>
                    </div>
                  )}
                </div>
              </div>
              <div className="border-t border-gray-200">
                <dl>
                  {Object.entries(nodeProps).map(([key, value], idx) => (
                    <div key={key} className={`${idx % 2 === 0 ? 'bg-gray-50' : 'bg-white'} px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6`}>
                      <dt className="text-sm font-medium text-gray-500 capitalize">
                        {key.replace(/([A-Z])/g, ' $1').replace(/^./, str => str.toUpperCase())}
                      </dt>
                      <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                        {Array.isArray(value) ? (
                          <div className="flex flex-wrap gap-1">
                            {value.map((item, index) => (
                              <span key={index} className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                                {String(item)}
                              </span>
                            ))}
                          </div>
                        ) : key === 'sources' && typeof value === 'string' ? (
                          <a href={value} target="_blank" rel="noopener noreferrer" className="text-indigo-600 hover:text-indigo-900 break-all">
                            {value}
                          </a>
                        ) : (
                          <span className="break-words">{String(value)}</span>
                        )}
                      </dd>
                    </div>
                  ))}
                </dl>
              </div>
            </div>

            {/* Relationships Card */}
            <div className="bg-white shadow-sm overflow-hidden sm:rounded-lg">
              <div className="px-4 py-5 sm:px-6 border-b border-gray-200">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">
                    Relationships ({connections.length})
                  </h3>
                  <button
                    onClick={handleAddRelationship}
                    disabled={isUpdating || isDeleting}
                    className="inline-flex items-center px-3 py-1 border border-transparent text-xs font-medium rounded text-indigo-700 bg-indigo-100 hover:bg-indigo-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
                  >
                    Add New
                  </button>
                </div>
              </div>
              
              <div className="border-t border-gray-200">
                {connections.length === 0 ? (
                  <div className="px-4 py-8 text-center">
                    <svg className="mx-auto h-8 w-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                    </svg>
                    <p className="mt-2 text-sm text-gray-500">No relationships found.</p>
                    <button
                      onClick={handleAddRelationship}
                      className="mt-3 text-indigo-600 hover:text-indigo-900 text-sm font-medium"
                    >
                      Create the first relationship
                    </button>
                  </div>
                ) : (
                  <ul className="divide-y divide-gray-200">
                    {connections.map((connection) => {
                      const connProps = connection.properties || {}
                      const rel = connection.relationship || { properties: {} as Record<string, any>, direction: 'out', type: '' }
                      const relProps = rel.properties || {}

                      const connDisplayName = connProps.name || connProps.title || `Node ${connection.id}`
                      const connType = connection.type || 'Unknown'
                      const relDirection = rel.direction === 'out' ? 'out' : 'in'
                      const relType = rel.type || ''

                      return (
                        <li key={connection.id} className="px-4 py-4 sm:px-6 hover:bg-gray-50">
                          <div className="flex items-center justify-between">
                            <div className="flex flex-col space-y-1">
                              <div className="flex items-center space-x-3">
                                <div className={`w-2 h-2 rounded-full ${
                                  connType === 'Person' ? 'bg-green-500' :
                                  connType === 'Organization' ? 'bg-blue-500' :
                                  'bg-yellow-500'
                                }`} />
                                <Link
                                  href={`/nodes/${connection.id}`}
                                  className="text-sm font-medium text-indigo-600 hover:text-indigo-900"
                                >
                                  {connDisplayName}
                                </Link>
                                <span className={`inline-flex items-center px-1.5 py-0.5 rounded-full text-xs font-medium ${
                                  relDirection === 'out' ? 'bg-blue-100 text-blue-800' : 'bg-purple-100 text-purple-800'
                                }`}>
                                  {relDirection === 'out' ? '→' : '←'} {relType}
                                </span>
                              </div>
                              
                              <p className="text-xs text-gray-500">
                                {relDirection === 'out' ? 
                                  `${nodeProps.name || nodeProps.title || 'This node'} ${relType.toLowerCase()} ${connProps.name || connProps.title || 'this node'}` :
                                  `${connProps.name || connProps.title || 'This node'} ${relType.toLowerCase()} ${nodeProps.name || nodeProps.title || 'this node'}`
                                }
                              </p>
                              
                              {Object.keys(relProps).length > 0 && (
                                <div className="text-xs text-gray-500">
                                  <strong>Details:</strong> {Object.entries(relProps)
                                    .map(([key, value]) => `${key}: ${value}`)
                                    .join(', ')}
                                </div>
                              )}
                            </div>
                            
                            <div className="flex items-center space-x-2">
                              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                                {connType}
                              </span>
                            </div>
                          </div>
                        </li>
                      )
                    })}
                  </ul>
                )}
              </div>
            </div>
          </div>

          {/* Add Relationship Modal */}
          {showRelationshipForm && (
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-40">
              <div className="bg-white rounded-lg shadow-xl max-w-lg w-full p-6">
                <div className="flex justify-between items-center mb-4">
                  <h2 className="text-lg font-medium text-gray-900">Add Relationship</h2>
                  <button
                    type="button"
                    className="text-gray-400 hover:text-gray-500"
                    onClick={() => setShowRelationshipForm(false)}
                  >
                    <span className="sr-only">Close</span>
                    <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
                
                <div className="mb-4 p-3 bg-blue-50 rounded-md">
                  <p className="text-sm text-blue-800">
                    <strong>Source:</strong> {nodeProps.name || nodeProps.title || node.id}
                  </p>
                </div>
                
                <RelationshipForm
                  sourceId={node.id}
                  onComplete={handleRelationshipComplete}
                  onError={handleRelationshipError}
                />
              </div>
            </div>
          )}
        </div>
      </div>
    </>
  )
}
