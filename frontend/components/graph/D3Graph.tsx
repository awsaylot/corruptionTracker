import { useEffect, useRef, useState } from 'react';
import * as d3 from 'd3';
import { Node, NodeWithConnections, api } from '../../utils/api';

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

interface D3GraphProps {
    onNodeSelect?: (nodeId: string) => void;
    onRelationshipSelect?: (relationshipData: RelationshipData) => void;
    selectedNode?: string;
    refreshTrigger?: number;
}

const NODE_COLORS = {
    Person: '#4CAF50',
    Organization: '#2196F3',
    Event: '#FFC107'
};

// Color scheme for relationship types
const RELATIONSHIP_COLORS = {
    'WORKS_FOR': '#3b82f6',
    'KNOWS': '#10b981',
    'ATTENDED': '#f59e0b',
    'LOCATED_AT': '#ef4444',
    'OWNS': '#8b5cf6',
    'MARRIED_TO': '#ec4899',
    'PARENT_OF': '#14b8a6',
    'PART_OF': '#f97316',
    'MEMBER_OF': '#06b6d4',
    'FOUNDED': '#84cc16',
    'DEFAULT': '#6b7280'
};

interface D3Node extends d3.SimulationNodeDatum {
    id: string;
    label: string;
    type: string;
    properties: Record<string, any>;
    x?: number;
    y?: number;
    fx?: number | null;
    fy?: number | null;
}

interface D3Link {
    source: D3Node;
    target: D3Node;
    type: string;
    properties?: Record<string, any>;
    index?: number;
}

