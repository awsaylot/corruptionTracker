import React, { useEffect, useRef, useCallback } from 'react';
import * as d3 from 'd3';

interface Node {
    id: string;
    label: string;
    type: string;
    properties: Record<string, any>;
}

interface Edge {
    source: string;
    target: string;
    type: string;
    properties: Record<string, any>;
}

interface GraphData {
    nodes: Node[];
    edges: Edge[];
}

interface GraphVisualizerProps {
    data: GraphData;
    width?: number;
    height?: number;
    onNodeClick?: (node: Node) => void;
    onEdgeClick?: (edge: Edge) => void;
}

const GraphVisualizer: React.FC<GraphVisualizerProps> = ({
    data,
    width = 800,
    height = 600,
    onNodeClick,
    onEdgeClick
}) => {
    const svgRef = useRef<SVGSVGElement>(null);
    const simulation = useRef<d3.Simulation<d3.SimulationNodeDatum, undefined>>();

    const initializeSimulation = useCallback(() => {
        if (!svgRef.current) return;

        // Clear existing SVG content
        d3.select(svgRef.current).selectAll("*").remove();

        // Create SVG container
        const svg = d3.select(svgRef.current)
            .attr("viewBox", `0 0 ${width} ${height}`)
            .attr("preserveAspectRatio", "xMidYMid meet");

        // Create arrow marker for directed edges
        svg.append("defs").selectAll("marker")
            .data(["arrow"])
            .join("marker")
            .attr("id", d => d)
            .attr("viewBox", "0 -5 10 10")
            .attr("refX", 25)
            .attr("refY", 0)
            .attr("markerWidth", 6)
            .attr("markerHeight", 6)
            .attr("orient", "auto")
            .append("path")
            .attr("fill", "#999")
            .attr("d", "M0,-5L10,0L0,5");

        // Create container group for zoom/pan
        const g = svg.append("g");

        // Add zoom behavior
        const zoom = d3.zoom<SVGSVGElement, unknown>()
            .scaleExtent([0.1, 4])
            .on("zoom", (event) => {
                g.attr("transform", event.transform);
            });

        svg.call(zoom);

        // Create edge elements
        const edges = g.append("g")
            .selectAll("line")
            .data(data.edges)
            .join("line")
            .attr("stroke", "#999")
            .attr("stroke-opacity", 0.6)
            .attr("stroke-width", 1)
            .attr("marker-end", "url(#arrow)")
            .on("click", (event, d) => {
                event.stopPropagation();
                onEdgeClick?.(d);
            });

        // Create node elements
        const nodes = g.append("g")
            .selectAll("g")
            .data(data.nodes)
            .join("g")
            .call(drag(simulation.current as any))
            .on("click", (event, d) => {
                event.stopPropagation();
                onNodeClick?.(d);
            });

        // Add circles for nodes
        nodes.append("circle")
            .attr("r", 10)
            .attr("fill", d => getNodeColor(d.type));

        // Add labels for nodes
        nodes.append("text")
            .text(d => d.label)
            .attr("x", 15)
            .attr("y", 5)
            .attr("font-size", "12px");

        // Create force simulation
        simulation.current = d3.forceSimulation(data.nodes as d3.SimulationNodeDatum[])
            .force("link", d3.forceLink(data.edges)
                .id((d: any) => d.id)
                .distance(100))
            .force("charge", d3.forceManyBody().strength(-200))
            .force("center", d3.forceCenter(width / 2, height / 2))
            .force("collision", d3.forceCollide().radius(30));

        // Update positions on each tick
        simulation.current.on("tick", () => {
            edges
                .attr("x1", (d: any) => d.source.x)
                .attr("y1", (d: any) => d.source.y)
                .attr("x2", (d: any) => d.target.x)
                .attr("y2", (d: any) => d.target.y);

            nodes
                .attr("transform", (d: any) => `translate(${d.x},${d.y})`);
        });

    }, [data, width, height, onNodeClick, onEdgeClick]);

    // Initialize when component mounts or data changes
    useEffect(() => {
        initializeSimulation();
        return () => {
            if (simulation.current) {
                simulation.current.stop();
            }
        };
    }, [data, initializeSimulation]);

    return (
        <svg
            ref={svgRef}
            width={width}
            height={height}
            className="bg-white rounded-lg shadow"
        />
    );
};

// Helper function to generate node colors based on type
const getNodeColor = (type: string): string => {
    const colors: Record<string, string> = {
        'Person': '#4299e1',   // blue
        'Organization': '#48bb78', // green
        'Event': '#ed8936',    // orange
        'Location': '#9f7aea', // purple
        'Document': '#f687b3'  // pink
    };
    return colors[type] || '#a0aec0'; // default gray
};

// Drag behavior helper function
const drag = (simulation: d3.Simulation<d3.SimulationNodeDatum, undefined>) => {
    const dragstarted = (event: any) => {
        if (!event.active) simulation.alphaTarget(0.3).restart();
        event.subject.fx = event.subject.x;
        event.subject.fy = event.subject.y;
    };

    const dragged = (event: any) => {
        event.subject.fx = event.x;
        event.subject.fy = event.y;
    };

    const dragended = (event: any) => {
        if (!event.active) simulation.alphaTarget(0);
        event.subject.fx = null;
        event.subject.fy = null;
    };

    return d3.drag<any, any>()
        .on("start", dragstarted)
        .on("drag", dragged)
        .on("end", dragended);
};

export default GraphVisualizer;
