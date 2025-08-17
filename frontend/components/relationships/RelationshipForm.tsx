// components/RelationshipForm.tsx - improved target resolution + error handling (copy/paste ready)
import { useState, useEffect } from 'react';
import { api, Node } from '../../utils/api';
import LoadingSpinner from '../ui/LoadingSpinner';

interface RelationshipFormProps {
  // accept numeric Neo4j ids (0) or string ids (UUIDs)
  sourceId?: string | number;
  onComplete?: () => void;
  onError?: (error: string) => void;
}

const RELATIONSHIP_TYPES = {
  Financial: [
    'FUNDED_BY',
    'GAVE_MONEY_TO',
    'INVESTED_IN',
    'RECEIVED_FROM',
  ],
  Professional: [
    'APPOINTED_BY',
    'RESIGNED_FROM',
    'WORKS_FOR',
    'LEADS',
  ],
  Actions: [
    'APPROVED_PROJECT_FOR',
    'INVESTIGATED_BY',
    'MET_WITH',
    'PARTNERED_WITH',
  ],
  Influence: [
    'LOBBIED_FOR',
    'SUPPORTED',
    'OPPOSED',
    'BRIBED',
  ]
} as const;

const RelationshipForm: React.FC<RelationshipFormProps> = ({ sourceId, onComplete, onError }) => {
  // debug: confirm what's being passed in (remove or comment out in prod)
  console.debug('RelationshipForm sourceId:', sourceId);

  type RelationshipType = typeof RELATIONSHIP_TYPES[keyof typeof RELATIONSHIP_TYPES][number];

  const [formData, setFormData] = useState<{
    targetId: string;
    category: keyof typeof RELATIONSHIP_TYPES;
    type: RelationshipType;
    date: string;
    description: string;
    sources: string;
  }>({
    targetId: '',
    category: 'Financial',
    type: RELATIONSHIP_TYPES['Financial'][0],
    date: '',
    description: '',
    sources: '',
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchLoading, setSearchLoading] = useState(false);
  const [targetNodes, setTargetNodes] = useState<Node[]>([]);
  const [showTargetSearch, setShowTargetSearch] = useState(false);
  const [targetSearchQuery, setTargetSearchQuery] = useState('');

  useEffect(() => {
    loadTargetNodes();
    setFormData(prev => ({ ...prev, targetId: '' }));
    setTargetSearchQuery('');
  }, [sourceId]);

  const loadTargetNodes = async () => {
    if (sourceId == null) return;

    try {
      setSearchLoading(true);
      const allNodes = await api.getAllNodes();
      setTargetNodes(allNodes.filter(node => String(node.id) !== String(sourceId)));
    } catch (err) {
      console.error('Failed to load target nodes:', err);
    } finally {
      setSearchLoading(false);
    }
  };

  const handleCategoryChange = (category: keyof typeof RELATIONSHIP_TYPES) => {
    setFormData(prev => ({
      ...prev,
      category,
      type: RELATIONSHIP_TYPES[category][0]
    }));
  };

  const handleTargetSearch = async (query: string) => {
    setTargetSearchQuery(query);
    setShowTargetSearch(true);

    if (!query.trim()) {
      loadTargetNodes();
      return;
    }

    try {
      setSearchLoading(true);
      const results = await api.searchNodes(query);
      const nodes = 'nodes' in results ? results.nodes : (results as Node[]);
      setTargetNodes(nodes.filter((node: Node) => String(node.id) !== String(sourceId)));
    } catch (err) {
      console.error('Search failed:', err);
    } finally {
      setSearchLoading(false);
    }
  };

  const resolveTargetIfNeeded = async (): Promise<string | null> => {
    // If targetId is already set, we're done
    if (formData.targetId) return formData.targetId;

    // If user typed something, try to resolve it
    const q = targetSearchQuery?.trim();
    if (!q) return null;

    // If they typed an ID-like value, try to use it first
    const byIdCandidate = targetNodes.find(n => String(n.id) === q);
    if (byIdCandidate) return String(byIdCandidate.id);

    // Otherwise search the API for that query
    try {
      setSearchLoading(true);
      const results = await api.searchNodes(q);
      const nodes = 'nodes' in results ? results.nodes : (results as Node[]);
      const filtered = nodes.filter((n: Node) => String(n.id) !== String(sourceId));

      if (filtered.length === 1) {
        return String(filtered[0].id);
      }

      if (filtered.length === 0) {
        setError('No matching node found for the target. Please pick one from the list.');
        return null;
      }

      setError('Multiple matches found — please pick the correct target from the list.');
      return null;
    } catch (err) {
      console.error('Resolve target failed:', err);
      setError('Failed to resolve target node. Please choose from the dropdown.');
      return null;
    } finally {
      setSearchLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Only check for null/undefined — allow 0 or empty-string IDs if applicable
    if (sourceId == null) {
      const errorMsg = 'Source node ID is required';
      setError(errorMsg);
      onError?.(errorMsg);
      return;
    }

    setError(null);
    setLoading(true);

    try {
      const resolvedTargetId = await resolveTargetIfNeeded();
      if (!resolvedTargetId) {
        setLoading(false);
        return;
      }

      const relationshipProperties: Record<string, any> = {};
      if (formData.date) relationshipProperties.date = formData.date;
      if (formData.description) relationshipProperties.description = formData.description;
      if (formData.sources) {
        relationshipProperties.sources = formData.sources
          .split('\n')
          .map(s => s.trim())
          .filter(Boolean);
      }

      const payload = {
        sourceId: String(sourceId),
        targetId: String(resolvedTargetId),
        type: formData.type,
        properties: relationshipProperties
      };
      console.debug('createRelationship payload:', payload);

      await api.createRelationship(
        String(sourceId),
        String(resolvedTargetId),
        formData.type,
        relationshipProperties
      );

      setFormData({
        targetId: '',
        category: 'Financial',
        type: RELATIONSHIP_TYPES.Financial[0],
        date: '',
        description: '',
        sources: '',
      });
      setTargetSearchQuery('');
      setShowTargetSearch(false);
      setTargetNodes([]);

      onComplete?.();
    } catch (err: any) {
      let errorMsg = 'Failed to create relationship';
      if (err?.response?.data?.error) errorMsg = String(err.response.data.error);
      else if (err?.message) errorMsg = err.message;
      setError(errorMsg);
      onError?.(errorMsg);
      console.error('createRelationship error:', err);
    } finally {
      setLoading(false);
    }
  };

  const selectTarget = (node: Node) => {
    setFormData(prev => ({ ...prev, targetId: String(node.id) }));
    setShowTargetSearch(false);
    setTargetSearchQuery(node.properties?.name || node.properties?.title || String(node.id));
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="bg-red-50 border-l-4 border-red-400 p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <p className="text-sm text-red-700">{error}</p>
            </div>
          </div>
        </div>
      )}

      {/* Target Node Selection */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">Target Node</label>
        <div className="relative">
          <input
            type="text"
            value={targetSearchQuery}
            onChange={(e) => {
              handleTargetSearch(e.target.value);
              setFormData(prev => ({ ...prev, targetId: '' })); // clear previously selected id
            }}
            onFocus={() => setShowTargetSearch(true)}
            placeholder="Search for target node..."
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
          />

          {searchLoading && (
            <div className="absolute right-3 top-3">
              <LoadingSpinner size="small" />
            </div>
          )}
        </div>

        {/* Target Node Dropdown */}
        {showTargetSearch && targetNodes.length > 0 && (
          <div className="absolute z-10 w-full mt-1 bg-white border rounded-md shadow-lg max-h-60 overflow-y-auto">
            {targetNodes.map((node) => (
              <button
                key={node.id}
                type="button"
                onClick={() => selectTarget(node)}
                className="w-full px-4 py-2 text-left hover:bg-gray-100 focus:bg-gray-100 focus:outline-none"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <div className={`w-2 h-2 rounded-full ${
                      node.type === 'Person' ? 'bg-green-500' :
                      node.type === 'Organization' ? 'bg-blue-500' :
                      'bg-yellow-500'
                    }`} />
                    <span className="text-sm font-medium">
                      {node.properties?.name || node.properties?.title || String(node.id)}
                    </span>
                  </div>
                  <span className="text-xs text-gray-500">{node.type}</span>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Category */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">Category</label>
        <select
          value={formData.category}
          onChange={(e) => handleCategoryChange(e.target.value as keyof typeof RELATIONSHIP_TYPES)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
        >
          {Object.keys(RELATIONSHIP_TYPES).map(category => (
            <option key={category} value={category}>{category}</option>
          ))}
        </select>
      </div>

      {/* Relationship Type */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">Relationship Type</label>
        <select
          value={formData.type}
          onChange={(e) => setFormData(prev => ({ ...prev, type: e.target.value as RelationshipType }))}
          className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
        >
          {RELATIONSHIP_TYPES[formData.category].map(type => (
            <option key={type} value={type}>{type.replace(/_/g, ' ')}</option>
          ))}
        </select>
      </div>

      {/* Date */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">Date (Optional)</label>
        <input
          type="date"
          value={formData.date}
          onChange={(e) => setFormData(prev => ({ ...prev, date: e.target.value }))}
          className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
        />
      </div>

      {/* Description */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">Description (Optional)</label>
        <textarea
          value={formData.description}
          onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
          className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
          rows={3}
          placeholder="Describe the relationship..."
        />
      </div>

      {/* Sources */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">Sources (Optional)</label>
        <textarea
          value={formData.sources}
          onChange={(e) => setFormData(prev => ({ ...prev, sources: e.target.value }))}
          className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
          rows={3}
          placeholder="One source URL per line..."
        />
      </div>

      {/* Actions */}
      <div className="flex justify-end space-x-3 pt-4">
        <button
          type="button"
          onClick={() => {
            setFormData({
              targetId: '',
              category: 'Financial',
              type: RELATIONSHIP_TYPES.Financial[0],
              date: '',
              description: '',
              sources: '',
            });
            setError(null);
            setTargetSearchQuery('');
          }}
          className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          disabled={loading}
        >
          Reset
        </button>
        <button
          type="submit"
          disabled={loading || sourceId == null}
          className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:bg-gray-300 disabled:cursor-not-allowed"
        >
          {loading ? (
            <>
              <LoadingSpinner size="small" color="white" />
              <span className="ml-2">Creating...</span>
            </>
          ) : (
            'Create Relationship'
          )}
        </button>
      </div>
    </form>
  );
};

export default RelationshipForm;