const D3Graph: React.FC<D3GraphProps> = ({ onNodeSelect, onRelationshipSelect, selectedNode, refreshTrigger }) => {
    const svgRef = useRef<SVGSVGElement>(null);
    const containerRef = useRef<SVGGElement | null>(null);
    const simulation = useRef<d3.Simulation<D3Node, D3Link>>();
    // Store the complete network data
    const [networkData, setNetworkData] = useState<NodeWithConnections[]>([]);
    // Store the currently visible nodes and links (for filtering)
    const nodesRef = useRef<D3Node[]>([]);
    const linksRef = useRef<D3Link[]>([]);

    const getRelationshipColor = (type: string): string => {
        return RELATIONSHIP_COLORS[type as keyof typeof RELATIONSHIP_COLORS] || RELATIONSHIP_COLORS.DEFAULT;
    };

    const updateGraph = (nodeData: NodeWithConnections) => {
        if (!nodeData) return;

        // Clear existing data
        nodesRef.current = [];
        linksRef.current = [];

        // Add main node first
        nodesRef.current.push({
            id: nodeData.id,
            label: (nodeData.properties && (nodeData.properties.name || nodeData.properties.title)) || `Unnamed ${nodeData.type}`,
            type: nodeData.type,
            properties: nodeData.properties || {},
            x: Math.random() * 800,
            y: Math.random() * 600
        });

        // Add all connected nodes before creating links
        nodeData.connections.forEach(conn => {
            if (!nodesRef.current.find(n => n.id === conn.id)) {
                nodesRef.current.push({
                    id: conn.id,
                    label: (conn.properties && (conn.properties.name || conn.properties.title)) || `Unnamed ${conn.type}`,
                    type: conn.type,
                    properties: conn.properties || {},
                    x: Math.random() * 800,
                    y: Math.random() * 600
                });
            }
        });

        // Now add links, ensuring both source and target nodes exist
        nodeData.connections.forEach(conn => {
            const sourceId = conn.relationship.direction === 'out' ? nodeData.id : conn.id;
            const targetId = conn.relationship.direction === 'out' ? conn.id : nodeData.id;
            
            // Verify both nodes exist before creating the link
            const sourceNode = nodesRef.current.find(n => n.id === sourceId);
            const targetNode = nodesRef.current.find(n => n.id === targetId);
            
            if (sourceNode && targetNode) {
                linksRef.current.push({
                    source: sourceNode,
                    target: targetNode,
                    type: conn.relationship.type,
                    properties: conn.relationship.properties
                });
            }
        });

        console.log('Updated graph data:', { 
            nodes: nodesRef.current.length, 
            links: linksRef.current.length 
        });

        renderGraph();
    };

    const renderGraph = () => {
        if (!svgRef.current || nodesRef.current.length === 0) return;

        console.log('Rendering graph with nodes:', nodesRef.current.length, 'links:', linksRef.current.length);

        const svg = d3.select(svgRef.current);
        const width = svg.node()?.getBoundingClientRect().width || 800;
        const height = svg.node()?.getBoundingClientRect().height || 600;

        // Clear previous content
        svg.selectAll('*').remove();

        // Create container for zoom/pan
        const container = svg.append('g');
        containerRef.current = container.node();

        // Set up zoom behavior
        const zoom = d3.zoom<SVGSVGElement, unknown>()
            .scaleExtent([0.1, 4])
            .on('zoom', (event) => {
                container.attr('transform', event.transform);
            });
        svg.call(zoom);

        // Create arrow marker for directed edges
        svg.append('defs').selectAll('marker')
            .data(['arrow'])
            .enter().append('marker')
            .attr('id', 'arrow')
            .attr('viewBox', '0 -5 10 10')
            .attr('refX', 20)
            .attr('refY', 0)
            .attr('markerWidth', 6)
            .attr('markerHeight', 6)
            .attr('orient', 'auto')
            .append('path')
            .attr('d', 'M0,-5L10,0L0,5')
            .attr('fill', '#999');

        // The links are already properly structured with D3Node objects
        const links = linksRef.current;

        // Create the simulation with validated data
        simulation.current = d3.forceSimulation(nodesRef.current)
            .force('link', d3.forceLink(links)
                .id((d: any) => d.id)
                .distance(150)) // Increased distance to make room for labels
            .force('charge', d3.forceManyBody().strength(-1200))
            .force('center', d3.forceCenter(width / 2, height / 2))
            .force('collide', d3.forceCollide().radius(60));

        // Create link groups (for line + label)
        const linkGroup = container.append('g')
            .attr('class', 'links')
            .selectAll('g')
            .data(links)
            .enter().append('g')
            .attr('class', 'link-group');

        // Create the actual lines
        const linkLines = linkGroup.append('line')
            .attr('stroke', d => getRelationshipColor(d.type))
            .attr('stroke-width', 2)
            .attr('marker-end', 'url(#arrow)')
            .attr('opacity', 0.8);

        // Create relationship label backgrounds (rectangles)
        const labelBg = linkGroup.append('rect')
            .attr('fill', d => getRelationshipColor(d.type))
            .attr('stroke', '#fff')
            .attr('stroke-width', 1)
            .attr('rx', 4)
            .attr('ry', 4)
            .attr('opacity', 0.9);

        // Create relationship labels (text)
        const labelText = linkGroup.append('text')
            .text(d => d.type)
            .attr('font-size', '10px')
            .attr('font-weight', 'bold')
            .attr('fill', '#fff')
            .attr('text-anchor', 'middle')
            .attr('dominant-baseline', 'central')
            .style('pointer-events', 'none');

        // Calculate text dimensions and set rectangle size
        labelText.each(function(d) {
            const element = this as SVGTextElement;
            const bbox = element.getBBox();
            const padding = 4;
            const parent = element.parentNode as Element;
            d3.select(parent).select('rect')
                .attr('width', bbox.width + padding * 2)
                .attr('height', bbox.height + padding * 2)
                .attr('x', -(bbox.width + padding * 2) / 2)
                .attr('y', -(bbox.height + padding * 2) / 2);
        });

        // Add hover effects and click handlers for relationship labels
        linkGroup
            .style('cursor', 'pointer')
            .on('click', function(event, d) {
                event.stopPropagation();
                
                // Create relationship data for the callback
                const relationshipData: RelationshipData = {
                    type: d.type,
                    properties: d.properties,
                    sourceNode: {
                        id: (d.source as D3Node).id,
                        label: (d.source as D3Node).label,
                        type: (d.source as D3Node).type
                    },
                    targetNode: {
                        id: (d.target as D3Node).id,
                        label: (d.target as D3Node).label,
                        type: (d.target as D3Node).type
                    }
                };
                
                console.log('Relationship clicked:', relationshipData);
                onRelationshipSelect?.(relationshipData);
            })
            .on('mouseenter', function(event, d) {
                d3.select(this).select('rect')
                    .attr('opacity', 1)
                    .attr('stroke-width', 2);
                
                // Show tooltip with relationship properties if they exist
                if (d.properties && Object.keys(d.properties).length > 0) {
                    const tooltip = d3.select('body').append('div')
                        .attr('class', 'relationship-tooltip')
                        .style('position', 'absolute')
                        .style('background', 'rgba(0,0,0,0.8)')
                        .style('color', 'white')
                        .style('padding', '8px')
                        .style('border-radius', '4px')
                        .style('font-size', '12px')
                        .style('pointer-events', 'none')
                        .style('z-index', '1000');

                    const props = Object.entries(d.properties)
                        .map(([key, value]) => `${key}: ${value}`)
                        .join('<br>');
                    
                    tooltip.html(`<strong>${d.type}</strong><br>${props}<br><em>Click to view details</em>`)
                        .style('left', (event.pageX + 10) + 'px')
                        .style('top', (event.pageY - 10) + 'px');
                } else {
                    const tooltip = d3.select('body').append('div')
                        .attr('class', 'relationship-tooltip')
                        .style('position', 'absolute')
                        .style('background', 'rgba(0,0,0,0.8)')
                        .style('color', 'white')
                        .style('padding', '8px')
                        .style('border-radius', '4px')
                        .style('font-size', '12px')
                        .style('pointer-events', 'none')
                        .style('z-index', '1000');
                    
                    tooltip.html(`<strong>${d.type}</strong><br><em>Click to view details</em>`)
                        .style('left', (event.pageX + 10) + 'px')
                        .style('top', (event.pageY - 10) + 'px');
                }
            })
            .on('mouseleave', function() {
                d3.select(this).select('rect')
                    .attr('opacity', 0.9)
                    .attr('stroke-width', 1);
                
                // Remove tooltip
                d3.selectAll('.relationship-tooltip').remove();
            });

        // Create the nodes
        const node = container.append('g')
            .attr('class', 'nodes')
            .selectAll('g')
            .data(nodesRef.current)
            .enter().append('g')
            .attr('class', 'node-group')
            .call(d3.drag<SVGGElement, D3Node>()
                .on('start', dragstarted)
                .on('drag', dragged)
                .on('end', dragended))
            .on('click', (event, d) => {
                event.stopPropagation();
                onNodeSelect?.(d.id);
            });

        // Add circles for nodes
        node.append('circle')
            .attr('r', d => d.type === 'Event' ? 12 : 10)
            .attr('fill', d => NODE_COLORS[d.type as keyof typeof NODE_COLORS] || '#9C27B0')
            .attr('stroke', '#fff')
            .attr('stroke-width', 2);

        // Add labels for nodes
        node.append('text')
            .text(d => d.label)
            .attr('x', 15)
            .attr('y', 4)
            .attr('font-size', '12px')
            .attr('font-weight', 'bold')
            .attr('fill', '#333')
            .style('pointer-events', 'none');

        // Highlight selected node
        if (selectedNode) {
            node.filter(d => d.id === selectedNode)
                .select('circle')
                .attr('stroke', '#000')
                .attr('stroke-width', 4);
        }

        // Update positions on each tick with proper type checking
        simulation.current.on('tick', () => {
            // Update link positions
            linkLines
                .attr('x1', function(d: any) {
                    return d.source?.x || 0;
                })
                .attr('y1', function(d: any) {
                    return d.source?.y || 0;
                })
                .attr('x2', function(d: any) {
                    return d.target?.x || 0;
                })
                .attr('y2', function(d: any) {
                    return d.target?.y || 0;
                });

            // Update label positions (centered on link)
            linkGroup.select('rect')
                .attr('transform', function(d: any) {
                    const sourceX = d.source?.x || 0;
                    const sourceY = d.source?.y || 0;
                    const targetX = d.target?.x || 0;
                    const targetY = d.target?.y || 0;
                    const midX = (sourceX + targetX) / 2;
                    const midY = (sourceY + targetY) / 2;
                    return `translate(${midX},${midY})`;
                });

            linkGroup.select('text')
                .attr('transform', function(d: any) {
                    const sourceX = d.source?.x || 0;
                    const sourceY = d.source?.y || 0;
                    const targetX = d.target?.x || 0;
                    const targetY = d.target?.y || 0;
                    const midX = (sourceX + targetX) / 2;
                    const midY = (sourceY + targetY) / 2;
                    return `translate(${midX},${midY})`;
                });

            // Update node positions
            node
                .attr('transform', function(d: any) {
                    const x = typeof d.x === 'number' ? d.x : 0;
                    const y = typeof d.y === 'number' ? d.y : 0;
                    return `translate(${x},${y})`;
                });
        });

        function dragstarted(event: d3.D3DragEvent<SVGGElement, D3Node, D3Node>, d: D3Node) {
            if (!event.active && simulation.current) simulation.current.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
        }

        function dragged(event: d3.D3DragEvent<SVGGElement, D3Node, D3Node>, d: D3Node) {
            d.fx = event.x;
            d.fy = event.y;
        }

        function dragended(event: d3.D3DragEvent<SVGGElement, D3Node, D3Node>, d: D3Node) {
            if (!event.active && simulation.current) simulation.current.alphaTarget(0);
            d.fx = null;
            d.fy = null;
        }
    };

    const [error, setError] = useState<string | null>(null);

    // Fetch network data on mount and when refresh is triggered
    useEffect(() => {
        const loadNetwork = async () => {
            try {
                const network = await api.getNetwork();
                // Validate network data before setting it
                if (Array.isArray(network)) {
                    const validNodes = network.filter(node => 
                        node && 
                        typeof node.id === 'string' && 
                        typeof node.type === 'string' && 
                        Array.isArray(node.connections)
                    );
                    setNetworkData(validNodes);
                    setError(null);
                } else {
                    console.error('Invalid network data format:', network);
                    setError('Invalid network data received');
                }
            } catch (error) {
                console.error('Failed to load network:', error);
                setError('Failed to load network data');
                setNetworkData([]);
            }
        };

        loadNetwork();
    }, [refreshTrigger]);

    // Update visualization whenever network data changes or when selection changes
    useEffect(() => {
        // Clear existing visualization data
        nodesRef.current = [];
        linksRef.current = [];

        // Process each node in the network if data exists
        if (networkData) {
            networkData.forEach(nodeData => {
            // Only add nodes that should be visible based on current selection
            const shouldShow = !selectedNode || nodeData.id === selectedNode || 
                             nodeData.connections.some(conn => conn.id === selectedNode);

            if (shouldShow && !nodesRef.current.find(n => n.id === nodeData.id)) {
                nodesRef.current.push({
                    id: nodeData.id,
                    label: (nodeData.properties && (nodeData.properties.name || nodeData.properties.title)) || `Unnamed ${nodeData.type}`,
                    type: nodeData.type,
                    properties: nodeData.properties || {},
                    x: Math.random() * 800,
                    y: Math.random() * 600
                });
            }

            // Process connections for each visible node
            if (shouldShow) {
                nodeData.connections.forEach(conn => {
                    const connShouldShow = !selectedNode || conn.id === selectedNode;
                    
                    if (connShouldShow && !nodesRef.current.find(n => n.id === conn.id)) {
                        nodesRef.current.push({
                            id: conn.id,
                            label: (conn.properties && (conn.properties.name || conn.properties.title)) || `Unnamed ${conn.type}`,
                            type: conn.type,
                            properties: conn.properties || {},
                            x: Math.random() * 800,
                            y: Math.random() * 600
                        });
                    }

                    if (connShouldShow) {
                        const sourceNode = nodesRef.current.find(n => n.id === (conn.relationship.direction === 'out' ? nodeData.id : conn.id));
                        const targetNode = nodesRef.current.find(n => n.id === (conn.relationship.direction === 'out' ? conn.id : nodeData.id));

                        if (sourceNode && targetNode) {
                            linksRef.current.push({
                                source: sourceNode,
                                target: targetNode,
                                type: conn.relationship.type,
                                properties: conn.relationship.properties
                            });
                        }
                    }
                });
            }
        });

            console.log('Visualization updated:', { 
                nodes: nodesRef.current.length, 
                links: linksRef.current.length,
                totalNodes: networkData.length
            });
        }

        // Only render if we have nodes to display
        if (nodesRef.current.length > 0) {
            renderGraph();
        }
    }, [networkData, selectedNode]);

    // Cleanup when component unmounts
    useEffect(() => {
        return () => {
            if (simulation.current) {
                simulation.current.stop();
            }
        };
    }, []);

    return (
        <div className="relative w-full h-full min-h-[600px]">
            {error && (
                <div className="absolute inset-0 flex items-center justify-center bg-white bg-opacity-90 z-10">
                    <div className="text-red-600 text-lg font-semibold">{error}</div>
                </div>
            )}
            <svg
                ref={svgRef}
                className="w-full h-full min-h-[600px] bg-white rounded-lg shadow-lg"
                style={{ cursor: 'grab', height: '80vh' }}
            />
        </div>
    );
};

export default D3Graph;