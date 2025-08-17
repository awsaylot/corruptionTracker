// pages/index.tsx - Added relationship selection handling
import { useState, useCallback, useEffect } from 'react'
import Head from 'next/head'
import { useRouter } from 'next/router'
import D3Graph from '../components/graph/D3Graph'
import SearchBar from '../components/search/SearchBar'
import NodeDetailsModal from '../components/nodes/NodeDetailsModal'
import RelationshipForm from '../components/relationships/RelationshipForm'
import NotificationContainer from '../components/ui/NotificationContainer'
import LoadingSpinner from '../components/ui/LoadingSpinner'
import { useNotification } from '../hooks/useNotification'
import { Node, api } from '../utils/api'

/**
 * Type guard: checks whether a Node has a `connections` array.
 * This makes rendering safe and satisfies TypeScript.
 */
function hasConnections(node: Node | null | undefined): node is Node & { connections: any[] } {
  return !!node && Array.isArray((node as any).connections)
}

// Interface for relationship data from D3Graph
interface RelationshipData {
  type: string;
  properties?: Record<string, any>;
  sourceNode: {
    id: string;
    label: string;
    type: string;
  };
  targetNode: {
    id: string;
    label: string;
    type: string;
  };
}

export default function Home() {
  const router = useRouter()
  const { notifications, addNotification, removeNotification } = useNotification()
  
  // State management
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)
  const [selectedRelationship, setSelectedRelationship] = useState<RelationshipData | null>(null)
  const [showNodeDetails, setShowNodeDetails] = useState(false)
  const [showRelationshipDetails, setShowRelationshipDetails] = useState(false)
  const [showRelationshipForm, setShowRelationshipForm] = useState(false)
  const [refreshTrigger, setRefreshTrigger] = useState(0)
  
  // Loading states
  const [isInitialLoading, setIsInitialLoading] = useState(true)
  const [isNodeLoading, setIsNodeLoading] = useState(false)
  
  // Error state
  const [globalError, setGlobalError] = useState<string | null>(null)

  // Initialize the application
  const initializeApp = async () => {
    try {
      setIsInitialLoading(true)
      setGlobalError(null)
      
      // Test API connection
      await api.getAllNodes()
      addNotification('success', 'Connected to graph database successfully')
      
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to connect to database'
      setGlobalError(errorMessage)
      addNotification('error', `Database connection failed: ${errorMessage}`)
    } finally {
      setIsInitialLoading(false)
    }
  }

  // Fixed graph node selection handler
  const handleGraphNodeSelect = useCallback(async (nodeData: any) => {
    try {
      setIsNodeLoading(true)
      setGlobalError(null)
      
      // Clear any selected relationship when selecting a node
      setSelectedRelationship(null)
      setShowRelationshipDetails(false)
      
      // Extract node ID from different possible formats
      let nodeId: string | null = null
      
      if (typeof nodeData === 'string') {
        nodeId = nodeData
      } else if (typeof nodeData === 'number') {
        // Handle numeric IDs from D3
        nodeId = String(nodeData)
      } else if (nodeData && typeof nodeData === 'object') {
        // Handle different node data formats from D3
        if (nodeData.id !== undefined && nodeData.id !== null) {
          nodeId = String(nodeData.id)
        } else if (nodeData.data && nodeData.data.id !== undefined && nodeData.data.id !== null) {
          nodeId = String(nodeData.data.id)
        } else if (nodeData.properties && nodeData.properties.id !== undefined && nodeData.properties.id !== null) {
          nodeId = String(nodeData.properties.id)
        }
      }
      
      if (nodeId === null || nodeId === 'null' || nodeId === 'undefined') {
        console.error('Unable to extract node ID from:', nodeData)
        addNotification('error', 'Unable to identify selected node')
        return
      }
      
      console.log('Fetching node data for ID:', nodeId)
      const fullNodeData = await api.getNode(nodeId)
      setSelectedNode(fullNodeData)
      setShowNodeDetails(true)

      addNotification('info', `Selected ${fullNodeData.properties?.name || fullNodeData.properties?.title || fullNodeData.id}`)
      
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to fetch node'
      setGlobalError(errorMessage)
      addNotification('error', `Failed to select node: ${errorMessage}`)
      console.error('Node selection error:', error)
    } finally {
      setIsNodeLoading(false)
    }
  }, [addNotification])

  // Handle relationship selection from D3Graph
  const handleRelationshipSelect = useCallback((relationshipData: RelationshipData) => {
    console.log('Relationship selected:', relationshipData)
    
    // Clear any selected node when selecting a relationship
    setSelectedNode(null)
    setShowNodeDetails(false)
    
    // Set the selected relationship and show details
    setSelectedRelationship(relationshipData)
    setShowRelationshipDetails(true)
    
    addNotification('info', `Selected ${relationshipData.type} relationship`)
  }, [addNotification])

  useEffect(() => {
    initializeApp()
  }, [])

  // Handle query parameter for node selection
  useEffect(() => {
    const { select } = router.query
    if (select && typeof select === 'string') {
      handleGraphNodeSelect(select)
    }
  }, [router.query, handleGraphNodeSelect])

  const handleNodeSelect = useCallback(async (node: Node) => {
    try {
      setIsNodeLoading(true)
      setGlobalError(null)
      
      // Clear relationship selection
      setSelectedRelationship(null)
      setShowRelationshipDetails(false)
      
      // Fetch full node details with connections
      const nodeData = await api.getNode(node.id)
      setSelectedNode(nodeData)
      setShowNodeDetails(true)
      
      addNotification('info', `Selected ${node.properties.name || node.properties.title || node.id}`)
      
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to fetch node details'
      setGlobalError(errorMessage)
      addNotification('error', `Failed to load node details: ${errorMessage}`)
    } finally {
      setIsNodeLoading(false)
    }
  }, [addNotification])

  const handleNodeUpdate = useCallback(async (nodeId: string, properties: any) => {
    try {
      setGlobalError(null)
      
      await api.updateNode(nodeId, properties)
      
      // Refresh the selected node data
      const updatedNode = await api.getNode(nodeId)
      setSelectedNode(updatedNode)
      
      // Trigger graph refresh
      setRefreshTrigger(prev => prev + 1)
      
      addNotification('success', 'Node updated successfully')
      
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to update node'
      setGlobalError(errorMessage)
      addNotification('error', `Update failed: ${errorMessage}`)
      throw error // Re-throw for the modal to handle
    }
  }, [addNotification])

  const handleNodeDelete = useCallback(async (nodeId: string) => {
    try {
      setGlobalError(null)
      
      await api.deleteNode(nodeId)
      
      // Clear selection and close modals
      setSelectedNode(null)
      setShowNodeDetails(false)
      setShowRelationshipForm(false)
      
      // Trigger graph refresh
      setRefreshTrigger(prev => prev + 1)
      
      addNotification('success', 'Node deleted successfully')
      
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to delete node'
      setGlobalError(errorMessage)
      addNotification('error', `Delete failed: ${errorMessage}`)
      throw error // Re-throw for the modal to handle
    }
  }, [addNotification])

  const handleCreateRelationship = useCallback(() => {
    if (!selectedNode) {
      addNotification('warning', 'Please select a node first')
      return
    }
    setShowRelationshipForm(true)
  }, [selectedNode, addNotification])

  const handleRelationshipComplete = useCallback(() => {
    setShowRelationshipForm(false)
    
    // Refresh the graph and selected node
    setRefreshTrigger(prev => prev + 1)
    
    if (selectedNode) {
      // Refresh selected node data to show new relationship
      api.getNode(selectedNode.id)
        .then(updatedNode => {
          setSelectedNode(updatedNode)
          addNotification('success', 'Relationship created successfully')
        })
        .catch(error => {
          const errorMessage = error instanceof Error ? error.message : 'Failed to refresh node data'
          addNotification('error', `Failed to refresh: ${errorMessage}`)
        })
    }
  }, [selectedNode, addNotification])

  const handleCloseModals = useCallback(() => {
    setShowNodeDetails(false)
    setShowRelationshipForm(false)
    setShowRelationshipDetails(false)
    setSelectedNode(null)
    setSelectedRelationship(null)
  }, [])

  const handleRetryConnection = useCallback(() => {
    initializeApp()
  }, [])

  const handleNavigateToNodes = () => {
    router.push('/nodes')
  }

  const handleNavigateToCreateNode = () => {
    router.push('/nodes/create')
  }



  // Loading screen for initial app load
  if (isInitialLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <LoadingSpinner size="large" text="Loading Graph Database..." />
        </div>
      </div>
    )
  }

  // Error screen for critical failures
  if (globalError && !selectedNode) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center max-w-md mx-auto">
          <div className="bg-red-50 border-l-4 border-red-400 p-4 mb-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">Connection Error</h3>
                <div className="mt-2 text-sm text-red-700">
                  <p>{globalError}</p>
                </div>
              </div>
            </div>
          </div>
          <button
            onClick={handleRetryConnection}
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            Retry Connection
          </button>
        </div>
      </div>
    )
  }

  const connectionsCount = hasConnections(selectedNode) ? selectedNode.connections.length : 0

  return (
    <>
      <Head>
        <title>Graph View - GraphDB</title>
        <meta name="description" content="Interactive graph visualization of entities and their relationships" />
      </Head>

      {/* Notification System */}
      <NotificationContainer 
        notifications={notifications} 
        onRemove={removeNotification} 
      />

      {/* Header with Actions */}
      <div className="mb-6">
        <div className="sm:flex sm:items-center sm:justify-between">
          <div className="sm:flex-auto">
            <h1 className="text-2xl font-bold text-gray-900">Graph View</h1>
            <p className="mt-2 text-sm text-gray-700">
              Interactive visualization of all nodes and their relationships.
            </p>
          </div>
          <div className="mt-4 sm:mt-0 sm:flex sm:space-x-3">
            <button
              onClick={handleNavigateToNodes}
              className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              View All Nodes
            </button>
            <button
              onClick={handleNavigateToCreateNode}
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              Create Node
            </button>
          </div>
        </div>
      </div>

      {/* Search Bar */}
      <div className="mb-6">
        <div className="bg-white rounded-lg shadow-sm p-4">
          <SearchBar onSelectNode={handleNodeSelect} />
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex gap-6 min-h-[600px]">
        {/* Graph Visualization */}
        <div className="flex-1">
          <div className="bg-white rounded-lg shadow-lg p-4 h-full">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-medium text-gray-900">Network Visualization</h2>
              <div className="flex items-center space-x-2">
                {isNodeLoading && <LoadingSpinner size="small" />}
                <button
                  onClick={() => setRefreshTrigger(prev => prev + 1)}
                  className="inline-flex items-center px-3 py-1 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
                >
                  Refresh
                </button>
              </div>
            </div>
            
            <div className="h-[calc(100vh-350px)] bg-gray-50 rounded border relative">
              <D3Graph 
                onNodeSelect={handleGraphNodeSelect}
                onRelationshipSelect={handleRelationshipSelect}
                selectedNode={selectedNode?.id}
                refreshTrigger={refreshTrigger}
              />
              
              {/* Overlay loading for graph operations */}
              {isNodeLoading && (
                <div className="absolute inset-0 bg-white bg-opacity-75 flex items-center justify-center">
                  <LoadingSpinner size="medium" text="Loading node details..." />
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Side Panel */}
        <div className="w-96 space-y-4">
          {/* Quick Actions */}
          <div className="bg-white rounded-lg shadow-sm p-4">
            <h3 className="text-sm font-medium text-gray-900 mb-3">Quick Actions</h3>
            <div className="space-y-2">
              <button
                onClick={handleCreateRelationship}
                disabled={!selectedNode}
                className="w-full inline-flex justify-center items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:bg-gray-300 disabled:cursor-not-allowed"
              >
                Add Relationship
              </button>
              <button
                onClick={() => router.push('/nodes/create')}
                className="w-full inline-flex justify-center items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                Create New Node
              </button>
            </div>
          </div>

          {/* Selected Node Info */}
          {selectedNode && (
            <div className="bg-white rounded-lg shadow-sm p-4">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-medium text-gray-900">Selected Node</h3>
                <button
                  onClick={() => setSelectedNode(null)}
                  className="text-gray-400 hover:text-gray-500"
                >
                  <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              
              <div className="space-y-2">
                <div className="flex items-center space-x-2">
                  <div className={`w-2 h-2 rounded-full ${
                    selectedNode.type === 'Person' ? 'bg-green-500' :
                    selectedNode.type === 'Organization' ? 'bg-blue-500' :
                    'bg-yellow-500'
                  }`} />
                  <span className="text-sm font-medium">
                    {selectedNode.properties.name || selectedNode.properties.title || selectedNode.id}
                  </span>
                  <span className="text-xs text-gray-500">({selectedNode.type})</span>
                </div>
                
                {hasConnections(selectedNode) && (
                  <p className="text-xs text-gray-500">
                    {selectedNode.connections.length} connection{selectedNode.connections.length !== 1 ? 's' : ''}
                  </p>
                )}
                
                <div className="flex space-x-2 mt-3">
                  <button
                    onClick={() => setShowNodeDetails(true)}
                    className="text-xs px-3 py-1 bg-indigo-100 text-indigo-700 rounded-full hover:bg-indigo-200"
                  >
                    View Details
                  </button>
                  <button
                    onClick={() => router.push(`/nodes/${selectedNode.id}`)}
                    className="text-xs px-3 py-1 bg-gray-100 text-gray-700 rounded-full hover:bg-gray-200"
                  >
                    Full Page
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Selected Relationship Info */}
          {selectedRelationship && (
            <div className="bg-white rounded-lg shadow-sm p-4">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-medium text-gray-900">Selected Relationship</h3>
                <button
                  onClick={() => {
                    setSelectedRelationship(null)
                    setShowRelationshipDetails(false)
                  }}
                  className="text-gray-400 hover:text-gray-500"
                >
                  <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              
              <div className="space-y-2">
                <div className="text-sm font-medium text-center py-2 px-3 bg-gray-100 rounded">
                  {selectedRelationship.type}
                </div>
                
                <div className="text-xs text-gray-600">
                  <div className="flex items-center justify-between">
                    <span>{selectedRelationship.sourceNode.label}</span>
                    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                    </svg>
                    <span>{selectedRelationship.targetNode.label}</span>
                  </div>
                </div>
                
                {selectedRelationship.properties && Object.keys(selectedRelationship.properties).length > 0 && (
                  <div className="mt-2 text-xs">
                    <span className="font-medium text-gray-700">Properties:</span>
                    {Object.entries(selectedRelationship.properties).map(([key, value]) => (
                      <div key={key} className="ml-2 text-gray-600">
                        {key}: {String(value)}
                      </div>
                    ))}
                  </div>
                )}
                
                <div className="flex space-x-2 mt-3">
                  <button
                    onClick={() => setShowRelationshipDetails(true)}
                    className="text-xs px-3 py-1 bg-purple-100 text-purple-700 rounded-full hover:bg-purple-200"
                  >
                    View Details
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Graph Statistics */}
          <div className="bg-white rounded-lg shadow-sm p-4">
            <h3 className="text-sm font-medium text-gray-900 mb-3">Graph Statistics</h3>
            <div className="space-y-2 text-sm text-gray-600">
              <div className="flex justify-between">
                <span>Selected Node:</span>
                <span>{selectedNode ? '1' : '0'}</span>
              </div>
              <div className="flex justify-between">
                <span>Selected Relationship:</span>
                <span>{selectedRelationship ? '1' : '0'}</span>
              </div>
              <div className="flex justify-between">
                <span>Connections:</span>
                <span>{connectionsCount}</span>
              </div>
              <div className="flex justify-between">
                <span>Refresh Count:</span>
                <span>{refreshTrigger}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Node Details Modal */}
      {selectedNode && showNodeDetails && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-40">
          <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
            <NodeDetailsModal
              node={selectedNode}
              onClose={() => setShowNodeDetails(false)}
              onDelete={() => handleNodeDelete(selectedNode.id)}
              onUpdate={(properties) => handleNodeUpdate(selectedNode.id, properties)}
            />
          </div>
        </div>
      )}

      {/* Relationship Details Modal */}
      {selectedRelationship && showRelationshipDetails && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-40">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-lg font-medium text-gray-900">Relationship Details</h2>
                <button
                  onClick={() => setShowRelationshipDetails(false)}
                  className="text-gray-400 hover:text-gray-500"
                >
                  <span className="sr-only">Close</span>
                  <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              
              <div className="space-y-4">
                <div className="bg-gray-50 p-4 rounded-lg">
                  <h3 className="text-lg font-semibold text-center mb-4">{selectedRelationship.type}</h3>
                  
                  <div className="flex items-center justify-between">
                    <div className="text-center flex-1">
                      <div className={`w-3 h-3 rounded-full mx-auto mb-2 ${
                        selectedRelationship.sourceNode.type === 'Person' ? 'bg-green-500' :
                        selectedRelationship.sourceNode.type === 'Organization' ? 'bg-blue-500' :
                        'bg-yellow-500'
                      }`} />
                      <div className="font-medium">{selectedRelationship.sourceNode.label}</div>
                      <div className="text-sm text-gray-500">({selectedRelationship.sourceNode.type})</div>
                    </div>
                    
                    <div className="flex-shrink-0 mx-4">
                      <svg className="h-8 w-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                      </svg>
                    </div>
                    
                    <div className="text-center flex-1">
                      <div className={`w-3 h-3 rounded-full mx-auto mb-2 ${
                        selectedRelationship.targetNode.type === 'Person' ? 'bg-green-500' :
                        selectedRelationship.targetNode.type === 'Organization' ? 'bg-blue-500' :
                        'bg-yellow-500'
                      }`} />
                      <div className="font-medium">{selectedRelationship.targetNode.label}</div>
                      <div className="text-sm text-gray-500">({selectedRelationship.targetNode.type})</div>
                    </div>
                  </div>
                </div>
                
                {selectedRelationship.properties && Object.keys(selectedRelationship.properties).length > 0 && (
                  <div>
                    <h4 className="font-medium text-gray-900 mb-2">Properties</h4>
                    <div className="bg-gray-50 rounded p-3">
                      <dl className="space-y-2">
                        {Object.entries(selectedRelationship.properties).map(([key, value]) => (
                          <div key={key} className="flex justify-between">
                            <dt className="text-sm font-medium text-gray-700">{key}:</dt>
                            <dd className="text-sm text-gray-900">{String(value)}</dd>
                          </div>
                        ))}
                      </dl>
                    </div>
                  </div>
                )}
                
                <div className="flex space-x-3 pt-4">
                  <button
                    onClick={() => handleGraphNodeSelect(selectedRelationship.sourceNode.id)}
                    className="flex-1 px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
                  >
                    View Source Node
                  </button>
                  <button
                    onClick={() => handleGraphNodeSelect(selectedRelationship.targetNode.id)}
                    className="flex-1 px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
                  >
                    View Target Node
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Relationship Form Modal */}
      {selectedNode && showRelationshipForm && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-40">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full p-6">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-lg font-medium text-gray-900">Create Relationship</h2>
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
            
            <div className="mb-4 p-3 bg-gray-50 rounded-md">
              <p className="text-sm text-gray-600">
                <strong>Source Node:</strong> {selectedNode.properties.name || selectedNode.properties.title || selectedNode.id}
              </p>
            </div>
            
            <RelationshipForm
              sourceId={selectedNode.id}
              onComplete={handleRelationshipComplete}
            />
          </div>
        </div>
      )}
    </>
  )
}