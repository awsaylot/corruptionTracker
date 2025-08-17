// pages/nodes/index.tsx - Updated with full integration
import { useState, useEffect } from 'react'
import Head from 'next/head'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { Node, api } from '../../utils/api'
import NodeDetailsModal from '../../components/nodes/NodeDetailsModal'
import NotificationContainer from '../../components/ui/NotificationContainer'
import LoadingSpinner from '../../components/ui/LoadingSpinner'
import { useNotification } from '../../hooks/useNotification'

const NodesPage = () => {
  const router = useRouter()
  const { notifications, addNotification, removeNotification } = useNotification()
  
  // State management
  const [nodes, setNodes] = useState<Node[]>([])
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)
  const [showDetails, setShowDetails] = useState(false)
  
  // Loading states
  const [isLoading, setIsLoading] = useState(true)
  const [isUpdating, setIsUpdating] = useState(false)
  const [isDeleting, setIsDeleting] = useState<string | null>(null)
  
  // Error and filter states
  const [error, setError] = useState<string | null>(null)
  const [typeFilter, setTypeFilter] = useState<string>('all')
  const [searchQuery, setSearchQuery] = useState('')

  useEffect(() => {
    fetchNodes()
  }, [])

  const fetchNodes = async () => {
    try {
      setIsLoading(true)
      setError(null)
      
      const fetchedNodes = await api.getAllNodes()
      setNodes(fetchedNodes)
      
      addNotification('success', `Loaded ${fetchedNodes.length} nodes`)
      
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch nodes'
      setError(errorMessage)
      addNotification('error', errorMessage)
    } finally {
      setIsLoading(false)
    }
  }

  const handleNodeUpdate = async (nodeId: string, properties: any) => {
    try {
      setIsUpdating(true)
      setError(null)
      
      await api.updateNode(nodeId, properties)
      
      // Update the node in our local state
      setNodes(prevNodes => 
        prevNodes.map(node => 
          node.id === nodeId 
            ? { ...node, properties: { ...node.properties, ...properties } }
            : node
        )
      )
      
      // Update selected node if it's the one being updated
      if (selectedNode?.id === nodeId) {
        setSelectedNode(prev => prev ? { ...prev, properties: { ...prev.properties, ...properties } } : null)
      }
      
      addNotification('success', 'Node updated successfully')
      
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to update node'
      setError(errorMessage)
      addNotification('error', errorMessage)
      throw err // Re-throw for the modal to handle
    } finally {
      setIsUpdating(false)
    }
  }

  const handleNodeDelete = async (nodeId: string) => {
    try {
      setIsDeleting(nodeId)
      setError(null)
      
      await api.deleteNode(nodeId)
      
      // Remove from local state
      setNodes(prevNodes => prevNodes.filter(node => node.id !== nodeId))
      
      // Clear selection if deleted node was selected
      if (selectedNode?.id === nodeId) {
        setSelectedNode(null)
        setShowDetails(false)
      }
      
      addNotification('success', 'Node deleted successfully')
      
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to delete node'
      setError(errorMessage)
      addNotification('error', errorMessage)
      throw err // Re-throw for the modal to handle
    } finally {
      setIsDeleting(null)
    }
  }

  const handleNodeSelect = (node: Node) => {
    setSelectedNode(node)
    setShowDetails(true)
  }

  const handleViewDetails = (nodeId: string) => {
    router.push(`/nodes/${nodeId}`)
  }

  const handleRetry = () => {
    fetchNodes()
  }

  // Filter nodes based on type and search query
  const filteredNodes = nodes.filter(node => {
    const matchesType = typeFilter === 'all' || node.type === typeFilter
    const matchesSearch = !searchQuery || 
      (node.properties.name && node.properties.name.toLowerCase().includes(searchQuery.toLowerCase())) ||
      (node.properties.title && node.properties.title.toLowerCase().includes(searchQuery.toLowerCase())) ||
      node.id.toLowerCase().includes(searchQuery.toLowerCase())
    
    return matchesType && matchesSearch
  })

  // Get unique node types for filter
  const nodeTypes = ['all', ...Array.from(new Set(nodes.map(node => node.type)))]

  return (
    <>
      <Head>
        <title>Nodes - GraphDB</title>
        <meta name="description" content="List and manage all nodes in the graph database" />
      </Head>

      <NotificationContainer 
        notifications={notifications} 
        onRemove={removeNotification} 
      />

      {/* Header */}
      <div className="sm:flex sm:items-center sm:justify-between mb-6">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-bold text-gray-900">Nodes</h1>
          <p className="mt-2 text-sm text-gray-700">
            A list of all nodes in your graph database. Total: {nodes.length}
          </p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
          <Link
            href="/nodes/create"
            className="inline-flex items-center justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:w-auto"
          >
            <svg className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Add Node
          </Link>
        </div>
      </div>

      {/* Filters and Search */}
      <div className="mb-6 bg-white rounded-lg shadow-sm p-4">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1">
            <label className="block text-sm font-medium text-gray-700 mb-2">Search Nodes</label>
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search by name, title, or ID..."
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>
          
          <div className="sm:w-48">
            <label className="block text-sm font-medium text-gray-700 mb-2">Filter by Type</label>
            <select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
            >
              {nodeTypes.map(type => (
                <option key={type} value={type}>
                  {type === 'all' ? 'All Types' : type}
                </option>
              ))}
            </select>
          </div>
          
          <div className="sm:w-32 flex items-end">
            <button
              onClick={handleRetry}
              disabled={isLoading}
              className="w-full inline-flex items-center justify-center px-3 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
            >
              {isLoading ? <LoadingSpinner size="small" /> : 'Refresh'}
            </button>
          </div>
        </div>
        
        <div className="mt-3 flex items-center justify-between text-sm text-gray-600">
          <span>Showing {filteredNodes.length} of {nodes.length} nodes</span>
          {(searchQuery || typeFilter !== 'all') && (
            <button
              onClick={() => {
                setSearchQuery('')
                setTypeFilter('all')
              }}
              className="text-indigo-600 hover:text-indigo-700"
            >
              Clear filters
            </button>
          )}
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

      {/* Nodes Table */}
      <div className="bg-white shadow-sm rounded-lg overflow-hidden">
        <div className="min-w-full divide-y divide-gray-200">
          {/* Table Header */}
          <div className="bg-gray-50 px-6 py-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-medium text-gray-900">Node List</h3>
              {isUpdating && (
                <div className="flex items-center space-x-2 text-sm text-gray-600">
                  <LoadingSpinner size="small" />
                  <span>Updating...</span>
                </div>
              )}
            </div>
          </div>

          {/* Loading State */}
          {isLoading ? (
            <div className="px-6 py-12 text-center">
              <LoadingSpinner size="medium" text="Loading nodes..." />
            </div>
          ) : filteredNodes.length === 0 ? (
            <div className="px-6 py-12 text-center">
              <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2 2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-2.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 009.586 13H7" />
              </svg>
              <h3 className="mt-2 text-sm font-medium text-gray-900">
                {searchQuery || typeFilter !== 'all' ? 'No matching nodes' : 'No nodes found'}
              </h3>
              <p className="mt-1 text-sm text-gray-500">
                {searchQuery || typeFilter !== 'all' ? 
                  'Try adjusting your search or filter criteria.' :
                  'Get started by creating your first node.'
                }
              </p>
              <div className="mt-6">
                <Link
                  href="/nodes/create"
                  className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  <svg className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                  </svg>
                  Create First Node
                </Link>
              </div>
            </div>
          ) : (
            /* Nodes Grid */
            <div className="divide-y divide-gray-200">
              {filteredNodes.map((node) => (
                <div key={node.id} className="px-6 py-4 hover:bg-gray-50">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-4 min-w-0 flex-1">
                      {/* Node Type Indicator */}
                      <div className={`w-3 h-3 rounded-full flex-shrink-0 ${
                        node.type === 'Person' ? 'bg-green-500' :
                        node.type === 'Organization' ? 'bg-blue-500' :
                        'bg-yellow-500'
                      }`} />
                      
                      {/* Node Info */}
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center space-x-2">
                          <h4 className="text-sm font-medium text-gray-900 truncate">
                            {node.properties.name || node.properties.title || `Node ${node.id}`}
                          </h4>
                          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            node.type === 'Person' ? 'bg-green-100 text-green-800' :
                            node.type === 'Organization' ? 'bg-blue-100 text-blue-800' :
                            'bg-yellow-100 text-yellow-800'
                          }`}>
                            {node.type}
                          </span>
                        </div>
                        
                        <div className="mt-1 flex items-center space-x-4 text-sm text-gray-500">
                          <span>ID: {node.id}</span>
                          {node.properties.aliases && node.properties.aliases.length > 0 && (
                            <span className="truncate">
                              Aliases: {node.properties.aliases.slice(0, 2).join(', ')}
                              {node.properties.aliases.length > 2 && ` +${node.properties.aliases.length - 2} more`}
                            </span>
                          )}
                        </div>
                      </div>
                    </div>

                    {/* Actions */}
                    <div className="flex items-center space-x-2">
                      {isDeleting === node.id ? (
                        <LoadingSpinner size="small" />
                      ) : (
                        <>
                          <button
                            onClick={() => handleViewDetails(node.id)}
                            className="text-indigo-600 hover:text-indigo-900 text-sm font-medium"
                          >
                            View
                          </button>
                          <button
                            onClick={() => handleNodeSelect(node)}
                            className="text-gray-600 hover:text-gray-900 text-sm font-medium"
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => router.push(`/?select=${node.id}`)}
                            className="text-green-600 hover:text-green-900 text-sm font-medium"
                          >
                            Graph
                          </button>
                        </>
                      )}
                    </div>
                  </div>
                  
                  {/* Additional Properties */}
                  {(node.properties.notes || node.properties.description) && (
                    <div className="mt-2 text-sm text-gray-600 line-clamp-2">
                      {node.properties.notes || node.properties.description}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Pagination/Load More (Future Enhancement) */}
      {filteredNodes.length > 50 && (
        <div className="mt-6 text-center">
          <p className="text-sm text-gray-500">
            Showing first 50 results. Use filters to narrow down your search.
          </p>
        </div>
      )}

      {/* Node Details Modal */}
      {selectedNode && showDetails && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-40">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="relative">
              {isUpdating && (
                <div className="absolute inset-0 bg-white bg-opacity-75 flex items-center justify-center z-10">
                  <LoadingSpinner size="medium" text="Updating node..." />
                </div>
              )}
              
              <NodeDetailsModal
                node={selectedNode}
                onClose={() => {
                  setShowDetails(false)
                  setSelectedNode(null)
                }}
                onDelete={() => handleNodeDelete(selectedNode.id)}
                onUpdate={(properties) => handleNodeUpdate(selectedNode.id, properties)}
              />
            </div>
          </div>
        </div>
      )}
    </>
  )
}

export default NodesPage