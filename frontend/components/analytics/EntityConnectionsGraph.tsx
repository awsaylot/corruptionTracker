import React, { useEffect, useRef } from 'react';
import * as d3 from 'd3';

interface EntityConnection {
    source: string | EntityNode;
    target: string | EntityNode;
    type: string;
    strength: number;
    properties?: Record<string, any>;
}

interface EntityNode extends d3.SimulationNodeDatum {
    id: string;
    name: string;
    type: string;
    risk_score?: number;
    properties?: Record<string, any>;
}

interface EntityConnectionsGraphProps {
    nodes: EntityNode[];
    connections: EntityConnection[];
    onNodeClick?: (nodeId: string) => void;
    onConnectionClick?: (connection: EntityConnection) => void;
    width?: number;
    height?: number;
}

const EntityConnectionsGraph: React.FC<EntityConnectionsGraphProps> = ({
    nodes,
    connections,
    onNodeClick,
    onConnectionClick,
    width = 800,
    height = 600
}) => {
    const svgRef = useRef<SVGSVGElement>(null);

    useEffect(() => {
        if (!svgRef.current || !nodes.length) return;

        // Clear previous graph
        d3.select(svgRef.current).selectAll("*").remove();

        // Create SVG container
        const svg = d3.select(svgRef.current)
            .attr("viewBox", [0, 0, width, height]);

        // Create force simulation
        const simulation = d3.forceSimulation<EntityNode>(nodes)
            .force("link", d3.forceLink<EntityNode, EntityConnection>(connections)
                .id(d => d.id)
                .distance(100))
            .force("charge", d3.forceManyBody<EntityNode>().strength(-200))
            .force("center", d3.forceCenter<EntityNode>(width / 2, height / 2))
            .force("collision", d3.forceCollide<EntityNode>().radius(50))
            .on("tick", () => {
                link
                    .attr("x1", d => (d.source as EntityNode).x!)
                    .attr("y1", d => (d.source as EntityNode).y!)
                    .attr("x2", d => (d.target as EntityNode).x!)
                    .attr("y2", d => (d.target as EntityNode).y!);

                node
                    .attr("transform", d => `translate(${d.x},${d.y})`);
            });

        // Create arrow marker
        svg.append("defs").selectAll("marker")
            .data(["arrow"])
            .join("marker")
            .attr("id", d => d)
            .attr("viewBox", "0 -5 10 10")
            .attr("refX", 15)
            .attr("refY", 0)
            .attr("markerWidth", 6)
            .attr("markerHeight", 6)
            .attr("orient", "auto")
            .append("path")
            .attr("fill", "#999")
            .attr("d", "M0,-5L10,0L0,5");

        // Draw links
        const link = svg.append("g")
            .selectAll("line")
            .data(connections)
            .join("line")
            .attr("stroke", "#999")
            .attr("stroke-opacity", 0.6)
            .attr("stroke-width", d => Math.sqrt(d.strength) * 2)
            .attr("marker-end", "url(#arrow)")
            .on("click", (event, d) => onConnectionClick?.(d));

        // Create node groups
        const node = svg.append("g")
            .selectAll("g")
            .data(nodes)
            .join("g")
            .call(d3.drag<SVGGElement, EntityNode>()
                .on("start", (event, d) => {
                    if (!event.active) simulation.alphaTarget(0.3).restart();
                    d.fx = d.x;
                    d.fy = d.y;
                })
                .on("drag", (event, d) => {
                    d.fx = event.x;
                    d.fy = event.y;
                })
                .on("end", (event) => {
                    if (!event.active) simulation.alphaTarget(0);
                    event.subject.fx = null;
                    event.subject.fy = null;
                }))
            .on("click", (event, d) => onNodeClick?.(d.id));

        // Add circles to nodes
        node.append("circle")
            .attr("r", 20)
            .attr("fill", d => getNodeColor(d.type, d.risk_score));

        // Add text labels to nodes
        node.append("text")
            .text(d => d.name)
            .attr("x", 25)
            .attr("y", 5)
            .attr("font-size", "12px");

        // Update positions on simulation tick
        simulation.on("tick", () => {
            link
                .attr("x1", d => (d.source as EntityNode).x ?? 0)
                .attr("y1", d => (d.source as EntityNode).y ?? 0)
                .attr("x2", d => (d.target as EntityNode).x ?? 0)
                .attr("y2", d => (d.target as EntityNode).y ?? 0);

            node
                .attr("transform", d => `translate(${d.x ?? 0},${d.y ?? 0})`);
        });

        // Drag functions
        function dragstarted(event: any) {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            event.subject.fx = event.subject.x;
            event.subject.fy = event.subject.y;
        }

        function dragged(event: any) {
            event.subject.fx = event.x;
            event.subject.fy = event.y;
        }

        function dragended(event: any) {
            if (!event.active) simulation.alphaTarget(0);
            event.subject.fx = null;
            event.subject.fy = null;
        }

        // Cleanup
        return () => {
            simulation.stop();
        };
    }, [nodes, connections, width, height, onNodeClick, onConnectionClick]);

    const getNodeColor = (type: string, risk_score?: number) => {
        if (risk_score !== undefined) {
            return d3.interpolateRdYlGn(1 - risk_score);
        }
        switch (type.toLowerCase()) {
            case 'person':
                return '#4CAF50';
            case 'organization':
                return '#2196F3';
            case 'event':
                return '#FFC107';
            default:
                return '#9E9E9E';
        }
    };

    return (
        <div className="relative w-full h-full">
            <svg
                ref={svgRef}
                width={width}
                height={height}
                className="w-full h-full"
            />
        </div>
    );
};

export default EntityConnectionsGraph;
